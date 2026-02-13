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
	receiverService services.ReceiverService,
	invoiceService services.InvoiceService,
	uploadService services.UploadService,
	analyticsService services.AnalyticsService,
	tagService services.TagService,
) *MCPServer {
	mcpServer := &MCPServer{
		dbService: dbService,
	}
	mcpServer.initializeTools(categoryService, companyService, receiverService, invoiceService, uploadService, analyticsService, tagService)
	return mcpServer
}

// initializeTools registers all invoice management tools
func (s *MCPServer) initializeTools(
	categoryService services.CategoryService,
	companyService services.CompanyService,
	receiverService services.ReceiverService,
	invoiceService services.InvoiceService,
	uploadService services.UploadService,
	analyticsService services.AnalyticsService,
	tagService services.TagService,
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
	manageCategoryTool := tools.NewManageCategoryTool(categoryService)
	srv.AddTool(manageCategoryTool.GetTool(), manageCategoryTool.GetHandler())

	// Company Tools
	manageCompanyTool := tools.NewManageCompanyTool(companyService)
	srv.AddTool(manageCompanyTool.GetTool(), manageCompanyTool.GetHandler())

	// Receiver Tools
	manageReceiverTool := tools.NewManageReceiverTool(receiverService)
	srv.AddTool(manageReceiverTool.GetTool(), manageReceiverTool.GetHandler())

	// Invoice Tools
	manageInvoiceTool := tools.NewManageInvoiceTool(invoiceService)
	srv.AddTool(manageInvoiceTool.GetTool(), manageInvoiceTool.GetHandler())

	// Invoice Item Tools
	manageInvoiceItemTool := tools.NewManageInvoiceItemTool(invoiceService)
	srv.AddTool(manageInvoiceItemTool.GetTool(), manageInvoiceItemTool.GetHandler())

	// Upload Tools
	getPresignedURLTool := tools.NewGetPresignedURLTool(uploadService)
	srv.AddTool(getPresignedURLTool.GetTool(), getPresignedURLTool.GetHandler())

	// Statistics Tools
	invoiceStatisticsTool := tools.NewInvoiceStatisticsTool(analyticsService)
	srv.AddTool(invoiceStatisticsTool.GetTool(), invoiceStatisticsTool.GetHandler())

	advancedSearchTool := tools.NewAdvancedInvoiceSearchTool(analyticsService, invoiceService, tagService, categoryService, companyService, receiverService)
	srv.AddTool(advancedSearchTool.GetTool(), advancedSearchTool.GetHandler())

	// Tag Tools
	manageTagTool := tools.NewManageTagTool(tagService)
	srv.AddTool(manageTagTool.GetTool(), manageTagTool.GetHandler())

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
		return `Category Management - manage_categories tool:

A unified tool for managing invoice categories. Use the 'action' parameter to specify the operation.

Actions:
- create: Create a new invoice category
  Parameters: name (required), description (required), color (required)

- list: List all categories with optional search
  Parameters: keyword, limit, offset

- get: Get a category by ID
  Parameters: category_id (required)

- update: Update an existing category
  Parameters: category_id (required), name, description, color

- delete: Delete a category
  Parameters: category_id (required)`

	case "company":
		return `Company Management - manage_companies tool:

A unified tool for managing companies. Use the 'action' parameter to specify the operation.

Actions:
- create: Create a new company
  Parameters: name (required), address, email, phone, website, tax_id, notes

- list: List all companies with optional search
  Parameters: keyword, limit, offset

- get: Get a company by ID
  Parameters: company_id (required)

- update: Update an existing company
  Parameters: company_id (required), name, address, email, phone, website, tax_id, notes

- delete: Delete a company
  Parameters: company_id (required)`

	case "receiver":
		return `Receiver Management - manage_receivers tool:

A unified tool for managing invoice receivers. Use the 'action' parameter to specify the operation.

Actions:
- create: Create a new invoice receiver
  Parameters: name (required), is_organization (boolean, default false)

- list: List all receivers with optional search
  Parameters: keyword, limit, offset

- get: Get a receiver by ID
  Parameters: receiver_id (required)

- update: Update an existing receiver
  Parameters: receiver_id (required), name, is_organization, other_names

- delete: Delete a receiver
  Parameters: receiver_id (required)

- merge: Merge multiple receivers into one
  Parameters: target_id (required), source_ids (required array)
  All invoices from source receivers will be moved to the target receiver.`

	case "tag":
		return `Tag Management - manage_tags tool:

A unified tool for managing invoice tags. Use the 'action' parameter to specify the operation.

Actions:
- create: Create a new invoice tag
  Parameters: name (required), color (hex code, default #6B7280)

- list: List all tags with optional search
  Parameters: keyword, limit, offset

- get: Get a tag by ID
  Parameters: tag_id (required)

- update: Update an existing tag
  Parameters: tag_id (required), name, color

- delete: Delete a tag
  Parameters: tag_id (required)

- add_to_invoice: Add a tag to an invoice
  Parameters: invoice_id (required), tag_id (required)

- remove_from_invoice: Remove a tag from an invoice
  Parameters: invoice_id (required), tag_id (required)

- search_invoices: Find invoices with a specific tag
  Parameters: tag_id (required), limit, offset`

	case "invoice":
		return `Invoice Management - manage_invoices tool:

A unified tool for managing invoices. Use the 'action' parameter to specify the operation.

Actions:
- create: Create a new invoice
  Parameters: title (required), description, currency, category_id, company_id, receiver_id,
              invoice_started_at, invoice_ended_at, original_download_link, tags,
              status (paid/unpaid/overdue), due_date, items

- list: List invoices with filtering and sorting
  Parameters: keyword, category_id, company_id, receiver_id, status, sort_by, sort_order, limit, offset

- get: Get an invoice by ID with all details
  Parameters: invoice_id (required)

- update: Update an existing invoice
  Parameters: invoice_id (required), and any fields to update

- delete: Delete an invoice
  Parameters: invoice_id (required)

- search: Full-text search across invoices
  Parameters: query (required)

- update_status: Update only the status of an invoice
  Parameters: invoice_id (required), status (required: paid/unpaid/overdue)

Invoice Item Management - manage_invoice_items tool:

A unified tool for managing invoice items. Use the 'action' parameter to specify the operation.

Actions:
- add: Add an item to an invoice
  Parameters: invoice_id (required), description (required), quantity, unit_price

- update: Update an invoice item
  Parameters: item_id (required), description, quantity, unit_price, target_amount

- delete: Delete an invoice item
  Parameters: item_id (required)

Statistics Tools:
- invoice_statistics: Get invoice statistics with time period filtering, grouping, and aggregations
  Parameters: period (last_day/last_week/last_month/last_year), days, category_id, company_id,
              receiver_id, status, keyword, group_by (day/week/month/category/company/receiver),
              include_aggregations
- advanced_invoice_search: Search across title, category, company, receiver, tags`

	case "upload":
		return `File Upload Tools:

1. get_presigned_url - Get a presigned URL for uploading a file
   Parameters: filename (required), content_type

   Usage: Use this to get a URL for directly uploading invoice attachments to S3.
   The returned URL can be used with PUT request to upload the file.
   After upload, use the returned key as the original_download_link in invoices.`

	case "all":
		return `Invoice Management MCP Tools Overview:

This MCP server provides unified tools for managing invoices, categories, companies, receivers, tags, and file uploads.
Each resource has a single tool with an 'action' parameter for CRUD operations.

CATEGORY MANAGEMENT (1 tool: manage_categories):
  Actions: create, list, get, update, delete
  Parameters: action (required), category_id, name, description, color, keyword, limit, offset

COMPANY MANAGEMENT (1 tool: manage_companies):
  Actions: create, list, get, update, delete
  Parameters: action (required), company_id, name, address, email, phone, website, tax_id, notes, keyword, limit, offset

RECEIVER MANAGEMENT (1 tool: manage_receivers):
  Actions: create, list, get, update, delete, merge
  Parameters: action (required), receiver_id, name, is_organization, other_names, keyword, limit, offset, target_id, source_ids

TAG MANAGEMENT (1 tool: manage_tags):
  Actions: create, list, get, update, delete, add_to_invoice, remove_from_invoice, search_invoices
  Parameters: action (required), tag_id, name, color, keyword, limit, offset, invoice_id

INVOICE MANAGEMENT (1 tool: manage_invoices):
  Actions: create, list, get, update, delete, search, update_status
  Parameters: action (required), invoice_id, title, description, currency, category_id, company_id,
              receiver_id, status, due_date, items, tags, keyword, sort_by, sort_order, limit, offset, query

INVOICE ITEM MANAGEMENT (1 tool: manage_invoice_items):
  Actions: add, update, delete
  Parameters: action (required), invoice_id, item_id, description, quantity, unit_price, target_amount

FILE UPLOAD (1 tool: get_presigned_url):
  Parameters: filename (required), content_type

STATISTICS (2 tools):
- invoice_statistics: Get statistics with period/grouping/aggregations
- advanced_invoice_search: Search across title, category, company, receiver, tags

All tools require authentication. Invoices are user-scoped.`

	default:
		return `Invalid category. Available categories: category, company, receiver, invoice, upload, all`
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
