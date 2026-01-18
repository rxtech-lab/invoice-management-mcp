# CLAUDE.md

Invoice Management System - A Go-based MCP server for managing invoices, categories, and companies.

## Architecture

- **MCP Server**: `github.com/mark3labs/mcp-go` for AI tool integration
- **REST API**: Fiber framework for HTTP endpoints
- **Database**: Turso (production) / SQLite in-memory (testing) with GORM ORM
- **Authentication**: MCPRouter via FiberApikeyMiddleware
- **File Storage**: S3-compatible storage (AWS S3, MinIO, etc.)

## Data Models

### InvoiceCategory
- `id` (uint) - Primary key
- `user_id` (string) - Index, required
- `name` (string) - Required
- `description` (text) - Optional
- `color` (varchar(7)) - Hex color

### InvoiceCompany
- `id` (uint) - Primary key
- `user_id` (string) - Index, required
- `name` (string) - Required
- `address`, `email`, `phone`, `website`, `tax_id`, `notes` - Optional fields

### Invoice
- `id` (uint) - Primary key
- `user_id` (string) - Index, required
- `title` (string) - Required
- `description` (text) - Optional
- `amount` (float64) - Default 0
- `currency` (varchar(3)) - Default 'USD'
- `category_id`, `company_id` - Foreign keys
- `status` (varchar(20)) - paid/unpaid/overdue
- `due_date` (timestamp) - Optional
- `invoice_started_at`, `invoice_ended_at` - Billing cycle dates
- `original_download_link` (text) - File URL

### InvoiceItem
- `id` (uint) - Primary key
- `invoice_id` (uint) - Foreign key, required
- `description` (string) - Required
- `quantity` (float64) - Default 1
- `unit_price` (float64) - Default 0
- `amount` (float64) - Computed: quantity * unit_price

## MCP Tools (21 total)

**Category**: `create_category`, `list_categories`, `get_category`, `update_category`, `delete_category`
**Company**: `create_company`, `list_companies`, `get_company`, `update_company`, `delete_company`
**Invoice**: `create_invoice`, `list_invoices`, `get_invoice`, `update_invoice`, `delete_invoice`, `search_invoices`, `update_invoice_status`
**Invoice Items**: `add_invoice_item`, `update_invoice_item`, `delete_invoice_item`
**Upload**: `upload_file`

## API Endpoints

### Categories
- `POST /api/categories` - Create category (201)
- `GET /api/categories` - List with search (`?keyword=`)
- `GET /api/categories/:id` - Get by ID
- `PUT /api/categories/:id` - Update
- `DELETE /api/categories/:id` - Delete (204)

### Companies
- `POST /api/companies` - Create company (201)
- `GET /api/companies` - List with search
- `GET /api/companies/:id` - Get by ID
- `PUT /api/companies/:id` - Update
- `DELETE /api/companies/:id` - Delete (204)

### Invoices
- `POST /api/invoices` - Create invoice (201)
- `GET /api/invoices` - List with filters, sort, search
- `GET /api/invoices/:id` - Get by ID (includes items)
- `PUT /api/invoices/:id` - Update
- `DELETE /api/invoices/:id` - Delete (204)
- `PATCH /api/invoices/:id/status` - Update status only

### Invoice Items
- `POST /api/invoices/:id/items` - Add item (201)
- `PUT /api/invoices/:invoice_id/items/:item_id` - Update item
- `DELETE /api/invoices/:invoice_id/items/:item_id` - Delete item (204)

### File Upload
- `POST /api/upload` - Upload file to S3 (201)
- `POST /api/upload/presigned` - Get presigned upload URL

### Health
- `GET /health` - Health check (no auth)

## Development Commands

```bash
# Build
make build       # Build the project
make run         # Run the server

# Testing
make test        # Run all tests
go test ./e2e/api/... -v -timeout 30s  # Run E2E tests

# Code Quality
make fmt         # Format code
make lint        # Run linter
make deps        # Download dependencies
```

## File Structure

```
invoice-management/
├── cmd/
│   └── server/main.go              # Server entry point
├── internal/
│   ├── api/
│   │   ├── server.go               # Fiber server setup
│   │   ├── category_handlers.go    # Category CRUD
│   │   ├── company_handlers.go     # Company CRUD
│   │   ├── invoice_handlers.go     # Invoice CRUD
│   │   ├── upload_handlers.go      # File upload
│   │   └── middleware/
│   │       └── auth.go             # Auth middleware
│   ├── mcp/
│   │   └── server.go               # MCP tools setup
│   ├── models/
│   │   ├── invoice.go
│   │   ├── invoice_category.go
│   │   ├── invoice_company.go
│   │   └── invoice_item.go
│   ├── services/
│   │   ├── db_service.go           # Turso/SQLite connection
│   │   ├── category_service.go
│   │   ├── company_service.go
│   │   ├── invoice_service.go
│   │   └── upload_service.go
│   ├── tools/                      # MCP tool implementations
│   │   ├── category_tools.go
│   │   ├── company_tools.go
│   │   ├── invoice_tools.go
│   │   ├── invoice_item_tools.go
│   │   └── upload_tools.go
│   └── utils/
│       ├── context.go              # Auth context helpers
│       └── jwt_authenticator.go    # JWT validation
├── e2e/
│   └── api/
│       ├── test_helpers.go
│       ├── category_test.go
│       ├── company_test.go
│       ├── invoice_test.go
│       └── upload_test.go
├── go.mod
├── Makefile
└── CLAUDE.md
```

