package api

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/api/middleware"
	"github.com/rxtech-lab/invoice-management/internal/assets"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
	auth "github.com/rxtech-lab/mcprouter-authenticator/authenticator"
	auth2 "github.com/rxtech-lab/mcprouter-authenticator/middleware"
	"github.com/rxtech-lab/mcprouter-authenticator/types"
)

type APIServer struct {
	app                    *fiber.App
	dbService              services.DBService
	categoryService        services.CategoryService
	companyService         services.CompanyService
	receiverService        services.ReceiverService
	invoiceService         services.InvoiceService
	uploadService          services.UploadService
	mcpServer              *mcpserver.MCPServer
	mcprouterAuthenticator *auth.ApikeyAuthenticator
	oauthAuthenticator     *middleware.OAuthAuthenticator
	port                   int
	authenticationEnabled  bool
}

// NewAPIServer creates a new API server instance
func NewAPIServer(
	dbService services.DBService,
	categoryService services.CategoryService,
	companyService services.CompanyService,
	receiverService services.ReceiverService,
	invoiceService services.InvoiceService,
	uploadService services.UploadService,
	mcpServer *mcpserver.MCPServer,
) *APIServer {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Add middleware
	app.Use(cors.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",
	}))

	// Initialize MCPRouter authenticator
	var mcprouterAuthenticator *auth.ApikeyAuthenticator
	if os.Getenv("MCPROUTER_SERVER_URL") != "" {
		mcprouterAuthenticator = auth.NewApikeyAuthenticator(os.Getenv("MCPROUTER_SERVER_URL"), http.DefaultClient)
		log.Println("MCPRouter authenticator initialized")
	}

	// Initialize OAuth authenticator
	var oauthAuthenticator *middleware.OAuthAuthenticator
	if oauthServerURL := os.Getenv("OAUTH_SERVER_URL"); oauthServerURL != "" {
		jwksEndpoint := oauthServerURL + "/.well-known/jwks.json"
		config := middleware.OAuthConfig{
			JWKSEndpoint: jwksEndpoint,
			Issuer:       os.Getenv("OAUTH_ISSUER"),
			Audience:     os.Getenv("OAUTH_AUDIENCE"),
		}

		var err error
		oauthAuthenticator, err = middleware.NewOAuthAuthenticator(config)
		if err != nil {
			log.Printf("Warning: Failed to initialize OAuth authenticator: %v", err)
		} else {
			log.Println("OAuth authenticator initialized")
		}
	}

	if mcprouterAuthenticator == nil && oauthAuthenticator == nil {
		log.Println("Warning: No authentication configured (MCPROUTER_SERVER_URL or OAUTH_SERVER_URL not set)")
	}

	srv := &APIServer{
		app:                    app,
		dbService:              dbService,
		categoryService:        categoryService,
		companyService:         companyService,
		receiverService:        receiverService,
		invoiceService:         invoiceService,
		uploadService:          uploadService,
		mcpServer:              mcpServer,
		mcprouterAuthenticator: mcprouterAuthenticator,
		oauthAuthenticator:     oauthAuthenticator,
	}
	return srv
}

// SetupRoutes configures all API routes
func (s *APIServer) SetupRoutes() {
	// Health check (no auth required)
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})

	// OpenAPI spec (no auth required)
	s.app.Get("/openapi", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/yaml")
		return c.Send(assets.OpenAPISpec)
	})

	// OpenAPI spec as JSON endpoint (no auth required)
	s.app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/yaml")
		return c.Send(assets.OpenAPISpec)
	})

	// Authentication check endpoint
	s.app.Get("/authentication", func(c *fiber.Ctx) error {
		user := c.Locals(middleware.AuthenticatedUserContextKey)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
		}
		authenticatedUser := user.(*utils.AuthenticatedUser)
		return c.JSON(fiber.Map{"status": "ok", "user": authenticatedUser})
	})

	// API routes group
	api := s.app.Group("/api")

	// Category routes
	api.Post("/categories", s.handleCreateCategory)
	api.Get("/categories", s.handleListCategories)
	api.Get("/categories/:id", s.handleGetCategory)
	api.Put("/categories/:id", s.handleUpdateCategory)
	api.Delete("/categories/:id", s.handleDeleteCategory)

	// Company routes
	api.Post("/companies", s.handleCreateCompany)
	api.Get("/companies", s.handleListCompanies)
	api.Get("/companies/:id", s.handleGetCompany)
	api.Put("/companies/:id", s.handleUpdateCompany)
	api.Delete("/companies/:id", s.handleDeleteCompany)

	// Receiver routes
	api.Post("/receivers", s.handleCreateReceiver)
	api.Get("/receivers", s.handleListReceivers)
	api.Get("/receivers/:id", s.handleGetReceiver)
	api.Put("/receivers/:id", s.handleUpdateReceiver)
	api.Delete("/receivers/:id", s.handleDeleteReceiver)

	// Invoice routes
	api.Post("/invoices", s.handleCreateInvoice)
	api.Get("/invoices", s.handleListInvoices)
	api.Get("/invoices/:id", s.handleGetInvoice)
	api.Put("/invoices/:id", s.handleUpdateInvoice)
	api.Delete("/invoices/:id", s.handleDeleteInvoice)
	api.Patch("/invoices/:id/status", s.handleUpdateInvoiceStatus)

	// Invoice item routes
	api.Post("/invoices/:id/items", s.handleAddInvoiceItem)
	api.Put("/invoices/:invoice_id/items/:item_id", s.handleUpdateInvoiceItem)
	api.Delete("/invoices/:invoice_id/items/:item_id", s.handleDeleteInvoiceItem)

	// Upload routes
	api.Post("/upload", s.handleUploadFile)
	api.Get("/upload/presigned", s.handleGetPresignedURL)
}

