package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/stretchr/testify/suite"
)

// AnalyticsCurrencyTestSuite tests that analytics properly use item-level target_amount
// for currency conversion instead of relying on invoice-level cached values
type AnalyticsCurrencyTestSuite struct {
	suite.Suite
	setup     *TestSetup
	fxService *services.MockFXService
}

func (s *AnalyticsCurrencyTestSuite) SetupTest() {
	// Create mock FX service with known rates
	s.fxService = services.NewMockFXService()

	// Set up exchange rates for testing:
	// HKD -> USD: 1 HKD = 0.128 USD (approximately 1 USD = 7.8 HKD)
	// EUR -> USD: 1 EUR = 1.1 USD
	s.fxService.SetRate("HKD", "USD", 0.128)
	s.fxService.SetRate("EUR", "USD", 1.1)

	s.setup = NewTestSetupWithFXService(s.T(), s.fxService)
}

func (s *AnalyticsCurrencyTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

// TestAnalyticsSummary_MultiCurrency verifies that analytics summary correctly
// aggregates invoices in different currencies using their USD-converted values
func (s *AnalyticsCurrencyTestSuite) TestAnalyticsSummary_MultiCurrency() {
	// Create USD invoice: 100 USD -> target_amount = 100 USD
	usdInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("USD Invoice", "USD")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(usdInvoiceID, "USD Item", 1, 100.00)
	s.Require().NoError(err)

	// Create HKD invoice: 780 HKD -> target_amount = 780 * 0.128 = 99.84 USD
	hkdInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("HKD Invoice", "HKD")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(hkdInvoiceID, "HKD Item", 1, 780.00)
	s.Require().NoError(err)

	// Create EUR invoice: 100 EUR -> target_amount = 100 * 1.1 = 110 USD
	eurInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("EUR Invoice", "EUR")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(eurInvoiceID, "EUR Item", 1, 100.00)
	s.Require().NoError(err)

	// Mark USD and EUR invoices as paid
	_, err = s.setup.MakeRequest("PATCH", fmt.Sprintf("/api/invoices/%d/status", usdInvoiceID), map[string]string{"status": "paid"})
	s.Require().NoError(err)
	_, err = s.setup.MakeRequest("PATCH", fmt.Sprintf("/api/invoices/%d/status", eurInvoiceID), map[string]string{"status": "paid"})
	s.Require().NoError(err)

	// Get analytics summary
	summary, err := s.setup.AnalyticsService.GetSummary(s.setup.TestUserID, services.Period1Month)
	s.Require().NoError(err)

	// Expected totals (all in USD):
	// Total: 100 + 99.84 + 110 = 309.84 USD
	// Paid: 100 + 110 = 210 USD
	// Unpaid: 99.84 USD (HKD invoice)
	s.InDelta(309.84, summary.TotalAmount, 0.01, "Total amount should be sum of USD-converted values")
	s.InDelta(210.0, summary.PaidAmount, 0.01, "Paid amount should be sum of paid invoices in USD")
	s.InDelta(99.84, summary.UnpaidAmount, 0.01, "Unpaid amount should be HKD invoice in USD")
	s.Equal(int64(3), summary.InvoiceCount)
}

// TestAnalyticsByCompany_MultiCurrency verifies that company breakdown uses USD values
// This matches the "By Company" chart scenario from the bug report
func (s *AnalyticsCurrencyTestSuite) TestAnalyticsByCompany_MultiCurrency() {
	// Create a company
	companyID, err := s.setup.CreateTestCompany("Test Company HK")
	s.Require().NoError(err)

	// Create HKD invoice for this company: 4117.66 HKD -> 527.06 USD
	hkdInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("HKD Company Invoice", "HKD")
	s.Require().NoError(err)

	// Update invoice to assign company
	_, err = s.setup.MakeRequest("PUT", fmt.Sprintf("/api/invoices/%d", hkdInvoiceID), map[string]interface{}{
		"company_id": companyID,
		"currency":   "HKD",
	})
	s.Require().NoError(err)

	// Add item with amount matching the bug report: 4117.66 HKD
	_, err = s.setup.CreateTestInvoiceItem(hkdInvoiceID, "HKD Item", 1, 4117.66)
	s.Require().NoError(err)

	// Mark as paid
	_, err = s.setup.MakeRequest("PATCH", fmt.Sprintf("/api/invoices/%d/status", hkdInvoiceID), map[string]string{"status": "paid"})
	s.Require().NoError(err)

	// Create USD invoice for same company: 500 USD
	usdInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("USD Company Invoice", "USD")
	s.Require().NoError(err)
	_, err = s.setup.MakeRequest("PUT", fmt.Sprintf("/api/invoices/%d", usdInvoiceID), map[string]interface{}{
		"company_id": companyID,
		"currency":   "USD",
	})
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(usdInvoiceID, "USD Item", 1, 500.00)
	s.Require().NoError(err)

	// Get analytics by company
	byCompany, err := s.setup.AnalyticsService.GetByCompany(s.setup.TestUserID, services.Period1Month)
	s.Require().NoError(err)

	// Find our test company
	var testCompany *services.AnalyticsGroupItem
	for i := range byCompany.Items {
		if byCompany.Items[i].Name == "Test Company HK" {
			testCompany = &byCompany.Items[i]
			break
		}
	}
	s.Require().NotNil(testCompany, "Test company should be in results")

	// Expected: 4117.66 * 0.128 + 500 = 527.06 + 500 = 1027.06 USD
	expectedTotal := 4117.66*0.128 + 500.00
	s.InDelta(expectedTotal, testCompany.TotalAmount, 0.01,
		"Company total should be USD-converted (got %f, expected %f)", testCompany.TotalAmount, expectedTotal)

	// Paid should be just the HKD invoice: 527.06 USD
	expectedPaid := 4117.66 * 0.128
	s.InDelta(expectedPaid, testCompany.PaidAmount, 0.01,
		"Company paid amount should be HKD invoice in USD (got %f, expected %f)", testCompany.PaidAmount, expectedPaid)
}

// TestAnalyticsByCategory_MultiCurrency verifies that category breakdown uses USD values
func (s *AnalyticsCurrencyTestSuite) TestAnalyticsByCategory_MultiCurrency() {
	// Create a category
	categoryID, err := s.setup.CreateTestCategory("Rent")
	s.Require().NoError(err)

	// Create HKD invoice: 1000 HKD -> 128 USD
	hkdInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("HKD Rent", "HKD")
	s.Require().NoError(err)
	_, err = s.setup.MakeRequest("PUT", fmt.Sprintf("/api/invoices/%d", hkdInvoiceID), map[string]interface{}{
		"category_id": categoryID,
		"currency":    "HKD",
	})
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(hkdInvoiceID, "HKD Rent Item", 1, 1000.00)
	s.Require().NoError(err)

	// Create EUR invoice: 200 EUR -> 220 USD
	eurInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("EUR Rent", "EUR")
	s.Require().NoError(err)
	_, err = s.setup.MakeRequest("PUT", fmt.Sprintf("/api/invoices/%d", eurInvoiceID), map[string]interface{}{
		"category_id": categoryID,
		"currency":    "EUR",
	})
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(eurInvoiceID, "EUR Rent Item", 1, 200.00)
	s.Require().NoError(err)

	// Get analytics by category
	byCategory, err := s.setup.AnalyticsService.GetByCategory(s.setup.TestUserID, services.Period1Month)
	s.Require().NoError(err)

	// Find our test category
	var rentCategory *services.AnalyticsGroupItem
	for i := range byCategory.Items {
		if byCategory.Items[i].Name == "Rent" {
			rentCategory = &byCategory.Items[i]
			break
		}
	}
	s.Require().NotNil(rentCategory, "Rent category should be in results")

	// Expected: 1000 * 0.128 + 200 * 1.1 = 128 + 220 = 348 USD
	expectedTotal := 1000.0*0.128 + 200.0*1.1
	s.InDelta(expectedTotal, rentCategory.TotalAmount, 0.01,
		"Category total should be USD-converted")
}

// TestAnalyticsByReceiver_MultiCurrency verifies that receiver breakdown uses USD values
func (s *AnalyticsCurrencyTestSuite) TestAnalyticsByReceiver_MultiCurrency() {
	// Create a receiver
	receiverID, err := s.setup.CreateTestReceiver("John Smith", false)
	s.Require().NoError(err)

	// Create HKD invoice: 500 HKD -> 64 USD
	hkdInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("HKD Payment", "HKD")
	s.Require().NoError(err)
	_, err = s.setup.MakeRequest("PUT", fmt.Sprintf("/api/invoices/%d", hkdInvoiceID), map[string]interface{}{
		"receiver_id": receiverID,
		"currency":    "HKD",
	})
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(hkdInvoiceID, "HKD Payment Item", 1, 500.00)
	s.Require().NoError(err)

	// Get analytics by receiver
	byReceiver, err := s.setup.AnalyticsService.GetByReceiver(s.setup.TestUserID, services.Period1Month)
	s.Require().NoError(err)

	// Find our test receiver
	var johnReceiver *services.AnalyticsGroupItem
	for i := range byReceiver.Items {
		if byReceiver.Items[i].Name == "John Smith" {
			johnReceiver = &byReceiver.Items[i]
			break
		}
	}
	s.Require().NotNil(johnReceiver, "John Smith should be in results")

	// Expected: 500 * 0.128 = 64 USD
	expectedTotal := 500.0 * 0.128
	s.InDelta(expectedTotal, johnReceiver.TotalAmount, 0.01,
		"Receiver total should be USD-converted")
}

// TestAnalyticsByTag_MultiCurrency verifies that tag breakdown uses USD values
func (s *AnalyticsCurrencyTestSuite) TestAnalyticsByTag_MultiCurrency() {
	// First create a tag
	tagResp, err := s.setup.MakeRequest("POST", "/api/tags", map[string]interface{}{
		"name":  "monthly",
		"color": "#FF5733",
	})
	s.Require().NoError(err)
	s.Require().Equal(201, tagResp.StatusCode)

	tagBody, err := s.setup.ReadResponseBody(tagResp)
	s.Require().NoError(err)
	tagID := int(tagBody["id"].(float64))

	// Create HKD invoice with the tag: 250 HKD -> 32 USD
	invoice := map[string]interface{}{
		"title":    "HKD Tagged Invoice",
		"currency": "HKD",
		"tag_ids":  []int{tagID},
		"items": []map[string]interface{}{
			{"description": "HKD Tagged Item", "quantity": 1, "unit_price": 250.00},
		},
	}
	resp, err := s.setup.MakeRequest("POST", "/api/invoices", invoice)
	s.Require().NoError(err)
	s.Require().Equal(201, resp.StatusCode)

	// Get analytics by tag
	byTag, err := s.setup.AnalyticsService.GetByTag(s.setup.TestUserID, services.Period1Month)
	s.Require().NoError(err)

	// Find our tag
	var monthlyTag *services.AnalyticsGroupItem
	for i := range byTag.Items {
		if byTag.Items[i].Name == "monthly" {
			monthlyTag = &byTag.Items[i]
			break
		}
	}
	s.Require().NotNil(monthlyTag, "monthly tag should be in results")

	// Expected: 250 * 0.128 = 32 USD
	expectedTotal := 250.0 * 0.128
	s.InDelta(expectedTotal, monthlyTag.TotalAmount, 0.01,
		"Tag total should be USD-converted")
}

// TestStatistics_GroupByDay_MultiCurrency verifies that daily breakdown uses USD values
// This matches the "Spending Trend" chart scenario from the bug report
func (s *AnalyticsCurrencyTestSuite) TestStatistics_GroupByDay_MultiCurrency() {
	today := time.Now()

	// Create HKD invoice dated today: 780 HKD -> 99.84 USD
	hkdInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("HKD Today", "HKD")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(hkdInvoiceID, "HKD Item", 1, 780.00)
	s.Require().NoError(err)

	// Update created_at to today
	db := s.setup.DBService.GetDB()
	err = db.Exec("UPDATE invoices SET created_at = ? WHERE id = ?", today, hkdInvoiceID).Error
	s.Require().NoError(err)

	// Create USD invoice dated yesterday: 100 USD
	yesterday := today.AddDate(0, 0, -1)
	usdInvoiceID, err := s.setup.CreateTestInvoiceWithCurrency("USD Yesterday", "USD")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(usdInvoiceID, "USD Item", 1, 100.00)
	s.Require().NoError(err)

	err = db.Exec("UPDATE invoices SET created_at = ? WHERE id = ?", yesterday, usdInvoiceID).Error
	s.Require().NoError(err)

	// Get statistics grouped by day
	opts := services.StatisticsOptions{
		Period:  services.PeriodLastWeek,
		GroupBy: services.GroupByDay,
	}
	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Find today's entry
	todayStr := today.Format("2006-01-02")
	var todayEntry *services.BreakdownItem
	for i := range stats.Breakdown {
		if stats.Breakdown[i].Date == todayStr {
			todayEntry = &stats.Breakdown[i]
			break
		}
	}
	s.Require().NotNil(todayEntry, "Today should be in breakdown")

	// Expected: 780 * 0.128 = 99.84 USD (not 780 HKD!)
	expectedTodayAmount := 780.0 * 0.128
	s.InDelta(expectedTodayAmount, todayEntry.Amount, 0.01,
		"Today's amount should be USD-converted (got %f, expected %f)", todayEntry.Amount, expectedTodayAmount)
}

// TestAnalytics_UsesItemTargetAmount verifies that analytics uses sum of
// invoice_items.target_amount, NOT invoices.target_amount
func (s *AnalyticsCurrencyTestSuite) TestAnalytics_UsesItemTargetAmount() {
	// Create HKD invoice with items
	invoiceID, err := s.setup.CreateTestInvoiceWithCurrency("Item Target Test", "HKD")
	s.Require().NoError(err)

	// Add two items:
	// Item 1: 100 HKD -> target_amount = 12.8 USD
	// Item 2: 200 HKD -> target_amount = 25.6 USD
	_, err = s.setup.CreateTestInvoiceItem(invoiceID, "Item 1", 1, 100.00)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(invoiceID, "Item 2", 1, 200.00)
	s.Require().NoError(err)

	// Get analytics summary
	summary, err := s.setup.AnalyticsService.GetSummary(s.setup.TestUserID, services.Period1Month)
	s.Require().NoError(err)

	// Expected: (100 + 200) * 0.128 = 38.4 USD from item target_amounts
	expectedTotal := 300.0 * 0.128
	s.InDelta(expectedTotal, summary.TotalAmount, 0.01,
		"Analytics should use sum of item target_amounts")
}

// TestAnalyticsSummary_MixedStatusMultiCurrency verifies paid/unpaid/overdue breakdown
// with multiple currencies
func (s *AnalyticsCurrencyTestSuite) TestAnalyticsSummary_MixedStatusMultiCurrency() {
	// Create paid HKD invoice: 1000 HKD -> 128 USD
	paidHKDID, err := s.setup.CreateTestInvoiceWithCurrency("Paid HKD", "HKD")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(paidHKDID, "Item", 1, 1000.00)
	s.Require().NoError(err)
	_, err = s.setup.MakeRequest("PATCH", fmt.Sprintf("/api/invoices/%d/status", paidHKDID), map[string]string{"status": "paid"})
	s.Require().NoError(err)

	// Create unpaid EUR invoice: 100 EUR -> 110 USD
	unpaidEURID, err := s.setup.CreateTestInvoiceWithCurrency("Unpaid EUR", "EUR")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(unpaidEURID, "Item", 1, 100.00)
	s.Require().NoError(err)
	// Default status is unpaid

	// Create overdue USD invoice: 50 USD
	overdueUSDID, err := s.setup.CreateTestInvoiceWithCurrency("Overdue USD", "USD")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceItem(overdueUSDID, "Item", 1, 50.00)
	s.Require().NoError(err)
	_, err = s.setup.MakeRequest("PATCH", fmt.Sprintf("/api/invoices/%d/status", overdueUSDID), map[string]string{"status": "overdue"})
	s.Require().NoError(err)

	// Get analytics summary
	summary, err := s.setup.AnalyticsService.GetSummary(s.setup.TestUserID, services.Period1Month)
	s.Require().NoError(err)

	// Expected totals:
	// Total: 128 + 110 + 50 = 288 USD
	// Paid: 128 USD
	// Unpaid: 110 USD
	// Overdue: 50 USD
	s.InDelta(288.0, summary.TotalAmount, 0.01)
	s.InDelta(128.0, summary.PaidAmount, 0.01)
	s.InDelta(110.0, summary.UnpaidAmount, 0.01)
	s.InDelta(50.0, summary.OverdueAmount, 0.01)
}

func TestAnalyticsCurrencySuite(t *testing.T) {
	suite.Run(t, new(AnalyticsCurrencyTestSuite))
}
