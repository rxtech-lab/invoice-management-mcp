package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/stretchr/testify/suite"
)

type FXTestSuite struct {
	suite.Suite
	setup     *TestSetup
	fxService *services.MockFXService
}

func (s *FXTestSuite) SetupTest() {
	// Create mock FX service with known rates
	s.fxService = services.NewMockFXService()

	// Set up exchange rates for testing
	// HKD -> USD: 1 HKD = 0.125 USD (i.e., 1 USD = 8 HKD)
	s.fxService.SetRate("HKD", "USD", 0.125)
	// USD -> USD is always 1.0 (handled automatically)

	s.setup = NewTestSetupWithFXService(s.T(), s.fxService)
}

func (s *FXTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

// TestCurrencyConversion_USDToHKDToUSD tests the full currency conversion cycle
// Scenario: USD -> HKD -> USD with FX rate 1 USD = 8 HKD
func (s *FXTestSuite) TestCurrencyConversion_USDToHKDToUSD() {
	// Step 1: Create USD invoice with 3 items x 10 USD each
	invoiceID, err := s.setup.CreateTestInvoiceWithCurrency("FX Test Invoice", "USD")
	s.Require().NoError(err)

	// Create 3 items with quantity=1, unit_price=10 each
	var itemIDs []uint
	for i := 0; i < 3; i++ {
		itemID, err := s.setup.CreateTestInvoiceItem(invoiceID, fmt.Sprintf("Item %d", i+1), 1, 10.00)
		s.Require().NoError(err)
		itemIDs = append(itemIDs, itemID)
	}

	// Step 2: Verify initial state - USD invoice, items have amount=10, target_amount=10
	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	invoice, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("USD", invoice["currency"])
	s.Equal(float64(30), invoice["amount"]) // 3 items x 10 = 30

	items := invoice["items"].([]interface{})
	s.Equal(3, len(items))
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		s.Equal(float64(10), itemMap["amount"])
		s.Equal(float64(10), itemMap["target_amount"]) // USD->USD = 1:1
		s.Equal(float64(1), itemMap["fx_rate_used"])
		s.Equal("USD", itemMap["target_currency"])
	}

	// Step 3: Change currency to HKD
	resp, err = s.setup.UpdateInvoiceCurrency(invoiceID, "HKD")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	// Step 4: Verify items - amount unchanged, target_amount recalculated
	// 10 HKD * 0.125 = 1.25 USD
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	invoice, _ = s.setup.ReadResponseBody(resp)

	s.Equal("HKD", invoice["currency"])
	s.Equal(float64(30), invoice["amount"]) // Amount unchanged (still 30)

	items = invoice["items"].([]interface{})
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		s.Equal(float64(10), itemMap["amount"])          // Amount unchanged
		s.Equal(float64(1.25), itemMap["target_amount"]) // 10 HKD * 0.125 = 1.25 USD
		s.Equal(float64(0.125), itemMap["fx_rate_used"])
		s.Equal("USD", itemMap["target_currency"])
	}

	// Step 5: Manually override target_amount for first item to 90
	targetAmountOverride := 90.0
	resp, err = s.setup.UpdateInvoiceItemWithTargetAmount(
		invoiceID, itemIDs[0], "Item 1", 1, 10.00, &targetAmountOverride)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	// Verify the override
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	invoice, _ = s.setup.ReadResponseBody(resp)

	items = invoice["items"].([]interface{})
	// Find item with id matching itemIDs[0]
	var item0 map[string]interface{}
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		if uint(itemMap["id"].(float64)) == itemIDs[0] {
			item0 = itemMap
			break
		}
	}
	s.Require().NotNil(item0, "Item 0 not found")
	s.Equal(float64(10), item0["amount"])
	s.Equal(float64(90), item0["target_amount"]) // Manually overridden
	// FX rate is recalculated: 90 / 10 = 9.0
	s.Equal(float64(9), item0["fx_rate_used"])

	// Step 6: Update unit_price of the manually overridden item to 20
	// This should recalculate target_amount (no longer using manual override)
	resp, err = s.setup.UpdateInvoiceItemWithTargetAmount(
		invoiceID, itemIDs[0], "Item 1", 1, 20.00, nil) // nil = auto-calculate
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	// Verify - amount = 20, target_amount = 20 * 0.125 = 2.5 USD
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	invoice, _ = s.setup.ReadResponseBody(resp)

	items = invoice["items"].([]interface{})
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		if uint(itemMap["id"].(float64)) == itemIDs[0] {
			item0 = itemMap
			break
		}
	}
	s.Equal(float64(20), item0["amount"])         // 1 * 20 = 20
	s.Equal(float64(2.5), item0["target_amount"]) // 20 HKD * 0.125 = 2.5 USD
	s.Equal(float64(0.125), item0["fx_rate_used"])

	// Step 7: Change back to USD
	resp, err = s.setup.UpdateInvoiceCurrency(invoiceID, "USD")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	// Step 8: Verify - all items should have target_amount = amount (USD->USD = 1:1)
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	invoice, _ = s.setup.ReadResponseBody(resp)

	s.Equal("USD", invoice["currency"])

	items = invoice["items"].([]interface{})
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		itemID := uint(itemMap["id"].(float64))
		amount := itemMap["amount"].(float64)
		targetAmount := itemMap["target_amount"].(float64)
		fxRate := itemMap["fx_rate_used"].(float64)

		// First item has amount=20, others have amount=10
		if itemID == itemIDs[0] {
			s.Equal(float64(20), amount)       // Still 20 (amount unchanged)
			s.Equal(float64(20), targetAmount) // USD->USD = 1:1
		} else {
			s.Equal(float64(10), amount)
			s.Equal(float64(10), targetAmount) // USD->USD = 1:1
		}
		s.Equal(float64(1), fxRate)
	}
}