// EnableAuthentication enables authentication middleware (OAuth and/or MCPRouter)
func (s *APIServer) EnableAuthentication() error {
	s.authenticationEnabled = true

	// OAuth Bearer token authentication (applied first, takes priority)
	if s.oauthAuthenticator != nil {
		log.Println("OAuth Bearer token authentication enabled")
		s.app.Use(middleware.FiberOAuthMiddleware(s.oauthAuthenticator, func(c *fiber.Ctx, user *utils.AuthenticatedUser) error {
			c.Locals(middleware.AuthenticatedUserContextKey, user)
			return nil
		}))
	}

	// MCPRouter API key authentication (as fallback/alternative)
	if s.mcprouterAuthenticator != nil {
		log.Println("MCPRouter authentication enabled")
		s.app.Use(auth2.FiberApikeyMiddleware(s.mcprouterAuthenticator, os.Getenv("MCPROUTER_SERVER_API_KEY"), func(c *fiber.Ctx, user *types.User) error {
			// Only set user if not already authenticated (OAuth takes priority)
			if c.Locals(middleware.AuthenticatedUserContextKey) == nil {
				authenticatedUser := &utils.AuthenticatedUser{
					Sub:   user.ID,
					Roles: []string{user.Role},
				}
				c.Locals(middleware.AuthenticatedUserContextKey, authenticatedUser)
			}
			return nil
		}))
	}

	if s.oauthAuthenticator == nil && s.mcprouterAuthenticator == nil {
		log.Println("Warning: No authenticator configured, authentication skipped")
	}

	return nil
}

// EnableStreamableHTTP enables the MCP Streamable HTTP server
func (s *APIServer) EnableStreamableHTTP() {
	if s.mcpServer == nil {
		log.Println("Warning: MCP server not set. Skipping StreamableHTTP.")
		return
	}

	streamableServer := mcpserver.NewStreamableHTTPServer(s.mcpServer)

	var mcpHandler fiber.Handler
	if s.authenticationEnabled {
		mcpHandler = s.createAuthenticatedMCPHandler(streamableServer)
		log.Println("MCP handlers enabled with authentication")
	} else {
		mcpHandler = s.createUnauthenticatedMCPHandler(streamableServer)
		log.Println("MCP handlers enabled without authentication")
	}

	s.app.All("/mcp", mcpHandler)
	s.app.All("/mcp/*", mcpHandler)
}

// Start starts the server on the specified address
func (s *APIServer) Start(addr string) error {
	return s.app.Listen(addr)
}

// StartAsync starts the server asynchronously on a random available port or specified port
func (s *APIServer) StartAsync(port *int) (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("failed to find available port: %w", err)
	}

	assignedPort := listener.Addr().(*net.TCPAddr).Port
	s.port = assignedPort

	if port != nil {
		s.port = *port
	}

	err = listener.Close()
	if err != nil {
		return 0, err
	}

	go func() {
		if err := s.app.Listen(fmt.Sprintf(":%d", s.port)); err != nil {
			log.Printf("Error starting API server: %v\n", err)
		}
	}()

	return s.port, nil
}

// Shutdown gracefully shuts down the server
func (s *APIServer) Shutdown() error {
	if s.oauthAuthenticator != nil {
		s.oauthAuthenticator.Close()
	}
	return s.app.Shutdown()
}

// GetPort returns the server port
func (s *APIServer) GetPort() int {
	return s.port
}

// SetMCPServer sets the MCP server instance
func (s *APIServer) SetMCPServer(mcpServer *mcpserver.MCPServer) {
	s.mcpServer = mcpServer
}

// GetMCPServer returns the MCP server instance
func (s *APIServer) GetMCPServer() *mcpserver.MCPServer {
	return s.mcpServer
}

// GetFiberApp returns the underlying Fiber app (for testing)
func (s *APIServer) GetFiberApp() *fiber.App {
	return s.app
}

// createAuthenticatedMCPHandler creates a Fiber handler that enforces authentication
func (s *APIServer) createAuthenticatedMCPHandler(streamableServer *mcpserver.StreamableHTTPServer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals(middleware.AuthenticatedUserContextKey)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
		authenticatedUser := user.(*utils.AuthenticatedUser)

		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if authenticatedUser != nil {
				ctx = utils.WithAuthenticatedUser(ctx, authenticatedUser)
			}
			r = r.WithContext(ctx)
			streamableServer.ServeHTTP(w, r)
		})

		return adaptor.HTTPHandler(httpHandler)(c)
	}
}

// createUnauthenticatedMCPHandler creates a Fiber handler without authentication
func (s *APIServer) createUnauthenticatedMCPHandler(streamableServer *mcpserver.StreamableHTTPServer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			streamableServer.ServeHTTP(w, r)
		})
		return adaptor.HTTPHandler(httpHandler)(c)
	}
}

// getUserID extracts the user ID from the request context
func (s *APIServer) getUserID(c *fiber.Ctx) (string, error) {
	user := c.Locals(middleware.AuthenticatedUserContextKey)
	if user == nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Not authenticated")
	}
	authenticatedUser := user.(*utils.AuthenticatedUser)
	return authenticatedUser.Sub, nil
}
