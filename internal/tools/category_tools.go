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

// CreateCategoryTool handles category creation
type CreateCategoryTool struct {
	service services.CategoryService
}

func NewCreateCategoryTool(service services.CategoryService) *CreateCategoryTool {
	return &CreateCategoryTool{service: service}
}

func (t *CreateCategoryTool) GetTool() mcp.Tool {
	return mcp.NewTool("create_category",
		mcp.WithDescription("Create a new invoice category"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Category name")),
		mcp.WithString("description", mcp.Description("Category description")),
		mcp.WithString("color", mcp.Description("Hex color code (e.g., #FF5733)")),
	)
}

func (t *CreateCategoryTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// ListCategoriesTool handles listing categories
type ListCategoriesTool struct {
	service services.CategoryService
}

func NewListCategoriesTool(service services.CategoryService) *ListCategoriesTool {
	return &ListCategoriesTool{service: service}
}

func (t *ListCategoriesTool) GetTool() mcp.Tool {
	return mcp.NewTool("list_categories",
		mcp.WithDescription("List invoice categories with optional search"),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 50)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
	)
}

func (t *ListCategoriesTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// GetCategoryTool handles getting a single category
type GetCategoryTool struct {
	service services.CategoryService
}

func NewGetCategoryTool(service services.CategoryService) *GetCategoryTool {
	return &GetCategoryTool{service: service}
}

func (t *GetCategoryTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_category",
		mcp.WithDescription("Get a category by ID"),
		mcp.WithNumber("category_id", mcp.Required(), mcp.Description("Category ID")),
	)
}

func (t *GetCategoryTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// UpdateCategoryTool handles category updates
type UpdateCategoryTool struct {
	service services.CategoryService
}

func NewUpdateCategoryTool(service services.CategoryService) *UpdateCategoryTool {
	return &UpdateCategoryTool{service: service}
}

func (t *UpdateCategoryTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_category",
		mcp.WithDescription("Update an existing category"),
		mcp.WithNumber("category_id", mcp.Required(), mcp.Description("Category ID")),
		mcp.WithString("name", mcp.Description("Category name")),
		mcp.WithString("description", mcp.Description("Category description")),
		mcp.WithString("color", mcp.Description("Hex color code")),
	)
}

func (t *UpdateCategoryTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// DeleteCategoryTool handles category deletion
type DeleteCategoryTool struct {
	service services.CategoryService
}

func NewDeleteCategoryTool(service services.CategoryService) *DeleteCategoryTool {
	return &DeleteCategoryTool{service: service}
}

func (t *DeleteCategoryTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_category",
		mcp.WithDescription("Delete a category"),
		mcp.WithNumber("category_id", mcp.Required(), mcp.Description("Category ID")),
	)
}

func (t *DeleteCategoryTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		categoryID := getUintArg(args, "category_id")
		if categoryID == 0 {
			return mcp.NewToolResultError("category_id is required"), nil
		}

		if err := t.service.DeleteCategory(userID, categoryID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete category: %v", err)), nil
		}

		return mcp.NewToolResultText(`{"success": true, "message": "Category deleted"}`), nil
	}
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
