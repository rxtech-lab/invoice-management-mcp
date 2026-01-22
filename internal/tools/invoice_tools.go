package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// CreateInvoiceTool handles invoice creation
type CreateInvoiceTool struct {
	service services.InvoiceService
}

func NewCreateInvoiceTool(service services.InvoiceService) *CreateInvoiceTool {
	return &CreateInvoiceTool{service: service}
}

func (t *CreateInvoiceTool) GetTool() mcp.Tool {
	return mcp.NewTool("create_invoice",
		mcp.WithDescription("Create a new invoice with items. Amount is calculated from invoice items. Duplicate detection: if an invoice with the same amount, billing dates, and receiver already exists, the existing invoice will be returned instead of creating a duplicate."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Invoice title")),
		mcp.WithString("description", mcp.Description("Invoice description")),
		mcp.WithNumber("receiver_id", mcp.Description("Receiver ID")),
		mcp.WithString("currency", mcp.Description("Currency code (default: USD)")),
		mcp.WithNumber("category_id", mcp.Description("Category ID")),
		mcp.WithNumber("company_id", mcp.Description("Company ID")),
		mcp.WithString("invoice_started_at", mcp.Description("Billing cycle start (RFC3339)")),
		mcp.WithString("invoice_ended_at", mcp.Description("Billing cycle end (RFC3339)")),
		mcp.WithString("original_download_link", mcp.Description("Link to original invoice file")),
		mcp.WithString("status", mcp.Description("Status: paid, unpaid, overdue (default: paid). Please justify the status base on the pdf file and the invoice items.")),
		mcp.WithString("due_date", mcp.Description("Due date (RFC3339)")),
		mcp.WithArray("items", mcp.Description("Invoice items array. Each item should have: description (string, required), quantity (number, default 1), unit_price (number, required). Example: [{\"description\": \"Service\", \"quantity\": 1, \"unit_price\": 100}]"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"description": map[string]any{"type": "string"},
					"quantity":    map[string]any{"type": "number"},
					"unit_price":  map[string]any{"type": "number"},
				},
				"required": []string{"description", "unit_price"},
			})),
		mcp.WithArray("tags", mcp.Description("Tag names to assign to the invoice (e.g., ['travel', 'business', 'Q1-2024']). Tags will be created if they don't exist."), mcp.Items(map[string]any{"type": "string"})),
	)
}

func (t *CreateInvoiceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		title, _ := args["title"].(string)
		description, _ := args["description"].(string)
		currency, _ := args["currency"].(string)
		if currency == "" {
			currency = "USD"
		}

		categoryID := getUintPtrArg(args, "category_id")
		companyID := getUintPtrArg(args, "company_id")
		receiverID := getUintPtrArg(args, "receiver_id")
		originalDownloadLink, _ := args["original_download_link"].(string)

		statusStr, _ := args["status"].(string)
		status := models.InvoiceStatusPaid
		if statusStr != "" {
			status = models.InvoiceStatus(statusStr)
		}

		invoiceStartedAt := parseTimeArg(args, "invoice_started_at")
		invoiceEndedAt := parseTimeArg(args, "invoice_ended_at")
		dueDate := parseTimeArg(args, "due_date")

		// Create invoice with items - amount is calculated from items
		invoice := &models.Invoice{
			Title:                title,
			Description:          description,
			ReceiverID:           receiverID,
			Currency:             currency,
			CategoryID:           categoryID,
			CompanyID:            companyID,
			InvoiceStartedAt:     invoiceStartedAt,
			InvoiceEndedAt:       invoiceEndedAt,
			OriginalDownloadLink: originalDownloadLink,
			Status:               status,
			DueDate:              dueDate,
		}

		// Parse and add items if provided
		if itemsRaw, ok := args["items"].([]interface{}); ok && len(itemsRaw) > 0 {
			for _, itemRaw := range itemsRaw {
				if itemMap, ok := itemRaw.(map[string]interface{}); ok {
					item := models.InvoiceItem{
						Description: getStringFromMap(itemMap, "description"),
						Quantity:    getFloatFromMap(itemMap, "quantity", 1),
						UnitPrice:   getFloatFromMap(itemMap, "unit_price", 0),
					}
					if item.Quantity == 0 {
						item.Quantity = 1
					}
					invoice.Items = append(invoice.Items, item)
				}
			}
		}

		createResult, err := t.service.CreateInvoice(userID, invoice)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create invoice: %v", err)), nil
		}

		// If duplicate found, return existing invoice with message
		if createResult.IsDuplicate {
			response := map[string]interface{}{
				"invoice":      createResult.Invoice,
				"is_duplicate": true,
				"message":      createResult.Message,
			}
			result, _ := json.Marshal(response)
			return mcp.NewToolResultText(string(result)), nil
		}

		// Set tags if provided (only for newly created invoices)
		if tagsRaw, ok := args["tags"].([]interface{}); ok && len(tagsRaw) > 0 {
			var tagNames []string
			for _, v := range tagsRaw {
				if name, ok := v.(string); ok && name != "" {
					tagNames = append(tagNames, name)
				}
			}
			if len(tagNames) > 0 {
				if err := t.service.SetInvoiceTags(userID, createResult.Invoice.ID, tagNames); err != nil {
					// Log error but don't fail the entire operation
					fmt.Printf("Warning: failed to set tags for invoice %d: %v\n", createResult.Invoice.ID, err)
				}
			}
		}

		created, _ := t.service.GetInvoiceByID(userID, createResult.Invoice.ID)
		result, _ := json.Marshal(created)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// ListInvoicesTool handles listing invoices
