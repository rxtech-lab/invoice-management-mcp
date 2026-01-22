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

	// Create invoices with different amounts to avoid duplicate detection
	invoiceID1, err := s.setup.CreateTestInvoice("Invoice 1", &categoryID, &companyID)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(invoiceID1, "Item 1", 1, 100.00)
	s.Require().NoError(err)

	invoiceID2, err := s.setup.CreateTestInvoice("Invoice 2", &categoryID, &companyID)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(invoiceID2, "Item 2", 1, 200.00) // Different amount
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

func (s *InvoiceTestSuite) TestCreateInvoice_DuplicateDetection() {
	// Create a receiver for the test
	receiverID, err := s.setup.CreateTestReceiver("Test Receiver", true)
	s.Require().NoError(err)

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	// Create first invoice with specific dates and receiver
	invoice1 := map[string]interface{}{
		"title":              "January Invoice",
		"description":        "Monthly services",
		"currency":           "USD",
		"receiver_id":        receiverID,
		"invoice_started_at": startDate.Format(time.RFC3339),
		"invoice_ended_at":   endDate.Format(time.RFC3339),
		"status":             "unpaid",
		"items": []map[string]interface{}{
			{
				"description": "Service",
				"quantity":    1,
				"unit_price":  100.00,
			},
		},
	}

	resp1, err := s.setup.MakeRequest("POST", "/api/invoices", invoice1)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp1.StatusCode)

	result1, err := s.setup.ReadResponseBody(resp1)
	s.Require().NoError(err)
	firstInvoiceID := result1["id"]

	// Create second invoice with same amount (from items), dates, and receiver - should be detected as duplicate
	invoice2 := map[string]interface{}{
		"title":              "Different Title - Same Invoice",
		"description":        "Different description",
		"currency":           "EUR", // Different currency
		"receiver_id":        receiverID,
		"invoice_started_at": startDate.Format(time.RFC3339),
		"invoice_ended_at":   endDate.Format(time.RFC3339),
		"status":             "paid",
		"items": []map[string]interface{}{
			{
				"description": "Different item description",
				"quantity":    2,
				"unit_price":  50.00, // Same total: 2 * 50 = 100
			},
		},
	}

	resp2, err := s.setup.MakeRequest("POST", "/api/invoices", invoice2)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp2.StatusCode)

	result2, err := s.setup.ReadResponseBody(resp2)
	s.Require().NoError(err)

	// Should return the first invoice (duplicate detected)
	s.Equal(firstInvoiceID, result2["id"], "Should return the existing invoice ID when duplicate detected")
	s.Equal("January Invoice", result2["title"], "Should return the original invoice title")
}

func (s *InvoiceTestSuite) TestCreateInvoice_DifferentAmountNotDuplicate() {
	receiverID, err := s.setup.CreateTestReceiver("Test Receiver 2", true)
	s.Require().NoError(err)

	startDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)

	// Create first invoice
	invoice1 := map[string]interface{}{
		"title":              "February Invoice 1",
		"receiver_id":        receiverID,
		"invoice_started_at": startDate.Format(time.RFC3339),
		"invoice_ended_at":   endDate.Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Service", "quantity": 1, "unit_price": 100.00},
		},
	}

	resp1, err := s.setup.MakeRequest("POST", "/api/invoices", invoice1)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp1.StatusCode)

	result1, err := s.setup.ReadResponseBody(resp1)
	s.Require().NoError(err)
	firstInvoiceID := result1["id"]

	// Create second invoice with different amount - should NOT be duplicate
	invoice2 := map[string]interface{}{
		"title":              "February Invoice 2",
		"receiver_id":        receiverID,
		"invoice_started_at": startDate.Format(time.RFC3339),
		"invoice_ended_at":   endDate.Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Service", "quantity": 1, "unit_price": 200.00}, // Different amount
		},
	}

	resp2, err := s.setup.MakeRequest("POST", "/api/invoices", invoice2)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp2.StatusCode)

	result2, err := s.setup.ReadResponseBody(resp2)
	s.Require().NoError(err)

	// Should be a new invoice (different amount)
	s.NotEqual(firstInvoiceID, result2["id"], "Should create new invoice when amount differs")
}

func (s *InvoiceTestSuite) TestCreateInvoice_DifferentDatesNotDuplicate() {
	receiverID, err := s.setup.CreateTestReceiver("Test Receiver 3", true)
	s.Require().NoError(err)

	// Create first invoice with January dates
	invoice1 := map[string]interface{}{
		"title":              "January Invoice",
		"receiver_id":        receiverID,
		"invoice_started_at": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"invoice_ended_at":   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Service", "quantity": 1, "unit_price": 100.00},
		},
	}

	resp1, err := s.setup.MakeRequest("POST", "/api/invoices", invoice1)
	s.Require().NoError(err)

	result1, err := s.setup.ReadResponseBody(resp1)
	s.Require().NoError(err)
	firstInvoiceID := result1["id"]

	// Create second invoice with February dates (same amount) - should NOT be duplicate
	invoice2 := map[string]interface{}{
		"title":              "February Invoice",
		"receiver_id":        receiverID,
		"invoice_started_at": time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"invoice_ended_at":   time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Service", "quantity": 1, "unit_price": 100.00}, // Same amount
		},
	}

	resp2, err := s.setup.MakeRequest("POST", "/api/invoices", invoice2)
	s.Require().NoError(err)

	result2, err := s.setup.ReadResponseBody(resp2)
	s.Require().NoError(err)

	// Should be a new invoice (different dates)
	s.NotEqual(firstInvoiceID, result2["id"], "Should create new invoice when dates differ")
}

