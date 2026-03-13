package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
)

// getArgsMap type-asserts the Arguments field to map[string]interface{}
func getArgsMap(args any) map[string]interface{} {
	if m, ok := args.(map[string]interface{}); ok {
		return m
	}
	return make(map[string]interface{})
}

// ManageCategoryTool is a unified CRUD tool for invoice categories
type ManageCategoryTool struct {
	service services.CategoryService
}

func NewManageCategoryTool(service services.CategoryService) *ManageCategoryTool {
	return &ManageCategoryTool{service: service}
}

func (t *ManageCategoryTool) GetTool() mcp.Tool {
	return mcp.NewTool("manage_categories",
		mcp.WithDescription("Unified CRUD tool for invoice categories. Use the 'action' parameter to specify the operation: create, list, get, update, or delete."),
		mcp.WithString("action", mcp.Required(), mcp.Description("The operation to perform: create, list, get, update, or delete"), mcp.Enum("create", "list", "get", "update", "delete")),
		mcp.WithNumber("category_id", mcp.Description("Category ID (required for get, update, delete)")),
		mcp.WithString("name", mcp.Description("Category name (required for create, optional for update)"), mcp.MaxLength(100)),
		mcp.WithString("description", mcp.Description("Category description (required for create, optional for update)")),
		mcp.WithString("color", mcp.Description("Hex color code (e.g., #FF5733). Please use different colors for different categories. (required for create, optional for update)")),
		mcp.WithString("keyword", mcp.Description("Search keyword (used with list)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results, default 50 (used with list)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (used with list)")),
	)
}

func (t *ManageCategoryTool) GetHandler() server.ToolHandlerFunc {
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
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown action: %s. Must be one of: create, list, get, update, delete", action)), nil
		}
	}
}

func (t *ManageCategoryTool) handleCreate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	color, _ := args["color"].(string)

	category := &models.InvoiceCategory{
		Name:        name,
		Description: description,
		Color:       color,
	}

	if err := t.service.CreateCategory(userID, category); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create category: %v", err)), nil
	}

	result, _ := json.Marshal(category)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageCategoryTool) handleList(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	keyword, _ := args["keyword"].(string)
	limit := getIntArg(args, "limit", 50)
	offset := getIntArg(args, "offset", 0)

	categories, total, err := t.service.ListCategories(userID, keyword, limit, offset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list categories: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]interface{}{
		"data":   categories,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageCategoryTool) handleGet(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	categoryID := getUintArg(args, "category_id")
	if categoryID == 0 {
		return mcp.NewToolResultError("category_id is required"), nil
	}

	category, err := t.service.GetCategoryByID(userID, categoryID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Category not found: %v", err)), nil
	}

	result, _ := json.Marshal(category)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageCategoryTool) handleUpdate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	categoryID := getUintArg(args, "category_id")
	if categoryID == 0 {
		return mcp.NewToolResultError("category_id is required"), nil
	}

	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	color, _ := args["color"].(string)

	category := &models.InvoiceCategory{
		ID:          categoryID,
		Name:        name,
		Description: description,
		Color:       color,
	}

	if err := t.service.UpdateCategory(userID, category); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update category: %v", err)), nil
	}

	updated, _ := t.service.GetCategoryByID(userID, categoryID)
	result, _ := json.Marshal(updated)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageCategoryTool) handleDelete(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	categoryID := getUintArg(args, "category_id")
	if categoryID == 0 {
		return mcp.NewToolResultError("category_id is required"), nil
	}

	if err := t.service.DeleteCategory(userID, categoryID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete category: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true, "message": "Category deleted"}`), nil
}

// Helper functions
func getUserIDFromContext(ctx context.Context) string {
	user, ok := utils.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		return ""
	}
	return user.Sub
}

func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}

func getUintArg(args map[string]interface{}, key string) uint {
	if v, ok := args[key].(float64); ok {
		return uint(v)
	}
	return 0
}
