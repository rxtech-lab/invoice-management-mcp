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

// ManageReceiverTool is a unified CRUD tool for invoice receivers
type ManageReceiverTool struct {
	service services.ReceiverService
}

func NewManageReceiverTool(service services.ReceiverService) *ManageReceiverTool {
	return &ManageReceiverTool{service: service}
}

func (t *ManageReceiverTool) GetTool() mcp.Tool {
	return mcp.NewTool("manage_receivers",
		mcp.WithDescription("Unified CRUD tool for invoice receivers. Use the 'action' parameter to specify the operation: create, list, get, update, delete, or merge."),
		mcp.WithString("action", mcp.Required(), mcp.Description("The operation to perform: create, list, get, update, delete, or merge"), mcp.Enum("create", "list", "get", "update", "delete", "merge")),
		mcp.WithNumber("receiver_id", mcp.Description("Receiver ID (required for get, update, delete)")),
		mcp.WithString("name", mcp.Description("Receiver name (required for create, optional for update)")),
		mcp.WithBoolean("is_organization", mcp.Description("Whether the receiver is an organization (default: false)")),
		mcp.WithArray("other_names", mcp.Description("Alternative names/aliases for the receiver (from merged receivers, used with update)"), mcp.Items(map[string]any{"type": "string"})),
		mcp.WithString("keyword", mcp.Description("Search keyword (used with list)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results, default 50 (used with list)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (used with list)")),
		mcp.WithNumber("target_id", mcp.Description("ID of the receiver to keep (required for merge)")),
		mcp.WithArray("source_ids", mcp.Description("IDs of receivers to merge into target (will be deleted, required for merge)"), mcp.Items(map[string]any{"type": "number"})),
	)
}

func (t *ManageReceiverTool) GetHandler() server.ToolHandlerFunc {
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
		case "merge":
			return t.handleMerge(userID, args)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown action: %s. Must be one of: create, list, get, update, delete, merge", action)), nil
		}
	}
}

func (t *ManageReceiverTool) handleCreate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	name, _ := args["name"].(string)
	isOrganization, _ := args["is_organization"].(bool)

	// Check for existing receiver with same name or alias
	existingReceiver, err := t.service.FindByNameOrAlias(userID, name)
	if err == nil && existingReceiver != nil {
		result, _ := json.Marshal(map[string]interface{}{
			"duplicate_found":   true,
			"existing_receiver": existingReceiver,
			"message":           fmt.Sprintf("A receiver with name '%s' already exists (or is an alias of receiver '%s'). Use the existing receiver instead.", name, existingReceiver.Name),
		})
		return mcp.NewToolResultText(string(result)), nil
	}

	receiver := &models.InvoiceReceiver{
		Name:           name,
		IsOrganization: isOrganization,
	}

	if err := t.service.CreateReceiver(userID, receiver); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create receiver: %v", err)), nil
	}

	result, _ := json.Marshal(receiver)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageReceiverTool) handleList(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	keyword, _ := args["keyword"].(string)
	limit := getIntArg(args, "limit", 50)
	offset := getIntArg(args, "offset", 0)

	receivers, total, err := t.service.ListReceivers(userID, keyword, limit, offset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list receivers: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]interface{}{
		"data":   receivers,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageReceiverTool) handleGet(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	receiverID := getUintArg(args, "receiver_id")
	if receiverID == 0 {
		return mcp.NewToolResultError("receiver_id is required"), nil
	}

	receiver, err := t.service.GetReceiverByID(userID, receiverID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Receiver not found: %v", err)), nil
	}

	result, _ := json.Marshal(receiver)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageReceiverTool) handleUpdate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	receiverID := getUintArg(args, "receiver_id")
	if receiverID == 0 {
		return mcp.NewToolResultError("receiver_id is required"), nil
	}

	existing, err := t.service.GetReceiverByID(userID, receiverID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Receiver not found: %v", err)), nil
	}

	if name, ok := args["name"].(string); ok && name != "" {
		existing.Name = name
	}
	if isOrganization, ok := args["is_organization"].(bool); ok {
		existing.IsOrganization = isOrganization
	}

	if otherNamesRaw, ok := args["other_names"].([]interface{}); ok {
		var otherNames []string
		for _, v := range otherNamesRaw {
			if s, ok := v.(string); ok {
				otherNames = append(otherNames, s)
			}
		}
		existing.OtherNames = otherNames
	}

	if err := t.service.UpdateReceiver(userID, existing); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update receiver: %v", err)), nil
	}

	updated, _ := t.service.GetReceiverByID(userID, receiverID)
	result, _ := json.Marshal(updated)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageReceiverTool) handleDelete(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	receiverID := getUintArg(args, "receiver_id")
	if receiverID == 0 {
		return mcp.NewToolResultError("receiver_id is required"), nil
	}

	if err := t.service.DeleteReceiver(userID, receiverID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete receiver: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true, "message": "Receiver deleted"}`), nil
}

func (t *ManageReceiverTool) handleMerge(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	targetID := getUintArg(args, "target_id")
	if targetID == 0 {
		return mcp.NewToolResultError("target_id is required"), nil
	}

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
