package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/api"
	"github.com/rxtech-lab/invoice-management/internal/api/middleware"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
	"github.com/stretchr/testify/require"
)

// TestSetup contains all test dependencies
type TestSetup struct {
	t               *testing.T
	DBService       services.DBService
	CategoryService services.CategoryService
	CompanyService  services.CompanyService
	ReceiverService services.ReceiverService
	InvoiceService  services.InvoiceService
	UploadService   services.UploadService
	APIServer       *api.APIServer
	App             *fiber.App
	TestUserID      string
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
	invoiceService := services.NewInvoiceService(db)
	uploadService := services.NewMockUploadService()

	// Create API server
	apiServer := api.NewAPIServer(
		dbService,
		categoryService,
		companyService,
		receiverService,
		invoiceService,
		uploadService,
		nil, // No MCP server for tests
	)

	// Add test authentication middleware before routes
	SetupTestAuthMiddleware(apiServer.GetFiberApp())

	// Setup routes
	apiServer.SetupRoutes()

	setup := &TestSetup{
		t:               t,
		DBService:       dbService,
		CategoryService: categoryService,
		CompanyService:  companyService,
		ReceiverService: receiverService,
		InvoiceService:  invoiceService,
		UploadService:   uploadService,
		APIServer:       apiServer,
		App:             apiServer.GetFiberApp(),
		TestUserID:      "test-user-123",
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
