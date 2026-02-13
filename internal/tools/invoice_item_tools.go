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

// ManageInvoiceItemTool is a unified tool for managing invoice items
type ManageInvoiceItemTool struct {
	service services.InvoiceService
}

func NewManageInvoiceItemTool(service services.InvoiceService) *ManageInvoiceItemTool {
	return &ManageInvoiceItemTool{service: service}
}

func (t *ManageInvoiceItemTool) GetTool() mcp.Tool {
	return mcp.NewTool("manage_invoice_items",
		mcp.WithDescription("Unified tool for managing invoice items. Use the 'action' parameter to specify the operation: add, update, or delete."),
		mcp.WithString("action", mcp.Required(), mcp.Description("The operation to perform: add, update, or delete"), mcp.Enum("add", "update", "delete")),
		mcp.WithNumber("invoice_id", mcp.Description("Invoice ID (required for add)")),
		mcp.WithNumber("item_id", mcp.Description("Item ID (required for update, delete)")),
		mcp.WithString("description", mcp.Description("Item description (required for add, optional for update)")),
		mcp.WithNumber("quantity", mcp.Description("Quantity (default 1 for add)")),
		mcp.WithNumber("unit_price", mcp.Description("Unit price")),
		mcp.WithNumber("target_amount", mcp.Description("Manual override for USD amount (optional, auto-calculated if not provided, used with update)")),
	)
}

func (t *ManageInvoiceItemTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		action, _ := args["action"].(string)

		switch action {
		case "add":
			return t.handleAdd(userID, args)
		case "update":
			return t.handleUpdate(userID, args)
		case "delete":
			return t.handleDelete(userID, args)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown action: %s. Must be one of: add, update, delete", action)), nil
		}
	}
}

func (t *ManageInvoiceItemTool) handleAdd(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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

func (t *ManageInvoiceItemTool) handleUpdate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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

func (t *ManageInvoiceItemTool) handleDelete(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	itemID := getUintArg(args, "item_id")
	if itemID == 0 {
		return mcp.NewToolResultError("item_id is required"), nil
	}

	if err := t.service.DeleteInvoiceItem(userID, itemID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete item: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true, "message": "Item deleted"}`), nil
}