type ListInvoicesTool struct {
	service services.InvoiceService
}

func NewListInvoicesTool(service services.InvoiceService) *ListInvoicesTool {
	return &ListInvoicesTool{service: service}
}

func (t *ListInvoicesTool) GetTool() mcp.Tool {
	return mcp.NewTool("list_invoices",
		mcp.WithDescription("List invoices with filtering and sorting"),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
		mcp.WithNumber("category_id", mcp.Description("Filter by category ID")),
		mcp.WithNumber("company_id", mcp.Description("Filter by company ID")),
		mcp.WithNumber("receiver_id", mcp.Description("Filter by receiver ID")),
		mcp.WithString("status", mcp.Description("Filter by status: paid, unpaid, overdue")),
		mcp.WithString("sort_by", mcp.Description("Sort by: created_at, amount, due_date, title")),
		mcp.WithString("sort_order", mcp.Description("Sort order: asc, desc")),
		mcp.WithNumber("limit", mcp.Description("Maximum results (default 50)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
	)
}

func (t *ListInvoicesTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		opts := services.InvoiceListOptions{
			Keyword:   getStringArg(args, "keyword"),
			SortBy:    getStringArg(args, "sort_by"),
			SortOrder: getStringArg(args, "sort_order"),
			Limit:     getIntArg(args, "limit", 50),
			Offset:    getIntArg(args, "offset", 0),
		}

		if categoryID := getUintPtrArg(args, "category_id"); categoryID != nil {
			opts.CategoryID = categoryID
		}
		if companyID := getUintPtrArg(args, "company_id"); companyID != nil {
			opts.CompanyID = companyID
		}
		if receiverID := getUintPtrArg(args, "receiver_id"); receiverID != nil {
			opts.ReceiverID = receiverID
		}
		if statusStr := getStringArg(args, "status"); statusStr != "" {
			status := models.InvoiceStatus(statusStr)
			opts.Status = &status
		}

		invoices, total, err := t.service.ListInvoices(userID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list invoices: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]interface{}{
			"data":   invoices,
			"total":  total,
			"limit":  opts.Limit,
			"offset": opts.Offset,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

// GetInvoiceTool handles getting a single invoice
type GetInvoiceTool struct {
	service services.InvoiceService
}

func NewGetInvoiceTool(service services.InvoiceService) *GetInvoiceTool {
	return &GetInvoiceTool{service: service}
}

func (t *GetInvoiceTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_invoice",
		mcp.WithDescription("Get an invoice by ID with all details"),
		mcp.WithNumber("invoice_id", mcp.Required(), mcp.Description("Invoice ID")),
	)
}

func (t *GetInvoiceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		invoiceID := getUintArg(args, "invoice_id")
		if invoiceID == 0 {
			return mcp.NewToolResultError("invoice_id is required"), nil
		}

		invoice, err := t.service.GetInvoiceByID(userID, invoiceID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invoice not found: %v", err)), nil
		}

		result, _ := json.Marshal(invoice)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// UpdateInvoiceTool handles invoice updates
type UpdateInvoiceTool struct {
	service services.InvoiceService
}

func NewUpdateInvoiceTool(service services.InvoiceService) *UpdateInvoiceTool {
	return &UpdateInvoiceTool{service: service}
}

func (t *UpdateInvoiceTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_invoice",
		mcp.WithDescription("Update an existing invoice. Note: amount is calculated from invoice items and cannot be set directly. You can update tags to re-categorize the invoice."),
		mcp.WithNumber("invoice_id", mcp.Required(), mcp.Description("Invoice ID")),
		mcp.WithString("title", mcp.Description("Invoice title")),
		mcp.WithString("description", mcp.Description("Invoice description")),
		mcp.WithNumber("receiver_id", mcp.Description("Receiver ID")),
		mcp.WithString("currency", mcp.Description("Currency code")),
		mcp.WithNumber("category_id", mcp.Description("Category ID")),
		mcp.WithNumber("company_id", mcp.Description("Company ID")),
		mcp.WithString("original_download_link", mcp.Description("Link to original invoice file")),
		mcp.WithString("status", mcp.Description("Status: paid, unpaid, overdue")),
		mcp.WithString("due_date", mcp.Description("Due date (RFC3339)")),
		mcp.WithArray("tags", mcp.Description("Tag names to assign to the invoice (replaces existing tags). Pass empty array to remove all tags."), mcp.Items(map[string]any{"type": "string"})),
	)
}

func (t *UpdateInvoiceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		invoiceID := getUintArg(args, "invoice_id")
		if invoiceID == 0 {
			return mcp.NewToolResultError("invoice_id is required"), nil
		}

		title, _ := args["title"].(string)
		description, _ := args["description"].(string)
		currency, _ := args["currency"].(string)
		originalDownloadLink, _ := args["original_download_link"].(string)

		statusStr, _ := args["status"].(string)
		status := models.InvoiceStatus(statusStr)

		dueDate := parseTimeArg(args, "due_date")

		// Note: Amount is not set here - it's calculated from invoice items
		invoice := &models.Invoice{
			ID:                   invoiceID,
			Title:                title,
			Description:          description,
			ReceiverID:           getUintPtrArg(args, "receiver_id"),
			Currency:             currency,
			CategoryID:           getUintPtrArg(args, "category_id"),
			CompanyID:            getUintPtrArg(args, "company_id"),
			OriginalDownloadLink: originalDownloadLink,
			Status:               status,
			DueDate:              dueDate,
		}

		if err := t.service.UpdateInvoice(userID, invoice); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update invoice: %v", err)), nil
		}

		// Update tags if provided
		if tagsRaw, ok := args["tags"].([]interface{}); ok {
			var tagNames []string
			for _, v := range tagsRaw {
				if name, ok := v.(string); ok && name != "" {
					tagNames = append(tagNames, name)
				}
			}
			// SetInvoiceTags handles empty array (removes all tags)
			if err := t.service.SetInvoiceTags(userID, invoiceID, tagNames); err != nil {
				fmt.Printf("Warning: failed to set tags for invoice %d: %v\n", invoiceID, err)
			}
		}

		updated, _ := t.service.GetInvoiceByID(userID, invoiceID)
		result, _ := json.Marshal(updated)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// DeleteInvoiceTool handles invoice deletion
type DeleteInvoiceTool struct {
	service services.InvoiceService
}

func NewDeleteInvoiceTool(service services.InvoiceService) *DeleteInvoiceTool {
	return &DeleteInvoiceTool{service: service}
}

func (t *DeleteInvoiceTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_invoice",
		mcp.WithDescription("Delete an invoice"),
		mcp.WithNumber("invoice_id", mcp.Required(), mcp.Description("Invoice ID")),
	)
}

func (t *DeleteInvoiceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		invoiceID := getUintArg(args, "invoice_id")
		if invoiceID == 0 {
			return mcp.NewToolResultError("invoice_id is required"), nil
		}

		if err := t.service.DeleteInvoice(userID, invoiceID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete invoice: %v", err)), nil
		}

		return mcp.NewToolResultText(`{"success": true, "message": "Invoice deleted"}`), nil
	}
}

// SearchInvoicesTool handles invoice search
type SearchInvoicesTool struct {
	service services.InvoiceService
}

func NewSearchInvoicesTool(service services.InvoiceService) *SearchInvoicesTool {
	return &SearchInvoicesTool{service: service}
}

func (t *SearchInvoicesTool) GetTool() mcp.Tool {
	return mcp.NewTool("search_invoices",
		mcp.WithDescription("Full-text search across invoices"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
	)
}

func (t *SearchInvoicesTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		query, _ := args["query"].(string)
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		invoices, err := t.service.SearchInvoices(userID, query)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]interface{}{
			"data":  invoices,
			"count": len(invoices),
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

// UpdateInvoiceStatusTool handles status updates
type UpdateInvoiceStatusTool struct {
	service services.InvoiceService
}

func NewUpdateInvoiceStatusTool(service services.InvoiceService) *UpdateInvoiceStatusTool {
	return &UpdateInvoiceStatusTool{service: service}
}

func (t *UpdateInvoiceStatusTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_invoice_status",
		mcp.WithDescription("Update only the status of an invoice"),
		mcp.WithNumber("invoice_id", mcp.Required(), mcp.Description("Invoice ID")),
		mcp.WithString("status", mcp.Required(), mcp.Description("New status: paid, unpaid, overdue")),
	)
}

func (t *UpdateInvoiceStatusTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		invoiceID := getUintArg(args, "invoice_id")
		if invoiceID == 0 {
			return mcp.NewToolResultError("invoice_id is required"), nil
		}

		statusStr, _ := args["status"].(string)
		if statusStr == "" {
			return mcp.NewToolResultError("status is required"), nil
		}

		status := models.InvoiceStatus(statusStr)
		if err := t.service.UpdateInvoiceStatus(userID, invoiceID, status); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update status: %v", err)), nil
		}

		updated, _ := t.service.GetInvoiceByID(userID, invoiceID)
		result, _ := json.Marshal(updated)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// Helper functions
func getStringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func getFloatArg(args map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := args[key].(float64); ok {
		return v
	}
	return defaultVal
}

func getUintPtrArg(args map[string]interface{}, key string) *uint {
	if v, ok := args[key].(float64); ok && v > 0 {
		id := uint(v)
		return &id
	}
	return nil
}

func parseTimeArg(args map[string]interface{}, key string) *time.Time {
	if v, ok := args[key].(string); ok && v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			return &t
		}
	}
	return nil
}

// getStringFromMap extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// getFloatFromMap extracts a float64 value from a map with a default value
func getFloatFromMap(m map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return defaultVal
}
