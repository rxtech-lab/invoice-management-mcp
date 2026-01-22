package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// AddInvoiceItemTool handles adding items to invoices
type AddInvoiceItemTool struct {
	service services.InvoiceService
}

func NewAddInvoiceItemTool(service services.InvoiceService) *AddInvoiceItemTool {
	return &AddInvoiceItemTool{service: service}
}

func (t *AddInvoiceItemTool) GetTool() mcp.Tool {
	return mcp.NewTool("add_invoice_item",
		mcp.WithDescription("Add an item to an invoice"),
		mcp.WithNumber("invoice_id", mcp.Required(), mcp.Description("Invoice ID")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Item description")),
		mcp.WithNumber("quantity", mcp.Description("Quantity (default 1)")),
		mcp.WithNumber("unit_price", mcp.Description("Unit price")),
	)
}

func (t *AddInvoiceItemTool) GetHandler() server.ToolHandlerFunc {
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

		description, _ := args["description"].(string)
		if description == "" {
			return mcp.NewToolResultError("description is required"), nil
		}

		quantity := getFloatArg(args, "quantity", 1)
		unitPrice := getFloatArg(args, "unit_price", 0)

		item := &models.InvoiceItem{
			Description: description,
			Quantity:    quantity,
			UnitPrice:   unitPrice,
		}

		if err := t.service.AddInvoiceItem(userID, invoiceID, item); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add item: %v", err)), nil
		}

		result, _ := json.Marshal(item)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// UpdateInvoiceItemTool handles updating invoice items
type UpdateInvoiceItemTool struct {
	service services.InvoiceService
}

func NewUpdateInvoiceItemTool(service services.InvoiceService) *UpdateInvoiceItemTool {
	return &UpdateInvoiceItemTool{service: service}
}

func (t *UpdateInvoiceItemTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_invoice_item",
		mcp.WithDescription("Update an invoice item"),
		mcp.WithNumber("item_id", mcp.Required(), mcp.Description("Item ID")),
		mcp.WithString("description", mcp.Description("Item description")),
		mcp.WithNumber("quantity", mcp.Description("Quantity")),
		mcp.WithNumber("unit_price", mcp.Description("Unit price")),
		mcp.WithNumber("target_amount", mcp.Description("Manual override for USD amount (optional, auto-calculated if not provided)")),
	)
}

func (t *UpdateInvoiceItemTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		itemID := getUintArg(args, "item_id")
		if itemID == 0 {
			return mcp.NewToolResultError("item_id is required"), nil
		}

		description, _ := args["description"].(string)
		quantity := getFloatArg(args, "quantity", 0)
		unitPrice := getFloatArg(args, "unit_price", 0)

		// Handle optional target_amount override
		var targetAmountOverride *float64
		if targetAmount, ok := args["target_amount"].(float64); ok {
			targetAmountOverride = &targetAmount
		}

		item := &models.InvoiceItem{
			Description: description,
			Quantity:    quantity,
			UnitPrice:   unitPrice,
		}

		if err := t.service.UpdateInvoiceItem(userID, itemID, item, targetAmountOverride, false); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update item: %v", err)), nil
		}

		updated, _ := t.service.GetInvoiceItem(userID, itemID)
		result, _ := json.Marshal(updated)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// DeleteInvoiceItemTool handles deleting invoice items
type DeleteInvoiceItemTool struct {
	service services.InvoiceService
}

func NewDeleteInvoiceItemTool(service services.InvoiceService) *DeleteInvoiceItemTool {
	return &DeleteInvoiceItemTool{service: service}
}

func (t *DeleteInvoiceItemTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_invoice_item",
		mcp.WithDescription("Delete an invoice item"),
		mcp.WithNumber("item_id", mcp.Required(), mcp.Description("Item ID")),
	)
}

func (t *DeleteInvoiceItemTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		itemID := getUintArg(args, "item_id")
		if itemID == 0 {
			return mcp.NewToolResultError("item_id is required"), nil
		}

		if err := t.service.DeleteInvoiceItem(userID, itemID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete item: %v", err)), nil
		}

		return mcp.NewToolResultText(`{"success": true, "message": "Item deleted"}`), nil
	}
}
