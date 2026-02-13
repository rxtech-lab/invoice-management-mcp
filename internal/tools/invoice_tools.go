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

// ManageInvoiceTool is a unified CRUD tool for invoices
type ManageInvoiceTool struct {
	service services.InvoiceService
}

func NewManageInvoiceTool(service services.InvoiceService) *ManageInvoiceTool {
	return &ManageInvoiceTool{service: service}
}

func (t *ManageInvoiceTool) GetTool() mcp.Tool {
	return mcp.NewTool("manage_invoices",
		mcp.WithDescription("Unified tool for invoice management. Use the 'action' parameter to specify the operation: create, list, get, update, delete, search, or update_status."),
		mcp.WithString("action", mcp.Required(), mcp.Description("The operation to perform: create, list, get, update, delete, search, or update_status"), mcp.Enum("create", "list", "get", "update", "delete", "search", "update_status")),
		mcp.WithNumber("invoice_id", mcp.Description("Invoice ID (required for get, update, delete, update_status)")),
		mcp.WithString("title", mcp.Description("Invoice title (required for create, optional for update)")),
		mcp.WithString("description", mcp.Description("Invoice description (used with create, update)")),
		mcp.WithNumber("receiver_id", mcp.Description("Receiver ID (used with create, update, list)")),
		mcp.WithString("currency", mcp.Description("Currency code, default: USD (used with create, update)")),
		mcp.WithNumber("category_id", mcp.Description("Category ID (used with create, update, list)")),
		mcp.WithNumber("company_id", mcp.Description("Company ID (used with create, update, list)")),
		mcp.WithString("invoice_started_at", mcp.Description("Billing cycle start, RFC3339 (used with create)")),
		mcp.WithString("invoice_ended_at", mcp.Description("Billing cycle end, RFC3339 (used with create)")),
		mcp.WithString("original_download_link", mcp.Description("Link to original invoice file (used with create, update)")),
		mcp.WithString("status", mcp.Description("Status: paid, unpaid, overdue (used with create, update, update_status, list)")),
		mcp.WithString("due_date", mcp.Description("Due date, RFC3339 (used with create, update)")),
		mcp.WithArray("items", mcp.Description("Invoice items array. Each item should have: description (string, required), quantity (number, default 1), unit_price (number, required). Example: [{\"description\": \"Service\", \"quantity\": 1, \"unit_price\": 100}] (used with create)"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"description": map[string]any{"type": "string"},
					"quantity":    map[string]any{"type": "number"},
					"unit_price":  map[string]any{"type": "number"},
				},
				"required": []string{"description", "unit_price"},
			})),
		mcp.WithArray("tags", mcp.Description("Tag names to assign to the invoice (used with create, update). Pass empty array to remove all tags."), mcp.Items(map[string]any{"type": "string"})),
		mcp.WithString("keyword", mcp.Description("Search keyword (used with list)")),
		mcp.WithString("sort_by", mcp.Description("Sort by: created_at, amount, due_date, title (used with list)")),
		mcp.WithString("sort_order", mcp.Description("Sort order: asc, desc (used with list)")),
		mcp.WithNumber("limit", mcp.Description("Maximum results, default 50 (used with list)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (used with list)")),
		mcp.WithString("query", mcp.Description("Search query (required for search)")),
	)
}

func (t *ManageInvoiceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		action, _ := args["action"].(string)

		switch action {
		case "create":
			return t.handleCreate(userID, args)
		case "list":
			return t.handleList(userID, args)
		case "get":
			return t.handleGet(userID, args)
		case "update":
			return t.handleUpdate(userID, args)
		case "delete":
			return t.handleDelete(userID, args)
		case "search":
			return t.handleSearch(userID, args)
		case "update_status":
			return t.handleUpdateStatus(userID, args)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown action: %s. Must be one of: create, list, get, update, delete, search, update_status", action)), nil
		}
	}
}

func (t *ManageInvoiceTool) handleCreate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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

func (t *ManageInvoiceTool) handleList(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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

func (t *ManageInvoiceTool) handleGet(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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

func (t *ManageInvoiceTool) handleUpdate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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

func (t *ManageInvoiceTool) handleDelete(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	invoiceID := getUintArg(args, "invoice_id")
	if invoiceID == 0 {
		return mcp.NewToolResultError("invoice_id is required"), nil
	}

	if err := t.service.DeleteInvoice(userID, invoiceID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete invoice: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true, "message": "Invoice deleted"}`), nil
}

func (t *ManageInvoiceTool) handleSearch(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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

func (t *ManageInvoiceTool) handleUpdateStatus(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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
