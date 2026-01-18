package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/tools"
)

// MCPServer wraps the MCP server with invoice management tools
type MCPServer struct {
	server    *server.MCPServer
	dbService services.DBService
}

// NewMCPServer creates a new MCP server with invoice management tools
func NewMCPServer(
	dbService services.DBService,
	categoryService services.CategoryService,
	companyService services.CompanyService,
	invoiceService services.InvoiceService,
	uploadService services.UploadService,
) *MCPServer {
	mcpServer := &MCPServer{
		dbService: dbService,
	}
	mcpServer.initializeTools(categoryService, companyService, invoiceService, uploadService)
	return mcpServer
}

// initializeTools registers all invoice management tools
func (s *MCPServer) initializeTools(
	categoryService services.CategoryService,
	companyService services.CompanyService,
	invoiceService services.InvoiceService,
	uploadService services.UploadService,
) {
	srv := server.NewMCPServer(
		"Invoice Management MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	srv.EnableSampling()

	// Add usage prompt
	srv.AddPrompt(mcp.NewPrompt("invoice-management-usage",
		mcp.WithPromptDescription("Instructions and guidance for using invoice management tools"),
		mcp.WithArgument("tool_category",
			mcp.ArgumentDescription("Category of tools to get instructions for (category, company, invoice, upload, or all)"),
			mcp.RequiredArgument(),
		),
	), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		category := request.Params.Arguments["tool_category"]
		if category == "" {
			return nil, fmt.Errorf("tool_category is required")
		}

		instructions := getToolInstructions(category)

		return mcp.NewGetPromptResult(
			fmt.Sprintf("Invoice Management Tools - %s", category),
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(instructions),
				),
			},
		), nil
	})

	// Category Tools
	createCategoryTool := tools.NewCreateCategoryTool(categoryService)
	srv.AddTool(createCategoryTool.GetTool(), createCategoryTool.GetHandler())

	listCategoriesResult := tools.NewListCategoriesTool(categoryService)
	srv.AddTool(listCategoriesResult.GetTool(), listCategoriesResult.GetHandler())

	getCategoryTool := tools.NewGetCategoryTool(categoryService)
	srv.AddTool(getCategoryTool.GetTool(), getCategoryTool.GetHandler())

	updateCategoryTool := tools.NewUpdateCategoryTool(categoryService)
	srv.AddTool(updateCategoryTool.GetTool(), updateCategoryTool.GetHandler())

	deleteCategoryTool := tools.NewDeleteCategoryTool(categoryService)
	srv.AddTool(deleteCategoryTool.GetTool(), deleteCategoryTool.GetHandler())

	// Company Tools
	createCompanyTool := tools.NewCreateCompanyTool(companyService)
	srv.AddTool(createCompanyTool.GetTool(), createCompanyTool.GetHandler())

	listCompaniesTool := tools.NewListCompaniesTool(companyService)
	srv.AddTool(listCompaniesTool.GetTool(), listCompaniesTool.GetHandler())

	getCompanyTool := tools.NewGetCompanyTool(companyService)
	srv.AddTool(getCompanyTool.GetTool(), getCompanyTool.GetHandler())

	updateCompanyTool := tools.NewUpdateCompanyTool(companyService)
	srv.AddTool(updateCompanyTool.GetTool(), updateCompanyTool.GetHandler())

	deleteCompanyTool := tools.NewDeleteCompanyTool(companyService)
	srv.AddTool(deleteCompanyTool.GetTool(), deleteCompanyTool.GetHandler())

	// Invoice Tools
	createInvoiceTool := tools.NewCreateInvoiceTool(invoiceService)
	srv.AddTool(createInvoiceTool.GetTool(), createInvoiceTool.GetHandler())

	listInvoicesTool := tools.NewListInvoicesTool(invoiceService)
	srv.AddTool(listInvoicesTool.GetTool(), listInvoicesTool.GetHandler())

	getInvoiceTool := tools.NewGetInvoiceTool(invoiceService)
	srv.AddTool(getInvoiceTool.GetTool(), getInvoiceTool.GetHandler())

	updateInvoiceTool := tools.NewUpdateInvoiceTool(invoiceService)
	srv.AddTool(updateInvoiceTool.GetTool(), updateInvoiceTool.GetHandler())

	deleteInvoiceTool := tools.NewDeleteInvoiceTool(invoiceService)
	srv.AddTool(deleteInvoiceTool.GetTool(), deleteInvoiceTool.GetHandler())

	searchInvoicesTool := tools.NewSearchInvoicesTool(invoiceService)
	srv.AddTool(searchInvoicesTool.GetTool(), searchInvoicesTool.GetHandler())

	updateInvoiceStatusTool := tools.NewUpdateInvoiceStatusTool(invoiceService)
	srv.AddTool(updateInvoiceStatusTool.GetTool(), updateInvoiceStatusTool.GetHandler())

	// Invoice Item Tools
	addInvoiceItemTool := tools.NewAddInvoiceItemTool(invoiceService)
	srv.AddTool(addInvoiceItemTool.GetTool(), addInvoiceItemTool.GetHandler())

	updateInvoiceItemTool := tools.NewUpdateInvoiceItemTool(invoiceService)
	srv.AddTool(updateInvoiceItemTool.GetTool(), updateInvoiceItemTool.GetHandler())

	deleteInvoiceItemTool := tools.NewDeleteInvoiceItemTool(invoiceService)
	srv.AddTool(deleteInvoiceItemTool.GetTool(), deleteInvoiceItemTool.GetHandler())

	// Upload Tools
	getPresignedURLTool := tools.NewGetPresignedURLTool(uploadService)
	srv.AddTool(getPresignedURLTool.GetTool(), getPresignedURLTool.GetHandler())

	s.server = srv
}