// TestCurrencyConversion_HKDToUSD tests converting from HKD to USD
// Scenario: Create HKD invoice with items, then change to USD
func (s *FXTestSuite) TestCurrencyConversion_HKDToUSD() {
	// Step 1: Create HKD invoice with 3 items x 80 HKD each
	invoiceID, err := s.setup.CreateTestInvoiceWithCurrency("HKD Invoice", "HKD")
	s.Require().NoError(err)

	// Create 3 items with quantity=1, unit_price=80 each
	for i := 0; i < 3; i++ {
		_, err := s.setup.CreateTestInvoiceItem(invoiceID, fmt.Sprintf("Item %d", i+1), 1, 80.00)
		s.Require().NoError(err)
	}

	// Step 2: Verify initial state
	// HKD invoice, items have amount=80, target_amount = 80 * 0.125 = 10 USD
	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	invoice, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("HKD", invoice["currency"])
	s.Equal(float64(240), invoice["amount"]) // 3 items x 80 = 240 HKD

	items := invoice["items"].([]interface{})
	s.Equal(3, len(items))
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		s.Equal(float64(80), itemMap["amount"])
		s.Equal(float64(10), itemMap["target_amount"]) // 80 HKD * 0.125 = 10 USD
		s.Equal(float64(0.125), itemMap["fx_rate_used"])
		s.Equal("USD", itemMap["target_currency"])
	}

	// Step 3: Change currency to USD
	resp, err = s.setup.UpdateInvoiceCurrency(invoiceID, "USD")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	// Step 4: Verify items - amount unchanged (80), target_amount = 80 USD (1:1)
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	invoice, _ = s.setup.ReadResponseBody(resp)

	s.Equal("USD", invoice["currency"])
	s.Equal(float64(240), invoice["amount"]) // Amount unchanged

	items = invoice["items"].([]interface{})
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		s.Equal(float64(80), itemMap["amount"])        // Amount unchanged
		s.Equal(float64(80), itemMap["target_amount"]) // USD->USD = 1:1
		s.Equal(float64(1), itemMap["fx_rate_used"])
		s.Equal("USD", itemMap["target_currency"])
	}
}