func (s *InvoiceTestSuite) TestCreateInvoice_DifferentReceiverNotDuplicate() {
	receiverID1, err := s.setup.CreateTestReceiver("Receiver 1", true)
	s.Require().NoError(err)
	receiverID2, err := s.setup.CreateTestReceiver("Receiver 2", true)
	s.Require().NoError(err)

	startDate := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

	// Create first invoice with receiver 1
	invoice1 := map[string]interface{}{
		"title":              "March Invoice 1",
		"receiver_id":        receiverID1,
		"invoice_started_at": startDate.Format(time.RFC3339),
		"invoice_ended_at":   endDate.Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Service", "quantity": 1, "unit_price": 100.00},
		},
	}

	resp1, err := s.setup.MakeRequest("POST", "/api/invoices", invoice1)
	s.Require().NoError(err)

	result1, err := s.setup.ReadResponseBody(resp1)
	s.Require().NoError(err)
	firstInvoiceID := result1["id"]

	// Create second invoice with different receiver (same amount, dates) - should NOT be duplicate
	invoice2 := map[string]interface{}{
		"title":              "March Invoice 2",
		"receiver_id":        receiverID2,
		"invoice_started_at": startDate.Format(time.RFC3339),
		"invoice_ended_at":   endDate.Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Service", "quantity": 1, "unit_price": 100.00}, // Same amount
		},
	}

	resp2, err := s.setup.MakeRequest("POST", "/api/invoices", invoice2)
	s.Require().NoError(err)

	result2, err := s.setup.ReadResponseBody(resp2)
	s.Require().NoError(err)

	// Should be a new invoice (different receiver)
	s.NotEqual(firstInvoiceID, result2["id"], "Should create new invoice when receiver differs")
}

func (s *InvoiceTestSuite) TestCreateInvoice_NullDatesMatchNull() {
	receiverID, err := s.setup.CreateTestReceiver("Test Receiver Null", true)
	s.Require().NoError(err)

	// Create first invoice with no dates
	invoice1 := map[string]interface{}{
		"title":       "Invoice Without Dates 1",
		"receiver_id": receiverID,
		// No invoice_started_at or invoice_ended_at
		"items": []map[string]interface{}{
			{"description": "Service", "quantity": 1, "unit_price": 100.00},
		},
	}

	resp1, err := s.setup.MakeRequest("POST", "/api/invoices", invoice1)
	s.Require().NoError(err)

	result1, err := s.setup.ReadResponseBody(resp1)
	s.Require().NoError(err)
	firstInvoiceID := result1["id"]

	// Create second invoice with no dates (same amount, receiver) - should be duplicate
	invoice2 := map[string]interface{}{
		"title":       "Invoice Without Dates 2",
		"receiver_id": receiverID,
		// No invoice_started_at or invoice_ended_at
		"items": []map[string]interface{}{
			{"description": "Other Service", "quantity": 2, "unit_price": 50.00}, // Same total: 100
		},
	}

	resp2, err := s.setup.MakeRequest("POST", "/api/invoices", invoice2)
	s.Require().NoError(err)

	result2, err := s.setup.ReadResponseBody(resp2)
	s.Require().NoError(err)

	// Should return the first invoice (null dates match null dates)
	s.Equal(firstInvoiceID, result2["id"], "Invoices with null dates should be detected as duplicates")
}

func (s *InvoiceTestSuite) TestCreateInvoice_NullReceiverMatchNull() {
	startDate := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC)

	// Create first invoice with no receiver
	invoice1 := map[string]interface{}{
		"title":              "Invoice Without Receiver 1",
		"invoice_started_at": startDate.Format(time.RFC3339),
		"invoice_ended_at":   endDate.Format(time.RFC3339),
		// No receiver_id
		"items": []map[string]interface{}{
			{"description": "Service", "quantity": 1, "unit_price": 100.00},
		},
	}

	resp1, err := s.setup.MakeRequest("POST", "/api/invoices", invoice1)
	s.Require().NoError(err)

	result1, err := s.setup.ReadResponseBody(resp1)
	s.Require().NoError(err)
	firstInvoiceID := result1["id"]

	// Create second invoice with no receiver (same amount, dates) - should be duplicate
	invoice2 := map[string]interface{}{
		"title":              "Invoice Without Receiver 2",
		"invoice_started_at": startDate.Format(time.RFC3339),
		"invoice_ended_at":   endDate.Format(time.RFC3339),
		// No receiver_id
		"items": []map[string]interface{}{
			{"description": "Other Service", "quantity": 2, "unit_price": 50.00}, // Same total: 100
		},
	}

	resp2, err := s.setup.MakeRequest("POST", "/api/invoices", invoice2)
	s.Require().NoError(err)

	result2, err := s.setup.ReadResponseBody(resp2)
	s.Require().NoError(err)

	// Should return the first invoice (null receiver matches null receiver)
	s.Equal(firstInvoiceID, result2["id"], "Invoices with null receiver should be detected as duplicates")
}

func TestInvoiceSuite(t *testing.T) {
	suite.Run(t, new(InvoiceTestSuite))
}
