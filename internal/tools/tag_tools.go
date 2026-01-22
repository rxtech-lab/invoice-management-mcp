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

// CreateTagTool handles tag creation
type CreateTagTool struct {
	service services.TagService
}

func NewCreateTagTool(service services.TagService) *CreateTagTool {
	return &CreateTagTool{service: service}
}

func (t *CreateTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("create_tag",
		mcp.WithDescription("Create a new invoice tag"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Tag name"), mcp.MaxLength(100)),
		mcp.WithString("color", mcp.Description("Hex color code (e.g., #FF5733). Please use different colors for different tags.")),
	)
}

func (t *CreateTagTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// ListTagsTool handles listing tags
type ListTagsTool struct {
	service services.TagService
}

func NewListTagsTool(service services.TagService) *ListTagsTool {
	return &ListTagsTool{service: service}
}

func (t *ListTagsTool) GetTool() mcp.Tool {
	return mcp.NewTool("list_tags",
		mcp.WithDescription("List invoice tags with optional search"),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 50)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
	)
}

func (t *ListTagsTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// GetTagTool handles getting a single tag
type GetTagTool struct {
	service services.TagService
}

func NewGetTagTool(service services.TagService) *GetTagTool {
	return &GetTagTool{service: service}
}

func (t *GetTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_tag",
		mcp.WithDescription("Get a tag by ID"),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID")),
	)
}

func (t *GetTagTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// UpdateTagTool handles updating a tag
type UpdateTagTool struct {
	service services.TagService
}

func NewUpdateTagTool(service services.TagService) *UpdateTagTool {
	return &UpdateTagTool{service: service}
}

func (t *UpdateTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_tag",
		mcp.WithDescription("Update an existing tag"),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID")),
		mcp.WithString("name", mcp.Description("New tag name")),
		mcp.WithString("color", mcp.Description("New hex color code")),
	)
}

func (t *UpdateTagTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		tagID := getUintArg(args, "tag_id")
		if tagID == 0 {
			return mcp.NewToolResultError("tag_id is required"), nil
		}

		// Get existing tag
		tag, err := t.service.GetTagByID(userID, tagID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tag not found: %v", err)), nil
		}

		// Update fields if provided
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
}

// DeleteTagTool handles tag deletion
type DeleteTagTool struct {
	service services.TagService
}

func NewDeleteTagTool(service services.TagService) *DeleteTagTool {
	return &DeleteTagTool{service: service}
}

func (t *DeleteTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_tag",
		mcp.WithDescription("Delete a tag"),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID")),
	)
}

func (t *DeleteTagTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		tagID := getUintArg(args, "tag_id")
		if tagID == 0 {
			return mcp.NewToolResultError("tag_id is required"), nil
		}

		if err := t.service.DeleteTag(userID, tagID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete tag: %v", err)), nil
		}

		return mcp.NewToolResultText(`{"success": true}`), nil
	}
}

// AddTagToInvoiceTool handles adding a tag to an invoice
type AddTagToInvoiceTool struct {
	service services.TagService
}

func NewAddTagToInvoiceTool(service services.TagService) *AddTagToInvoiceTool {
	return &AddTagToInvoiceTool{service: service}
}

func (t *AddTagToInvoiceTool) GetTool() mcp.Tool {
	return mcp.NewTool("add_tag_to_invoice",
		mcp.WithDescription("Add a tag to an invoice"),
		mcp.WithNumber("invoice_id", mcp.Required(), mcp.Description("Invoice ID")),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID")),
	)
}

func (t *AddTagToInvoiceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// RemoveTagFromInvoiceTool handles removing a tag from an invoice
type RemoveTagFromInvoiceTool struct {
	service services.TagService
}

func NewRemoveTagFromInvoiceTool(service services.TagService) *RemoveTagFromInvoiceTool {
	return &RemoveTagFromInvoiceTool{service: service}
}

func (t *RemoveTagFromInvoiceTool) GetTool() mcp.Tool {
	return mcp.NewTool("remove_tag_from_invoice",
		mcp.WithDescription("Remove a tag from an invoice"),
		mcp.WithNumber("invoice_id", mcp.Required(), mcp.Description("Invoice ID")),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID")),
	)
}

func (t *RemoveTagFromInvoiceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// SearchInvoicesByTagsTool handles searching invoices by tag IDs
type SearchInvoicesByTagsTool struct {
	service services.TagService
}

func NewSearchInvoicesByTagsTool(service services.TagService) *SearchInvoicesByTagsTool {
	return &SearchInvoicesByTagsTool{service: service}
}

func (t *SearchInvoicesByTagsTool) GetTool() mcp.Tool {
	return mcp.NewTool("search_invoices_by_tag",
		mcp.WithDescription("Find invoices that have a specific tag"),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID to search for")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 50)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
	)
}

func (t *SearchInvoicesByTagsTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// MergeReceiversTool handles merging multiple receivers into one
type MergeReceiversTool struct {
	service services.ReceiverService
}

func NewMergeReceiversTool(service services.ReceiverService) *MergeReceiversTool {
	return &MergeReceiversTool{service: service}
}

func (t *MergeReceiversTool) GetTool() mcp.Tool {
	return mcp.NewTool("merge_receivers",
		mcp.WithDescription("Merge multiple receivers into one. All invoices from source receivers will be moved to the target receiver."),
		mcp.WithNumber("target_id", mcp.Required(), mcp.Description("ID of the receiver to keep")),
		mcp.WithArray("source_ids", mcp.Required(), mcp.Description("IDs of receivers to merge into target (will be deleted)")),
	)
}

func (t *MergeReceiversTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		targetID := getUintArg(args, "target_id")
		if targetID == 0 {
			return mcp.NewToolResultError("target_id is required"), nil
		}

		// Get source IDs from array
		sourceIDsRaw, ok := args["source_ids"].([]interface{})
		if !ok || len(sourceIDsRaw) == 0 {
			return mcp.NewToolResultError("source_ids is required and must be a non-empty array"), nil
		}

		sourceIDs := make([]uint, 0, len(sourceIDsRaw))
		for _, v := range sourceIDsRaw {
			if id, ok := v.(float64); ok && id > 0 {
				sourceIDs = append(sourceIDs, uint(id))
			}
		}

		if len(sourceIDs) == 0 {
			return mcp.NewToolResultError("source_ids must contain valid IDs"), nil
		}

		receiver, affectedCount, err := t.service.MergeReceivers(userID, targetID, sourceIDs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to merge receivers: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]interface{}{
			"receiver":          receiver,
			"merged_count":      len(sourceIDs),
			"invoices_affected": affectedCount,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}
