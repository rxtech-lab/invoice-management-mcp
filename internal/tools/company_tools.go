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

// CreateCompanyTool handles company creation
type CreateCompanyTool struct {
	service services.CompanyService
}

func NewCreateCompanyTool(service services.CompanyService) *CreateCompanyTool {
	return &CreateCompanyTool{service: service}
}

func (t *CreateCompanyTool) GetTool() mcp.Tool {
	return mcp.NewTool("create_company",
		mcp.WithDescription("Create a new company"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Company name")),
		mcp.WithString("address", mcp.Description("Company address")),
		mcp.WithString("email", mcp.Description("Contact email")),
		mcp.WithString("phone", mcp.Description("Phone number")),
		mcp.WithString("website", mcp.Description("Website URL")),
		mcp.WithString("tax_id", mcp.Description("Tax ID or VAT number")),
		mcp.WithString("notes", mcp.Description("Additional notes")),
	)
}

func (t *CreateCompanyTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// ListCompaniesTool handles listing companies
type ListCompaniesTool struct {
	service services.CompanyService
}

func NewListCompaniesTool(service services.CompanyService) *ListCompaniesTool {
	return &ListCompaniesTool{service: service}
}

func (t *ListCompaniesTool) GetTool() mcp.Tool {
	return mcp.NewTool("list_companies",
		mcp.WithDescription("List companies with optional search"),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 50)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
	)
}

func (t *ListCompaniesTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// GetCompanyTool handles getting a single company
type GetCompanyTool struct {
	service services.CompanyService
}

func NewGetCompanyTool(service services.CompanyService) *GetCompanyTool {
	return &GetCompanyTool{service: service}
}

func (t *GetCompanyTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_company",
		mcp.WithDescription("Get a company by ID"),
		mcp.WithNumber("company_id", mcp.Required(), mcp.Description("Company ID")),
	)
}

func (t *GetCompanyTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// UpdateCompanyTool handles company updates
type UpdateCompanyTool struct {
	service services.CompanyService
}

func NewUpdateCompanyTool(service services.CompanyService) *UpdateCompanyTool {
	return &UpdateCompanyTool{service: service}
}

func (t *UpdateCompanyTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_company",
		mcp.WithDescription("Update an existing company"),
		mcp.WithNumber("company_id", mcp.Required(), mcp.Description("Company ID")),
		mcp.WithString("name", mcp.Description("Company name")),
		mcp.WithString("address", mcp.Description("Company address")),
		mcp.WithString("email", mcp.Description("Contact email")),
		mcp.WithString("phone", mcp.Description("Phone number")),
		mcp.WithString("website", mcp.Description("Website URL")),
		mcp.WithString("tax_id", mcp.Description("Tax ID or VAT number")),
		mcp.WithString("notes", mcp.Description("Additional notes")),
	)
}

func (t *UpdateCompanyTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
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
}

// DeleteCompanyTool handles company deletion
type DeleteCompanyTool struct {
	service services.CompanyService
}

func NewDeleteCompanyTool(service services.CompanyService) *DeleteCompanyTool {
	return &DeleteCompanyTool{service: service}
}

func (t *DeleteCompanyTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_company",
		mcp.WithDescription("Delete a company"),
		mcp.WithNumber("company_id", mcp.Required(), mcp.Description("Company ID")),
	)
}

func (t *DeleteCompanyTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		companyID := getUintArg(args, "company_id")
		if companyID == 0 {
			return mcp.NewToolResultError("company_id is required"), nil
		}

		if err := t.service.DeleteCompany(userID, companyID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete company: %v", err)), nil
		}

		return mcp.NewToolResultText(`{"success": true, "message": "Company deleted"}`), nil
	}
}
