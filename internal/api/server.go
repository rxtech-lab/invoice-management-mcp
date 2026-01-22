package api

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/api/handlers"
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
	tagService             services.TagService
	invoiceService         services.InvoiceService
	uploadService          services.UploadService
	fileUploadService      services.FileUploadService
	analyticsService       services.AnalyticsService
	fileUnlinkService      services.FileUnlinkService
	pdfService             services.PDFService
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
	tagService services.TagService,
	invoiceService services.InvoiceService,
	uploadService services.UploadService,
	fileUploadService services.FileUploadService,
	analyticsService services.AnalyticsService,
	fileUnlinkService services.FileUnlinkService,
	pdfService services.PDFService,
	mcpServer *mcpserver.MCPServer,
) *APIServer {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		// Custom error handler to properly handle errors from generated code
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Check if it's already a Fiber error
			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
			}
			// For other errors (like from generated code), return 400 for validation errors
			errMsg := err.Error()
			if strings.HasPrefix(errMsg, "Query argument") || strings.HasPrefix(errMsg, "Path argument") {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsg})
			}
			// Default to 500 for unexpected errors
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
		},
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
		tagService:             tagService,
		invoiceService:         invoiceService,
		uploadService:          uploadService,
		fileUploadService:      fileUploadService,
		analyticsService:       analyticsService,
		fileUnlinkService:      fileUnlinkService,
		pdfService:             pdfService,
		mcpServer:              mcpServer,
		mcprouterAuthenticator: mcprouterAuthenticator,
		oauthAuthenticator:     oauthAuthenticator,
	}
	return srv
}

// SetupRoutes configures all API routes
func (s *APIServer) SetupRoutes() {
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

	// Create strict handlers with all services
	strictHandlers := handlers.NewStrictHandlers(
		s.categoryService,
		s.companyService,
		s.receiverService,
		s.tagService,
		s.invoiceService,
		s.uploadService,
		s.fileUploadService,
		s.analyticsService,
		s.fileUnlinkService,
		s.pdfService,
	)

	// Create strict handler wrapper (converts StrictServerInterface to ServerInterface)
	strictHandler := generated.NewStrictHandler(strictHandlers, nil)

	// Register all API routes using generated handlers
	// Middleware checks authentication and passes user to Go context
	generated.RegisterHandlersWithOptions(s.app, strictHandler, generated.FiberServerOptions{
		BaseURL: "",
		Middlewares: []generated.MiddlewareFunc{
			func(c *fiber.Ctx) error {
				// Skip auth check for health endpoint
				if c.Path() == "/health" {
					return c.Next()
				}

				// Check if user is authenticated
				user := c.Locals(middleware.AuthenticatedUserContextKey)
				if user == nil {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"error": "Unauthorized",
					})
				}

				// Pass authenticated user to Go context for strict handlers
				ctx := utils.WithAuthenticatedUser(c.UserContext(), user.(*utils.AuthenticatedUser))

				// Pass Authorization header to Go context if available (for file unlinking)
				if authHeader := c.Locals("Authorization"); authHeader != nil {
					ctx = utils.WithAuthorizationHeader(ctx, authHeader.(string))
				}

				c.SetUserContext(ctx)
				return c.Next()
			},
		},
	})
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