// SendMessageToAiClient sends a message to the AI client
func (s *MCPServer) SendMessageToAiClient(messages []mcp.SamplingMessage) error {
	samplingRequest := mcp.CreateMessageRequest{
		CreateMessageParams: mcp.CreateMessageParams{
			Messages: messages,
		},
	}

	samplingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	serverFromCtx := server.ServerFromContext(samplingCtx)
	_, err := serverFromCtx.RequestSampling(samplingCtx, samplingRequest)
	if err != nil {
		return err
	}
	return nil
}

// getToolInstructions returns instructions for the specified tool category
func getToolInstructions(category string) string {
	switch category {
	case "category":
		return `Category Management Tools:

1. create_category - Create a new invoice category
   Parameters: name (required), description, color

2. list_categories - List all categories with optional search
   Parameters: keyword, limit, offset

3. get_category - Get a category by ID
   Parameters: category_id (required)

4. update_category - Update an existing category
   Parameters: category_id (required), name, description, color

5. delete_category - Delete a category
   Parameters: category_id (required)`

	case "company":
		return `Company Management Tools:

1. create_company - Create a new company
   Parameters: name (required), address, email, phone, website, tax_id, notes

2. list_companies - List all companies with optional search
   Parameters: keyword, limit, offset

3. get_company - Get a company by ID
   Parameters: company_id (required)

4. update_company - Update an existing company
   Parameters: company_id (required), name, address, email, phone, website, tax_id, notes

5. delete_company - Delete a company
   Parameters: company_id (required)`

	case "invoice":
		return `Invoice Management Tools:

1. create_invoice - Create a new invoice
   Parameters: title (required), description, amount, currency, category_id, company_id,
               invoice_started_at, invoice_ended_at, original_download_link, tags,
               status (paid/unpaid/overdue), due_date, items

2. list_invoices - List invoices with filtering and sorting
   Parameters: keyword, category_id, company_id, status, sort_by, sort_order, limit, offset

3. get_invoice - Get an invoice by ID with all details
   Parameters: invoice_id (required)

4. update_invoice - Update an existing invoice
   Parameters: invoice_id (required), and any fields to update

5. delete_invoice - Delete an invoice
   Parameters: invoice_id (required)

6. search_invoices - Full-text search across invoices
   Parameters: query (required)

7. update_invoice_status - Update only the status of an invoice
   Parameters: invoice_id (required), status (required: paid/unpaid/overdue)

Invoice Item Tools:
8. add_invoice_item - Add an item to an invoice
   Parameters: invoice_id (required), description (required), quantity, unit_price

9. update_invoice_item - Update an invoice item
   Parameters: item_id (required), description, quantity, unit_price

10. delete_invoice_item - Delete an invoice item
    Parameters: item_id (required)`

	case "upload":
		return `File Upload Tools:

1. get_presigned_url - Get a presigned URL for uploading a file
   Parameters: filename (required), content_type

   Usage: Use this to get a URL for directly uploading invoice attachments to S3.
   The returned URL can be used with PUT request to upload the file.
   After upload, use the returned key as the original_download_link in invoices.`

	case "all":
		return `Invoice Management MCP Tools Overview:

This MCP server provides tools for managing invoices, categories, companies, and file uploads.

CATEGORY MANAGEMENT (5 tools):
- create_category: Create a new category
- list_categories: List categories with search
- get_category: Get category details
- update_category: Update a category
- delete_category: Delete a category

COMPANY MANAGEMENT (5 tools):
- create_company: Create a new company
- list_companies: List companies with search
- get_company: Get company details
- update_company: Update a company
- delete_company: Delete a company

INVOICE MANAGEMENT (10 tools):
- create_invoice: Create a new invoice with items
- list_invoices: List with filters and sorting
- get_invoice: Get invoice with all details
- update_invoice: Update an invoice
- delete_invoice: Delete an invoice
- search_invoices: Full-text search
- update_invoice_status: Change invoice status
- add_invoice_item: Add item to invoice
- update_invoice_item: Update an item
- delete_invoice_item: Delete an item

FILE UPLOAD (1 tool):
- get_presigned_url: Get URL for file upload

All tools require authentication. Invoices are user-scoped.`

	default:
		return `Invalid category. Available categories: category, company, invoice, upload, all`
	}
}

// StartStdioServer starts the MCP server with stdio interface
func (s *MCPServer) StartStdioServer() error {
	return server.ServeStdio(s.server)
}

// StartStreamableHTTPServer starts the MCP server with streamable HTTP interface
func (s *MCPServer) StartStreamableHTTPServer() *server.StreamableHTTPServer {
	return server.NewStreamableHTTPServer(s.server)
}

// GetDBService returns the database service
func (s *MCPServer) GetDBService() services.DBService {
	return s.dbService
}

// GetServer returns the underlying MCP server
func (s *MCPServer) GetServer() *server.MCPServer {
	return s.server
}