// TestCreateInvoiceWithItems_FXCalculation tests FX calculation when creating invoice with items
func (s *FXTestSuite) TestCreateInvoiceWithItems_FXCalculation() {
	// Create HKD invoice with items in single request
	invoice := map[string]interface{}{
		"title":    "Invoice with Items",
		"currency": "HKD",
		"items": []map[string]interface{}{
			{"description": "Item 1", "quantity": 2, "unit_price": 40.00},  // 80 HKD
			{"description": "Item 2", "quantity": 1, "unit_price": 160.00}, // 160 HKD
		},
	}

	resp, err := s.setup.MakeRequest("POST", "/api/invoices", invoice)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, _ := s.setup.ReadResponseBody(resp)

	// Verify FX calculation on creation
	items := result["items"].([]interface{})
	s.Equal(2, len(items))

	// Find items by description
	var item1, item2 map[string]interface{}
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		if itemMap["description"] == "Item 1" {
			item1 = itemMap
		} else if itemMap["description"] == "Item 2" {
			item2 = itemMap
		}
	}

	s.Require().NotNil(item1, "Item 1 not found")
	s.Require().NotNil(item2, "Item 2 not found")

	s.Equal(float64(80), item1["amount"])        // 2 * 40 = 80
	s.Equal(float64(10), item1["target_amount"]) // 80 * 0.125 = 10 USD

	s.Equal(float64(160), item2["amount"])       // 1 * 160 = 160
	s.Equal(float64(20), item2["target_amount"]) // 160 * 0.125 = 20 USD
}

// TestManualTargetAmountOverride tests that manual target_amount override works correctly
func (s *FXTestSuite) TestManualTargetAmountOverride() {
	invoiceID, err := s.setup.CreateTestInvoiceWithCurrency("Override Test", "HKD")
	s.Require().NoError(err)
	itemID, err := s.setup.CreateTestInvoiceItem(invoiceID, "Test Item", 1, 100.00)
	s.Require().NoError(err)

	// Manual override
	targetAmount := 50.0
	resp, err := s.setup.UpdateInvoiceItemWithTargetAmount(invoiceID, itemID, "Test Item", 1, 100.00, &targetAmount)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	// Verify
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	invoice, _ := s.setup.ReadResponseBody(resp)

	items := invoice["items"].([]interface{})
	item := items[0].(map[string]interface{})
	s.Equal(float64(100), item["amount"])
	s.Equal(float64(50), item["target_amount"])
	s.Equal(float64(0.5), item["fx_rate_used"]) // 50 / 100 = 0.5
}

// TestMultipleItemsFXCalculation tests that multiple items have correct FX calculations
func (s *FXTestSuite) TestMultipleItemsFXCalculation() {
	invoiceID, err := s.setup.CreateTestInvoiceWithCurrency("Multi Item Test", "HKD")
	s.Require().NoError(err)

	// Create items with different amounts
	_, err = s.setup.CreateTestInvoiceItem(invoiceID, "Item 1", 1, 80.00) // target: 10 USD
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(invoiceID, "Item 2", 1, 160.00) // target: 20 USD
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(invoiceID, "Item 3", 1, 240.00) // target: 30 USD
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	invoice, _ := s.setup.ReadResponseBody(resp)

	s.Equal(float64(480), invoice["amount"]) // 80 + 160 + 240 = 480 HKD

	// Verify each item's target_amount
	items := invoice["items"].([]interface{})
	expectedTargets := map[string]float64{
		"Item 1": 10, // 80 * 0.125
		"Item 2": 20, // 160 * 0.125
		"Item 3": 30, // 240 * 0.125
	}
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		desc := itemMap["description"].(string)
		s.Equal(expectedTargets[desc], itemMap["target_amount"], "target_amount mismatch for %s", desc)
	}
}

// TestFXRateUsedTracking tests that fx_rate_used is correctly tracked
func (s *FXTestSuite) TestFXRateUsedTracking() {
	// Test with different FX rates
	s.fxService.SetRate("EUR", "USD", 1.1) // 1 EUR = 1.1 USD

	invoiceID, err := s.setup.CreateTestInvoiceWithCurrency("EUR Invoice", "EUR")
	s.Require().NoError(err)

	_, err = s.setup.CreateTestInvoiceItem(invoiceID, "EUR Item", 1, 100.00)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/invoices/%d", invoiceID), nil)
	s.Require().NoError(err)
	invoice, _ := s.setup.ReadResponseBody(resp)

	items := invoice["items"].([]interface{})
	item := items[0].(map[string]interface{})

	s.Equal(float64(100), item["amount"])
	s.InDelta(float64(110), item["target_amount"].(float64), 0.001) // 100 * 1.1 = 110 USD
	s.InDelta(float64(1.1), item["fx_rate_used"].(float64), 0.001)
}

func TestFXSuite(t *testing.T) {
	suite.Run(t, new(FXTestSuite))
}
