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

// ManageTagTool is a unified tool for all tag operations
type ManageTagTool struct {
	service services.TagService
}

func NewManageTagTool(service services.TagService) *ManageTagTool {
	return &ManageTagTool{service: service}
}

func (t *ManageTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("manage_tags",
		mcp.WithDescription("Unified tool for managing invoice tags. Use the 'action' parameter to specify the operation: create, list, get, update, delete, add_to_invoice, remove_from_invoice, or search_invoices."),
		mcp.WithString("action", mcp.Required(), mcp.Description("The operation to perform"), mcp.Enum("create", "list", "get", "update", "delete", "add_to_invoice", "remove_from_invoice", "search_invoices")),
		mcp.WithNumber("tag_id", mcp.Description("Tag ID (required for get, update, delete, add_to_invoice, remove_from_invoice, search_invoices)")),
		mcp.WithString("name", mcp.Description("Tag name (required for create, optional for update)"), mcp.MaxLength(100)),
		mcp.WithString("color", mcp.Description("Hex color code (e.g., #FF5733). Please use different colors for different tags.")),
		mcp.WithString("keyword", mcp.Description("Search keyword (used with list)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results, default 50 (used with list, search_invoices)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (used with list, search_invoices)")),
		mcp.WithNumber("invoice_id", mcp.Description("Invoice ID (required for add_to_invoice, remove_from_invoice)")),
	)
}

func (t *ManageTagTool) GetHandler() server.ToolHandlerFunc {
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
		case "add_to_invoice":
			return t.handleAddToInvoice(userID, args)
		case "remove_from_invoice":
			return t.handleRemoveFromInvoice(userID, args)
		case "search_invoices":
			return t.handleSearchInvoices(userID, args)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown action: %s. Must be one of: create, list, get, update, delete, add_to_invoice, remove_from_invoice, search_invoices", action)), nil
		}
	}
}

func (t *ManageTagTool) handleCreate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	name, _ := args["name"].(string)
	color, _ := args["color"].(string)

	if color == "" {
		color = "#6B7280" // Default gray color
	}

	tag := &models.InvoiceTag{
		Name:  name,
		Color: color,
	}

	if err := t.service.CreateTag(userID, tag); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create tag: %v", err)), nil
	}

	result, _ := json.Marshal(tag)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageTagTool) handleList(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	keyword, _ := args["keyword"].(string)
	limit := getIntArg(args, "limit", 50)
	offset := getIntArg(args, "offset", 0)

	tags, total, err := t.service.ListTags(userID, keyword, limit, offset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list tags: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]interface{}{
		"data":   tags,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageTagTool) handleGet(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	tagID := getUintArg(args, "tag_id")
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required"), nil
	}

	tag, err := t.service.GetTagByID(userID, tagID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tag not found: %v", err)), nil
	}

	result, _ := json.Marshal(tag)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageTagTool) handleUpdate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	tagID := getUintArg(args, "tag_id")
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required"), nil
	}

	tag, err := t.service.GetTagByID(userID, tagID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tag not found: %v", err)), nil
	}

	if name, ok := args["name"].(string); ok && name != "" {
		tag.Name = name
	}
	if color, ok := args["color"].(string); ok && color != "" {
		tag.Color = color
	}

	if err := t.service.UpdateTag(userID, tag); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update tag: %v", err)), nil
	}

	result, _ := json.Marshal(tag)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageTagTool) handleDelete(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	tagID := getUintArg(args, "tag_id")
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required"), nil
	}

	if err := t.service.DeleteTag(userID, tagID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete tag: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true}`), nil
}

func (t *ManageTagTool) handleAddToInvoice(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	invoiceID := getUintArg(args, "invoice_id")
	tagID := getUintArg(args, "tag_id")

	if invoiceID == 0 {
		return mcp.NewToolResultError("invoice_id is required"), nil
	}
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required"), nil
	}

	if err := t.service.AddTagToInvoice(userID, invoiceID, tagID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add tag to invoice: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true}`), nil
}

func (t *ManageTagTool) handleRemoveFromInvoice(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	invoiceID := getUintArg(args, "invoice_id")
	tagID := getUintArg(args, "tag_id")

	if invoiceID == 0 {
		return mcp.NewToolResultError("invoice_id is required"), nil
	}
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required"), nil
	}

	if err := t.service.RemoveTagFromInvoice(userID, invoiceID, tagID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove tag from invoice: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true}`), nil
}

func (t *ManageTagTool) handleSearchInvoices(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	tagID := getUintArg(args, "tag_id")
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required"), nil
	}

	limit := getIntArg(args, "limit", 50)
	offset := getIntArg(args, "offset", 0)

	invoices, total, err := t.service.GetInvoicesByTagID(userID, tagID, limit, offset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search invoices: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]interface{}{
		"data":   invoices,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
	return mcp.NewToolResultText(string(result)), nil
}
