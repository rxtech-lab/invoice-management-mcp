package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type InvoiceTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *InvoiceTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *InvoiceTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *InvoiceTestSuite) TestCreateInvoice() {
	// Create category and company first
	categoryID, err := s.setup.CreateTestCategory("Test Category")
	s.Require().NoError(err)
	companyID, err := s.setup.CreateTestCompany("Test Company")
	s.Require().NoError(err)

	// Note: amount is not sent - it's calculated from invoice items
	invoice := map[string]interface{}{
		"title":       "Monthly Services",
		"description": "Services for January 2024",
		"currency":    "USD",
		"category_id": categoryID,
		"company_id":  companyID,
		"status":      "unpaid",
		"due_date":    time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
	}

	resp, err := s.setup.MakeRequest("POST", "/api/invoices", invoice)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Monthly Services", result["title"])
	s.Equal("Services for January 2024", result["description"])
	s.Equal(float64(0), result["amount"]) // Amount is 0 because no items were added
	s.Equal("USD", result["currency"])
	s.Equal("unpaid", result["status"])
	s.NotNil(result["id"])
}

func (s *InvoiceTestSuite) TestCreateInvoiceMissingTitle() {
	invoice := map[string]interface{}{
		"currency": "USD",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/invoices", invoice)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *InvoiceTestSuite) TestListInvoices() {
	categoryID, _ := s.setup.CreateTestCategory("Test Category")
	companyID, _ := s.setup.CreateTestCompany("Test Company")

	_, err := s.setup.CreateTestInvoice("Invoice 1", &categoryID, &companyID)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoice("Invoice 2", &categoryID, &companyID)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/invoices", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.NotNil(result["data"])
	data := result["data"].([]interface{})
	s.GreaterOrEqual(len(data), 2)
}

func (s *InvoiceTestSuite) TestListInvoicesWithFilters() {
	categoryID, _ := s.setup.CreateTestCategory("Office")
	companyID, _ := s.setup.CreateTestCompany("Acme")

	_, err := s.setup.CreateTestInvoice("Office Supplies", &categoryID, &companyID)
	s.Require().NoError(err)

	// Filter by category
	resp, err := s.setup.MakeRequest("GET", "/api/invoices?category_id="+uintToString(categoryID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.GreaterOrEqual(len(data), 1)
}

func (s *InvoiceTestSuite) TestListInvoicesWithKeyword() {
	categoryID, _ := s.setup.CreateTestCategory("Test")
	companyID, _ := s.setup.CreateTestCompany("Test")

	_, err := s.setup.CreateTestInvoice("Office Supplies Invoice", &categoryID, &companyID)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoice("Travel Expenses", &categoryID, &companyID)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/invoices?keyword=office", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Equal(1, len(data))
}

func (s *InvoiceTestSuite) TestGetInvoice() {
	categoryID, _ := s.setup.CreateTestCategory("Test")
	companyID, _ := s.setup.CreateTestCompany("Test")

	invoiceID, err := s.setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Test Invoice", result["title"])
}

func (s *InvoiceTestSuite) TestGetInvoiceNotFound() {
	resp, err := s.setup.MakeRequest("GET", "/api/invoices/99999", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *InvoiceTestSuite) TestUpdateInvoice() {
	categoryID, _ := s.setup.CreateTestCategory("Test")
	companyID, _ := s.setup.CreateTestCompany("Test")

	invoiceID, err := s.setup.CreateTestInvoice("Original Invoice", &categoryID, &companyID)
	s.Require().NoError(err)

	// Note: amount is not sent - it's calculated from invoice items
	update := map[string]interface{}{
		"title":       "Updated Invoice",
		"description": "Updated description",
		"status":      "paid",
	}

	resp, err := s.setup.MakeRequest("PUT", "/api/invoices/"+uintToString(invoiceID), update)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Updated Invoice", result["title"])
	s.Equal("Updated description", result["description"])
	// Amount is not updated directly - it's calculated from items
	s.Equal("paid", result["status"])
}

func (s *InvoiceTestSuite) TestUpdateInvoiceStatus() {
	categoryID, _ := s.setup.CreateTestCategory("Test")
	companyID, _ := s.setup.CreateTestCompany("Test")

	invoiceID, err := s.setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)
	s.Require().NoError(err)

	update := map[string]interface{}{
		"status": "paid",
	}

	resp, err := s.setup.MakeRequest("PATCH", "/api/invoices/"+uintToString(invoiceID)+"/status", update)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("paid", result["status"])
}

func (s *InvoiceTestSuite) TestDeleteInvoice() {
	categoryID, _ := s.setup.CreateTestCategory("Test")
	companyID, _ := s.setup.CreateTestCompany("Test")

	invoiceID, err := s.setup.CreateTestInvoice("To Delete", &categoryID, &companyID)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("DELETE", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify deletion
	resp, err = s.setup.MakeRequest("GET", "/api/invoices/"+uintToString(invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *InvoiceTestSuite) TestAddInvoiceItem() {
	categoryID, _ := s.setup.CreateTestCategory("Test")
	companyID, _ := s.setup.CreateTestCompany("Test")

	invoiceID, err := s.setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)
	s.Require().NoError(err)

	item := map[string]interface{}{
		"description": "Consulting Services",
		"quantity":    10,
		"unit_price":  150.00,
	}

	resp, err := s.setup.MakeRequest("POST", "/api/invoices/"+uintToString(invoiceID)+"/items", item)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Consulting Services", result["description"])
	s.Equal(float64(10), result["quantity"])
	s.Equal(float64(150), result["unit_price"])
	s.NotNil(result["id"])
}

func (s *InvoiceTestSuite) TestUpdateInvoiceItem() {
	categoryID, _ := s.setup.CreateTestCategory("Test")
	companyID, _ := s.setup.CreateTestCompany("Test")

	invoiceID, err := s.setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)
	s.Require().NoError(err)

	itemID, err := s.setup.CreateTestInvoiceItem(invoiceID, "Original Item", 1, 100)
	s.Require().NoError(err)

	update := map[string]interface{}{
		"description": "Updated Item",
		"quantity":    5,
		"unit_price":  200.00,
	}

	resp, err := s.setup.MakeRequest("PUT", "/api/invoices/"+uintToString(invoiceID)+"/items/"+uintToString(itemID), update)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Updated Item", result["description"])
	s.Equal(float64(5), result["quantity"])
	s.Equal(float64(200), result["unit_price"])
}

func (s *InvoiceTestSuite) TestDeleteInvoiceItem() {
	categoryID, _ := s.setup.CreateTestCategory("Test")
	companyID, _ := s.setup.CreateTestCompany("Test")

	invoiceID, err := s.setup.CreateTestInvoice("Test Invoice", &categoryID, &companyID)
	s.Require().NoError(err)

	itemID, err := s.setup.CreateTestInvoiceItem(invoiceID, "To Delete", 1, 100)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("DELETE", "/api/invoices/"+uintToString(invoiceID)+"/items/"+uintToString(itemID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)
}

func TestInvoiceSuite(t *testing.T) {
	suite.Run(t, new(InvoiceTestSuite))
}
