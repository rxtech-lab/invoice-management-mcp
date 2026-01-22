package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/api"
	"github.com/rxtech-lab/invoice-management/internal/api/middleware"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
	"github.com/stretchr/testify/require"
)

// TestSetup contains all test dependencies
type TestSetup struct {
	t                *testing.T
	DBService        services.DBService
	CategoryService  services.CategoryService
	CompanyService   services.CompanyService
	ReceiverService  services.ReceiverService
	InvoiceService   services.InvoiceService
	UploadService    services.UploadService
	AnalyticsService services.AnalyticsService
	APIServer        *api.APIServer
	App              *fiber.App
	TestUserID       string
}

// NewTestSetup creates a new test setup with in-memory database
func NewTestSetup(t *testing.T) *TestSetup {
	// Create in-memory database
	dbService, err := services.NewSqliteDBService(":memory:")
	require.NoError(t, err, "Failed to create in-memory database")

	db := dbService.GetDB()

	// Create services
	categoryService := services.NewCategoryService(db)
	companyService := services.NewCompanyService(db)
	receiverService := services.NewReceiverService(db)
	tagService := services.NewTagService(db)
	invoiceService := services.NewInvoiceService(db)
	uploadService := services.NewMockUploadService()
	fileUploadService := services.NewFileUploadService(db)
	analyticsService := services.NewAnalyticsService(db)

	// Create file unlink service with empty URL (will skip unlinking)
	fileUnlinkService := services.NewFileUnlinkService(services.FileUnlinkConfig{
		FileServerURL: "",
	})

	// Create API server
	apiServer := api.NewAPIServer(
		dbService,
		categoryService,
		companyService,
		receiverService,
		tagService,
		invoiceService,
		uploadService,
		fileUploadService,
		analyticsService,
		fileUnlinkService,
		nil, // No PDF service for tests
		nil, // No MCP server for tests
	)

	// Add test authentication middleware before routes
	SetupTestAuthMiddleware(apiServer.GetFiberApp())

	// Setup routes
	apiServer.SetupRoutes()

	setup := &TestSetup{
		t:                t,
		DBService:        dbService,
		CategoryService:  categoryService,
		CompanyService:   companyService,
		ReceiverService:  receiverService,
		InvoiceService:   invoiceService,
		UploadService:    uploadService,
		AnalyticsService: analyticsService,
		APIServer:        apiServer,
		App:              apiServer.GetFiberApp(),
		TestUserID:       "test-user-123",
	}

	return setup
}

// Cleanup cleans up test resources
func (s *TestSetup) Cleanup() {
	if s.DBService != nil {
		s.DBService.Close()
	}
}

// MakeRequest makes an HTTP request to the test server
func (s *TestSetup) MakeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Add mock authentication header
	s.addAuthHeader(req)

	resp, err := s.App.Test(req, -1)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// MakeAuthenticatedRequest makes an authenticated HTTP request
