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

// ManageCompanyTool is a unified CRUD tool for invoice companies
type ManageCompanyTool struct {
	service services.CompanyService
}

func NewManageCompanyTool(service services.CompanyService) *ManageCompanyTool {
	return &ManageCompanyTool{service: service}
}

func (t *ManageCompanyTool) GetTool() mcp.Tool {
	return mcp.NewTool("manage_companies",
		mcp.WithDescription("Unified CRUD tool for invoice companies. Use the 'action' parameter to specify the operation: create, list, get, update, or delete."),
		mcp.WithString("action", mcp.Required(), mcp.Description("The operation to perform: create, list, get, update, or delete"), mcp.Enum("create", "list", "get", "update", "delete")),
		mcp.WithNumber("company_id", mcp.Description("Company ID (required for get, update, delete)")),
		mcp.WithString("name", mcp.Description("Company name (required for create, optional for update)")),
		mcp.WithString("address", mcp.Description("Company address (used with create, update)")),
		mcp.WithString("email", mcp.Description("Contact email (used with create, update)")),
		mcp.WithString("phone", mcp.Description("Phone number (used with create, update)")),
		mcp.WithString("website", mcp.Description("Website URL (used with create, update)")),
		mcp.WithString("tax_id", mcp.Description("Tax ID or VAT number (used with create, update)")),
		mcp.WithString("notes", mcp.Description("Additional notes (used with create, update)")),
		mcp.WithString("keyword", mcp.Description("Search keyword (used with list)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results, default 50 (used with list)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (used with list)")),
	)
}

func (t *ManageCompanyTool) GetHandler() server.ToolHandlerFunc {
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

func (t *ManageCompanyTool) handleCreate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	name, _ := args["name"].(string)
	address, _ := args["address"].(string)
	email, _ := args["email"].(string)
	phone, _ := args["phone"].(string)
	website, _ := args["website"].(string)
	taxID, _ := args["tax_id"].(string)
	notes, _ := args["notes"].(string)

	company := &models.InvoiceCompany{
		Name:    name,
		Address: address,
		Email:   email,
		Phone:   phone,
		Website: website,
		TaxID:   taxID,
		Notes:   notes,
	}

	if err := t.service.CreateCompany(userID, company); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create company: %v", err)), nil
	}

	result, _ := json.Marshal(company)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageCompanyTool) handleList(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	keyword, _ := args["keyword"].(string)
	limit := getIntArg(args, "limit", 50)
	offset := getIntArg(args, "offset", 0)

	companies, total, err := t.service.ListCompanies(userID, keyword, limit, offset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list companies: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]interface{}{
		"data":   companies,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageCompanyTool) handleGet(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	companyID := getUintArg(args, "company_id")
	if companyID == 0 {
		return mcp.NewToolResultError("company_id is required"), nil
	}

	company, err := t.service.GetCompanyByID(userID, companyID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Company not found: %v", err)), nil
	}

	result, _ := json.Marshal(company)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageCompanyTool) handleUpdate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	companyID := getUintArg(args, "company_id")
	if companyID == 0 {
		return mcp.NewToolResultError("company_id is required"), nil
	}

	name, _ := args["name"].(string)
	address, _ := args["address"].(string)
	email, _ := args["email"].(string)
	phone, _ := args["phone"].(string)
	website, _ := args["website"].(string)
	taxID, _ := args["tax_id"].(string)
	notes, _ := args["notes"].(string)

	company := &models.InvoiceCompany{
		ID:      companyID,
		Name:    name,
		Address: address,
		Email:   email,
		Phone:   phone,
		Website: website,
		TaxID:   taxID,
		Notes:   notes,
	}

	if err := t.service.UpdateCompany(userID, company); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update company: %v", err)), nil
	}

	updated, _ := t.service.GetCompanyByID(userID, companyID)
	result, _ := json.Marshal(updated)
	return mcp.NewToolResultText(string(result)), nil
}

func (t *ManageCompanyTool) handleDelete(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	companyID := getUintArg(args, "company_id")
	if companyID == 0 {
		return mcp.NewToolResultError("company_id is required"), nil
	}

	if err := t.service.DeleteCompany(userID, companyID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete company: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true, "message": "Company deleted"}`), nil
}