## Environment Variables

```bash
# Database
TURSO_DATABASE_URL=libsql://your-db.turso.io
TURSO_AUTH_TOKEN=your-token

# Local SQLite (fallback)
SQLITE_DB_PATH=invoice.db

# S3-compatible storage
S3_ENDPOINT=https://s3.amazonaws.com
S3_BUCKET=invoice-uploads
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_REGION=us-east-1
S3_USE_PATH_STYLE=false

# Authentication
MCPROUTER_SERVER_URL=https://your-mcprouter.com
MCPROUTER_SERVER_API_KEY=your-api-key
JWT_SECRET=your-jwt-secret
SCALEKIT_ENV_URL=https://your-auth-provider.com/.well-known/jwks.json

# Server
PORT=8080
```

## Authentication

The system supports multiple authentication methods:

1. **MCPRouter**: API key-based authentication via `FiberApikeyMiddleware`
2. **JWT**: Simple JWT authentication with `JWT_SECRET`
3. **OAuth/JWKS**: OAuth 2.0 with JWKS validation via `SCALEKIT_ENV_URL`

Authentication is optional - if no authentication environment variables are set, the API runs without authentication.

### Authentication Flow

#### HTTP API Authentication
1. Client sends request with `Authorization: Bearer <jwt-token>` header
2. Auth middleware validates token using JwtAuthenticator
3. Authenticated user stored in Fiber context via `c.Locals(AuthenticatedUserContextKey, user)`
4. Handlers access user with: `user := c.Locals(AuthenticatedUserContextKey).(*utils.AuthenticatedUser)`

#### MCP Tool Authentication
1. HTTP request to `/mcp/*` endpoints includes `Authorization: Bearer <jwt-token>` header
2. Custom handler extracts and validates JWT token
3. Authenticated user added to Go context via `utils.WithAuthenticatedUser(ctx, user)`
4. MCP tools access user with: `user, ok := utils.GetAuthenticatedUser(ctx)`

## Testing

Tests use in-memory SQLite databases and mock services:

```go
// Create test setup
setup := NewTestSetup(t)
defer setup.Cleanup()

// Make authenticated requests
resp, err := setup.MakeRequest("POST", "/api/categories", payload)
```

Test authentication is handled via `X-Test-User-ID` header in tests.

### Running Tests

```bash
# Run all tests with 30s timeout
go test ./... -timeout 30s

# Run E2E API tests
go test ./e2e/api/... -v -timeout 30s

# Run specific test suite
go test ./e2e/api -run TestCategorySuite -v -timeout 30s
go test ./e2e/api -run TestInvoiceSuite -v -timeout 30s
```

## Tool Implementation Pattern

All MCP tools follow a consistent structure:

```go
type CreateCategoryTool struct {
    service services.CategoryService
}

func NewCreateCategoryTool(service services.CategoryService) *CreateCategoryTool {
    return &CreateCategoryTool{service: service}
}

func (t *CreateCategoryTool) GetTool() mcp.Tool {
    return mcp.NewTool("create_category",
        mcp.WithDescription("Create a new category"),
        mcp.WithString("name", mcp.Required(), mcp.Description("Category name")),
        mcp.WithString("description", mcp.Description("Category description")),
        mcp.WithString("color", mcp.Description("Hex color code")),
    )
}

func (t *CreateCategoryTool) GetHandler() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        userID := getUserIDFromContext(ctx)
        if userID == "" {
            return mcp.NewToolResultError("Authentication required"), nil
        }

        args := getArgsMap(request.Params.Arguments)
        // ... implementation
    }
}
```

## Code Guidelines

1. Never use `fmt.Println` for logging - use structured logging
2. Test timeout policy: Never run tests longer than 30 seconds
3. All handlers require user authentication via context
4. Services are user-scoped - all operations filter by `user_id`
5. Use service methods for data access, never raw GORM queries in handlers/tools

## Key Dependencies

- `github.com/mark3labs/mcp-go` - MCP server framework
- `github.com/gofiber/fiber/v2` - HTTP framework
- `gorm.io/gorm` and `gorm.io/driver/sqlite` - ORM and database
- `github.com/tursodatabase/libsql-client-go/libsql` - Turso client
- `github.com/aws/aws-sdk-go-v2` - S3-compatible storage
- `github.com/rxtech-lab/mcprouter-authenticator` - MCPRouter authentication
- `github.com/stretchr/testify` - Testing utilities
