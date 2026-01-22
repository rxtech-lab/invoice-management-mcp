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

// CreateReceiverTool handles receiver creation
type CreateReceiverTool struct {
	service services.ReceiverService
}

func NewCreateReceiverTool(service services.ReceiverService) *CreateReceiverTool {
	return &CreateReceiverTool{service: service}
}

func (t *CreateReceiverTool) GetTool() mcp.Tool {
	return mcp.NewTool("create_receiver",
		mcp.WithDescription("Create a new invoice receiver. If a receiver with the same name (or alias) already exists, returns the existing receiver instead of creating a duplicate."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Receiver name")),
		mcp.WithBoolean("is_organization", mcp.Description("Whether the receiver is an organization (default: false)")),
	)
}

func (t *CreateReceiverTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		name, _ := args["name"].(string)
		isOrganization, _ := args["is_organization"].(bool)

		// Check for existing receiver with same name or alias
		existingReceiver, err := t.service.FindByNameOrAlias(userID, name)
		if err == nil && existingReceiver != nil {
			// Found an existing receiver - return it instead of creating duplicate
			result, _ := json.Marshal(map[string]interface{}{
				"duplicate_found":    true,
				"existing_receiver":  existingReceiver,
				"message":            fmt.Sprintf("A receiver with name '%s' already exists (or is an alias of receiver '%s'). Use the existing receiver instead.", name, existingReceiver.Name),
			})
			return mcp.NewToolResultText(string(result)), nil
		}

		// No duplicate found, proceed with creation
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
}

// ListReceiversTool handles listing receivers
type ListReceiversTool struct {
	service services.ReceiverService
}

func NewListReceiversTool(service services.ReceiverService) *ListReceiversTool {
	return &ListReceiversTool{service: service}
}

func (t *ListReceiversTool) GetTool() mcp.Tool {
	return mcp.NewTool("list_receivers",
		mcp.WithDescription("List invoice receivers with optional search"),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 50)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
	)
}

func (t *ListReceiversTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// GetReceiverTool handles getting a single receiver
type GetReceiverTool struct {
	service services.ReceiverService
}

func NewGetReceiverTool(service services.ReceiverService) *GetReceiverTool {
	return &GetReceiverTool{service: service}
}

func (t *GetReceiverTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_receiver",
		mcp.WithDescription("Get a receiver by ID"),
		mcp.WithNumber("receiver_id", mcp.Required(), mcp.Description("Receiver ID")),
	)
}

func (t *GetReceiverTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// UpdateReceiverTool handles receiver updates
type UpdateReceiverTool struct {
	service services.ReceiverService
}

func NewUpdateReceiverTool(service services.ReceiverService) *UpdateReceiverTool {
	return &UpdateReceiverTool{service: service}
}

func (t *UpdateReceiverTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_receiver",
		mcp.WithDescription("Update an existing receiver. Can modify name, organization status, and alternative names (aliases)."),
		mcp.WithNumber("receiver_id", mcp.Required(), mcp.Description("Receiver ID")),
		mcp.WithString("name", mcp.Description("Receiver name")),
		mcp.WithBoolean("is_organization", mcp.Description("Whether the receiver is an organization")),
		mcp.WithArray("other_names", mcp.Description("Alternative names/aliases for the receiver (from merged receivers)"), mcp.Items(map[string]any{"type": "string"})),
	)
}

func (t *UpdateReceiverTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		receiverID := getUintArg(args, "receiver_id")
		if receiverID == 0 {
			return mcp.NewToolResultError("receiver_id is required"), nil
		}

		// Get existing receiver
		existing, err := t.service.GetReceiverByID(userID, receiverID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Receiver not found: %v", err)), nil
		}

		// Update fields if provided
		if name, ok := args["name"].(string); ok && name != "" {
			existing.Name = name
		}
		if isOrganization, ok := args["is_organization"].(bool); ok {
			existing.IsOrganization = isOrganization
		}

		// Handle other_names array
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
}

// DeleteReceiverTool handles receiver deletion
type DeleteReceiverTool struct {
	service services.ReceiverService
}

func NewDeleteReceiverTool(service services.ReceiverService) *DeleteReceiverTool {
	return &DeleteReceiverTool{service: service}
}

func (t *DeleteReceiverTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_receiver",
		mcp.WithDescription("Delete a receiver"),
		mcp.WithNumber("receiver_id", mcp.Required(), mcp.Description("Receiver ID")),
	)
}

func (t *DeleteReceiverTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		receiverID := getUintArg(args, "receiver_id")
		if receiverID == 0 {
			return mcp.NewToolResultError("receiver_id is required"), nil
		}

		if err := t.service.DeleteReceiver(userID, receiverID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete receiver: %v", err)), nil
		}

		return mcp.NewToolResultText(`{"success": true, "message": "Receiver deleted"}`), nil
	}
}
