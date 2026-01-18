package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rxtech-lab/invoice-management/internal/api"
	mcpserver "github.com/rxtech-lab/invoice-management/internal/mcp"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using environment variables")
	}

	// Initialize database
	dbService, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbService.Close()

	// Get the underlying GORM DB for service creation
	db := dbService.GetDB()

	// Initialize services
	categoryService := services.NewCategoryService(db)
	companyService := services.NewCompanyService(db)
	invoiceService := services.NewInvoiceService(db)
	uploadService := initUploadService()

	// Initialize MCP server
	mcpSrv := mcpserver.NewMCPServer(
		dbService,
		categoryService,
		companyService,
		invoiceService,
		uploadService,
	)

	// Initialize API server
	port := getEnvOrDefault("PORT", "8080")
	apiServer := api.NewAPIServer(
		dbService,
		categoryService,
		companyService,
		invoiceService,
		uploadService,
		mcpSrv.GetServer(),
	)

	// Setup routes
	apiServer.SetupRoutes()

	// Enable authentication if configured
	if os.Getenv("MCPROUTER_SERVER_URL") != "" {
		if err := apiServer.EnableAuthentication(); err != nil {
			log.Printf("Warning: Failed to enable authentication: %v", err)
		} else {
			log.Println("Authentication enabled via MCPRouter")
		}
	}

	// Enable StreamableHTTP for MCP
	apiServer.EnableStreamableHTTP()

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		cancel()
		if err := apiServer.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := apiServer.Start(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	<-ctx.Done()
	log.Println("Server stopped")
}

func initDatabase() (services.DBService, error) {
	tursoURL := os.Getenv("TURSO_DATABASE_URL")
	tursoToken := os.Getenv("TURSO_AUTH_TOKEN")

	if tursoURL != "" {
		log.Println("Connecting to Turso database...")
		return services.NewTursoDBService(tursoURL, tursoToken)
	}

	// Fall back to local SQLite
	dbPath := getEnvOrDefault("SQLITE_DB_PATH", "invoice.db")
	log.Printf("Using local SQLite database: %s", dbPath)
	return services.NewSqliteDBService(dbPath)
}

func initUploadService() services.UploadService {
	bucket := os.Getenv("S3_BUCKET")

	if bucket == "" {
		log.Println("Warning: S3_BUCKET not configured, file uploads will not work")
		return nil
	}

	cfg := services.S3Config{
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Bucket:          bucket,
		AccessKeyID:     os.Getenv("S3_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("S3_SECRET_KEY"),
		Region:          getEnvOrDefault("S3_REGION", "us-east-1"),
		UsePathStyle:    os.Getenv("S3_USE_PATH_STYLE") == "true",
	}

	service, err := services.NewUploadService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize S3 upload service: %v", err)
		return nil
	}

	log.Printf("S3 upload service initialized (bucket: %s)", bucket)
	return service
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
