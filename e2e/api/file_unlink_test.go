package api

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type FileUnlinkTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *FileUnlinkTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *FileUnlinkTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

// TestDeleteInvoiceWithoutOAuthToken tests invoice deletion when no OAuth token is present
// Scenario: User authenticated via API key (not OAuth), so no Authorization header exists
func (s *FileUnlinkTestSuite) TestDeleteInvoiceWithoutOAuthToken() {
	// Create test invoice
	categoryID, _ := s.setup.CreateTestCategory("Test Category")
	companyID, _ := s.setup.CreateTestCompany("Test Company")
	invoiceID, err := s.setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)
	s.Require().NoError(err)

	// Setup mock file server that should NOT be called
	mockFileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Fail("File server should not be called when no OAuth token is present")
	}))
	defer mockFileServer.Close()

	// Create test setup with file server URL but without OAuth token
	setup := s.createTestSetupWithFileServer(mockFileServer.URL, false)
	defer setup.Cleanup()

	// Recreate the invoice in the new setup's database
	invoiceID, _ = setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)

	// Delete the invoice without OAuth token
	resp, err := setup.MakeRequest("DELETE", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify invoice is deleted
	resp, err = setup.MakeRequest("GET", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

// TestDeleteInvoiceWithOAuthTokenServerDown tests invoice deletion with retries when file server is down
// Scenario: OAuth token is present, but file server is unavailable
// Expected: Invoice still gets deleted successfully after retry attempts
func (s *FileUnlinkTestSuite) TestDeleteInvoiceWithOAuthTokenServerDown() {
	// Track number of retry attempts
	var attemptCount atomic.Int32

	// Setup mock file server that always returns 500 (server error)
	mockFileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount.Add(1)

		// Verify it's a DELETE request to /api/files/invoice
		s.Equal(http.MethodDelete, r.Method)
		s.Equal("/api/files/invoice", r.URL.Path)

		// Verify invoice_id query parameter exists
		invoiceIDStr := r.URL.Query().Get("invoice_id")
		s.NotEmpty(invoiceIDStr, "invoice_id query parameter should be present")

		// Verify Authorization header is present
		authHeader := r.Header.Get("Authorization")
		s.NotEmpty(authHeader, "Authorization header should be present")
		s.Contains(authHeader, "Bearer", "Authorization header should contain Bearer token")

		// Return 500 to simulate server error
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockFileServer.Close()

	// Create test setup with file server and OAuth token
	setup := s.createTestSetupWithFileServer(mockFileServer.URL, true)
	defer setup.Cleanup()

	// Create test invoice
	categoryID, _ := setup.CreateTestCategory("Test Category")
	companyID, _ := setup.CreateTestCompany("Test Company")
	invoiceID, _ := setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)

	// Delete the invoice with OAuth token but server down
	resp, err := setup.MakeRequestWithOAuth("DELETE", "/api/invoices/"+uintToString(invoiceID), nil, "test-token-123")
	s.Require().NoError(err)

	// Invoice should still be deleted successfully despite file server errors
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify the file server was called 3 times (initial + 2 retries)
	time.Sleep(100 * time.Millisecond) // Allow time for retries to complete
	s.Equal(int32(3), attemptCount.Load(), "File server should be called 3 times (initial attempt + 2 retries)")

	// Verify invoice is deleted
	resp, err = setup.MakeRequest("GET", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

// TestDeleteInvoiceWithOAuthTokenServerUp tests successful invoice and file unlinking
// Scenario: OAuth token is present and file server is working correctly
// Expected: File unlink API is called successfully, then invoice is deleted
func (s *FileUnlinkTestSuite) TestDeleteInvoiceWithOAuthTokenServerUp() {
	var receivedInvoiceID string
	var receivedAuthHeader string
	var unlinkCalled bool

	// Setup mock file server that returns 204 (success)
	mockFileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		unlinkCalled = true

		// Verify it's a DELETE request to /api/files/invoice
		s.Equal(http.MethodDelete, r.Method)
		s.Equal("/api/files/invoice", r.URL.Path)

		// Capture the invoice_id query parameter
		receivedInvoiceID = r.URL.Query().Get("invoice_id")
		s.NotEmpty(receivedInvoiceID, "invoice_id query parameter should be present")

		// Capture and verify Authorization header
		receivedAuthHeader = r.Header.Get("Authorization")
		s.NotEmpty(receivedAuthHeader, "Authorization header should be present")
		s.Contains(receivedAuthHeader, "Bearer test-token-456", "Authorization header should match")

		// Return 204 No Content (success)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockFileServer.Close()

	// Create test setup with file server and OAuth token
	setup := s.createTestSetupWithFileServer(mockFileServer.URL, true)
	defer setup.Cleanup()

	// Create test invoice
	categoryID, _ := setup.CreateTestCategory("Test Category")
	companyID, _ := setup.CreateTestCompany("Test Company")
	invoiceID, _ := setup.CreateTestInvoice("Test Invoice with File", &categoryID, &companyID)

	// Delete the invoice with OAuth token and working server
	resp, err := setup.MakeRequestWithOAuth("DELETE", "/api/invoices/"+uintToString(invoiceID), nil, "test-token-456")
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify the file unlink endpoint was called
	s.True(unlinkCalled, "File unlink API should have been called")

	// Verify the correct invoice ID was sent
	expectedInvoiceID := strconv.FormatUint(uint64(invoiceID), 10)
	s.Equal(expectedInvoiceID, receivedInvoiceID, "Correct invoice ID should be sent to file server")

	// Verify the Authorization header was passed correctly
	s.Equal("Bearer test-token-456", receivedAuthHeader, "Authorization header should be passed correctly")

	// Verify invoice is deleted
	resp, err = setup.MakeRequest("GET", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

// TestDeleteInvoiceWithOAuthTokenServerReturns404 tests when file server returns 404
// Scenario: File already doesn't exist on file server
// Expected: Invoice deletion proceeds successfully (404 is treated as success)
func (s *FileUnlinkTestSuite) TestDeleteInvoiceWithOAuthTokenServerReturns404() {
	var unlinkCalled bool

	// Setup mock file server that returns 404 (not found)
	mockFileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		unlinkCalled = true

		// Verify request
		s.Equal(http.MethodDelete, r.Method)
		s.Equal("/api/files/invoice", r.URL.Path)

		// Return 404 Not Found (file already unlinked)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockFileServer.Close()

	// Create test setup with file server and OAuth token
	setup := s.createTestSetupWithFileServer(mockFileServer.URL, true)
	defer setup.Cleanup()

	// Create test invoice
	categoryID, _ := setup.CreateTestCategory("Test Category")
	companyID, _ := setup.CreateTestCompany("Test Company")
	invoiceID, _ := setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)

	// Delete the invoice
	resp, err := setup.MakeRequestWithOAuth("DELETE", "/api/invoices/"+uintToString(invoiceID), nil, "test-token-789")
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify the file unlink endpoint was called
	s.True(unlinkCalled, "File unlink API should have been called")

	// Verify invoice is deleted
	resp, err = setup.MakeRequest("GET", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

// TestDeleteInvoiceWithoutFileServerURL tests invoice deletion when FILE_SERVER_URL is not configured
// Scenario: No file server URL is configured in environment
// Expected: Invoice deletion proceeds normally without attempting file unlink
func (s *FileUnlinkTestSuite) TestDeleteInvoiceWithoutFileServerURL() {
	// Create test setup without file server URL
	setup := s.createTestSetupWithFileServer("", true)
	defer setup.Cleanup()

	// Create test invoice
	categoryID, _ := setup.CreateTestCategory("Test Category")
	companyID, _ := setup.CreateTestCompany("Test Company")
	invoiceID, _ := setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)

	// Delete the invoice
	resp, err := setup.MakeRequestWithOAuth("DELETE", "/api/invoices/"+uintToString(invoiceID), nil, "test-token-unused")
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify invoice is deleted
	resp, err = setup.MakeRequest("GET", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

// Helper function to create test setup with file server configuration
func (s *FileUnlinkTestSuite) createTestSetupWithFileServer(fileServerURL string, withOAuth bool) *TestSetupWithFileServer {
	return NewTestSetupWithFileServer(s.T(), fileServerURL, withOAuth)
}

func TestFileUnlinkSuite(t *testing.T) {
	suite.Run(t, new(FileUnlinkTestSuite))
}