func (s *TestSetup) MakeAuthenticatedRequest(method, path string, body interface{}, userID string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Add mock user context
	s.addAuthHeaderWithUserID(req, userID)

	resp, err := s.App.Test(req, -1)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// addAuthHeader adds authentication header for testing
func (s *TestSetup) addAuthHeader(req *http.Request) {
	s.addAuthHeaderWithUserID(req, s.TestUserID)
}

// addAuthHeaderWithUserID adds authentication header with a specific user ID
func (s *TestSetup) addAuthHeaderWithUserID(req *http.Request, userID string) {
	// For testing, we'll use a mock JWT or a test header
	// The actual authentication is handled by middleware
	req.Header.Set("X-Test-User-ID", userID)
}

// ReadResponseBody reads the response body as a map
func (s *TestSetup) ReadResponseBody(resp *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ReadResponseBodyArray reads the response body as an array
func (s *TestSetup) ReadResponseBodyArray(resp *http.Response) ([]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateTestCategory creates a test category
func (s *TestSetup) CreateTestCategory(name string) (uint, error) {
	category := &struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}{
		Name:        name,
		Description: "Test category",
		Color:       "#FF5733",
	}

	resp, err := s.MakeRequest("POST", "/api/categories", category)
	if err != nil {
		return 0, err
	}

	result, err := s.ReadResponseBody(resp)
	if err != nil {
		return 0, err
	}

	return uint(result["id"].(float64)), nil
}

// CreateTestCompany creates a test company
func (s *TestSetup) CreateTestCompany(name string) (uint, error) {
	company := &struct {
		Name    string `json:"name"`
		Address string `json:"address"`
		Email   string `json:"email"`
	}{
		Name:    name,
		Address: "123 Test St",
		Email:   "test@example.com",
	}

	resp, err := s.MakeRequest("POST", "/api/companies", company)
	if err != nil {
		return 0, err
	}

	result, err := s.ReadResponseBody(resp)
	if err != nil {
		return 0, err
	}

	return uint(result["id"].(float64)), nil
}

// CreateTestInvoice creates a test invoice
// Note: amount is not set - it's calculated from invoice items
func (s *TestSetup) CreateTestInvoice(title string, categoryID, companyID *uint) (uint, error) {
	invoice := map[string]interface{}{
		"title":       title,
		"description": "Test invoice",
		"currency":    "USD",
		"status":      "unpaid",
	}

	if categoryID != nil {
		invoice["category_id"] = *categoryID
	}
	if companyID != nil {
		invoice["company_id"] = *companyID
	}

	resp, err := s.MakeRequest("POST", "/api/invoices", invoice)
	if err != nil {
		return 0, err
	}

	result, err := s.ReadResponseBody(resp)
	if err != nil {
		return 0, err
	}

	return uint(result["id"].(float64)), nil
}

// CreateTestInvoiceItem creates a test invoice item
func (s *TestSetup) CreateTestInvoiceItem(invoiceID uint, description string, quantity, unitPrice float64) (uint, error) {
	item := map[string]interface{}{
		"description": description,
		"quantity":    quantity,
		"unit_price":  unitPrice,
	}

	resp, err := s.MakeRequest("POST", "/api/invoices/"+uintToStringHelper(invoiceID)+"/items", item)
	if err != nil {
		return 0, err
	}

	result, err := s.ReadResponseBody(resp)
	if err != nil {
		return 0, err
	}

	return uint(result["id"].(float64)), nil
}

// CreateTestReceiver creates a test receiver
func (s *TestSetup) CreateTestReceiver(name string, isOrganization bool) (uint, error) {
	receiver := &struct {
		Name           string `json:"name"`
		IsOrganization bool   `json:"is_organization"`
	}{
		Name:           name,
		IsOrganization: isOrganization,
	}

	resp, err := s.MakeRequest("POST", "/api/receivers", receiver)
	if err != nil {
		return 0, err
	}

	result, err := s.ReadResponseBody(resp)
	if err != nil {
		return 0, err
	}

	return uint(result["id"].(float64)), nil
}

// uintToStringHelper converts uint to string
func uintToStringHelper(n uint) string {
	return fmt.Sprintf("%d", n)
}

// SetupTestAuthMiddleware sets up a test authentication middleware
func SetupTestAuthMiddleware(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		userID := c.Get("X-Test-User-ID")
		if userID != "" {
			user := &utils.AuthenticatedUser{
				Sub: userID,
			}
			c.Locals(middleware.AuthenticatedUserContextKey, user)
		}
		return c.Next()
	})
}

// CreateTestInvoiceWithStatus creates a test invoice with specific status and amount
func (s *TestSetup) CreateTestInvoiceWithStatus(title string, categoryID, companyID *uint, status string, amount float64) (uint, error) {
	// Create invoice
	invoiceID, err := s.CreateTestInvoice(title, categoryID, companyID)
	if err != nil {
		return 0, err
	}

	// Add item to set the amount
	_, err = s.CreateTestInvoiceItem(invoiceID, "Item", 1, amount)
	if err != nil {
		return 0, err
	}

	// Update status
	update := map[string]interface{}{"status": status}
	_, err = s.MakeRequest("PATCH", "/api/invoices/"+uintToStringHelper(invoiceID)+"/status", update)
	if err != nil {
		return 0, err
	}

	return invoiceID, nil
}

// CreateTestInvoiceOnDate creates a test invoice with specific date, status, and amount
// Note: This updates the created_at field directly in the database for testing purposes
func (s *TestSetup) CreateTestInvoiceOnDate(title string, categoryID, companyID *uint, status string, amount float64, createdAt time.Time) (uint, error) {
	// Create invoice with status and amount
	invoiceID, err := s.CreateTestInvoiceWithStatus(title, categoryID, companyID, status, amount)
	if err != nil {
		return 0, err
	}

	// Update created_at directly in the database
	db := s.DBService.GetDB()
	if err := db.Exec("UPDATE invoices SET created_at = ? WHERE id = ?", createdAt, invoiceID).Error; err != nil {
		return 0, err
	}

	return invoiceID, nil
}

// DaysAgo returns a time that is n days ago from now
func DaysAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n)
}

// TestSetupWithFileServer extends TestSetup to support file server configuration and OAuth tokens
type TestSetupWithFileServer struct {
	*TestSetup
	fileServerURL string
	withOAuth     bool
}

// NewTestSetupWithFileServer creates a test setup with file server URL configuration
func NewTestSetupWithFileServer(t *testing.T, fileServerURL string, withOAuth bool) *TestSetupWithFileServer {
	// Create in-memory database
	dbService, err := services.NewSqliteDBService(":memory:")
	require.NoError(t, err, "Failed to create in-memory database")

	db := dbService.GetDB()

	// Create services
	categoryService := services.NewCategoryService(db)
	companyService := services.NewCompanyService(db)
	receiverService := services.NewReceiverService(db)
	tagService := services.NewTagService(db)
	invoiceService := services.NewInvoiceService(db)
	uploadService := services.NewMockUploadService()
	fileUploadService := services.NewFileUploadService(db)
	analyticsService := services.NewAnalyticsService(db)

	// Create file unlink service with the provided URL
	fileUnlinkService := services.NewFileUnlinkService(services.FileUnlinkConfig{
		FileServerURL: fileServerURL,
		Timeout:       5 * time.Second,
	})

	// Create API server with file unlink service
	apiServer := api.NewAPIServer(
		dbService,
		categoryService,
		companyService,
		receiverService,
		tagService,
		invoiceService,
		uploadService,
		fileUploadService,
		analyticsService,
		fileUnlinkService,
		nil, // No PDF service for tests
		nil, // No MCP server for tests
	)

	// Add test authentication middleware before routes
	if withOAuth {
		SetupTestAuthMiddlewareWithOAuth(apiServer.GetFiberApp())
	} else {
		SetupTestAuthMiddleware(apiServer.GetFiberApp())
	}

	// Setup routes
	apiServer.SetupRoutes()

	baseSetup := &TestSetup{
		t:                t,
		DBService:        dbService,
		CategoryService:  categoryService,
		CompanyService:   companyService,
		ReceiverService:  receiverService,
		InvoiceService:   invoiceService,
		UploadService:    uploadService,
		AnalyticsService: analyticsService,
		APIServer:        apiServer,
		App:              apiServer.GetFiberApp(),
		TestUserID:       "test-user-123",
	}

	return &TestSetupWithFileServer{
		TestSetup:     baseSetup,
		fileServerURL: fileServerURL,
		withOAuth:     withOAuth,
	}
}

// MakeRequestWithOAuth makes an HTTP request with OAuth Bearer token
func (s *TestSetupWithFileServer) MakeRequestWithOAuth(method, path string, body interface{}, token string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", s.TestUserID)

	// Add OAuth Bearer token
	if token != "" {
		req.Header.Set("X-Test-OAuth-Token", "Bearer "+token)
	}

	resp, err := s.App.Test(req, -1)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SetupTestAuthMiddlewareWithOAuth sets up test authentication middleware that supports OAuth tokens
func SetupTestAuthMiddlewareWithOAuth(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		userID := c.Get("X-Test-User-ID")
		if userID != "" {
			user := &utils.AuthenticatedUser{
				Sub: userID,
			}
			c.Locals(middleware.AuthenticatedUserContextKey, user)

			// Store OAuth token in locals if present
			oauthToken := c.Get("X-Test-OAuth-Token")
			if oauthToken != "" {
				c.Locals("Authorization", oauthToken)
			}
		}
		return c.Next()
	})
}
