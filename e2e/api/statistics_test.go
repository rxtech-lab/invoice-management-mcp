package api

import (
	"testing"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/stretchr/testify/suite"
)

type StatisticsTestSuite struct {
	suite.Suite
	setup      *TestSetup
	categoryID uint
	companyID  uint
	receiverID uint
}

func (s *StatisticsTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())

	// Create test fixtures
	var err error
	s.categoryID, err = s.setup.CreateTestCategory("Utilities")
	s.Require().NoError(err)
	s.companyID, err = s.setup.CreateTestCompany("Electric Co")
	s.Require().NoError(err)
	s.receiverID, err = s.setup.CreateTestReceiver("John Doe", false)
	s.Require().NoError(err)

	// Create test invoices with varied data
	// Use DaysAgo to create invoices at different dates
	_, err = s.setup.CreateTestInvoiceOnDate("Electricity January", &s.categoryID, &s.companyID, "paid", 150.00, DaysAgo(5))
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceOnDate("Electricity February", &s.categoryID, &s.companyID, "unpaid", 175.00, DaysAgo(3))
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceOnDate("Water Bill", &s.categoryID, &s.companyID, "paid", 50.00, DaysAgo(2))
	s.Require().NoError(err)

	// Create category for services
	servicesID, err := s.setup.CreateTestCategory("Services")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestInvoiceOnDate("Consulting Fee", &servicesID, nil, "overdue", 500.00, DaysAgo(1))
	s.Require().NoError(err)

	_, err = s.setup.CreateTestInvoiceOnDate("Internet Service", &s.categoryID, nil, "paid", 80.00, DaysAgo(0))
	s.Require().NoError(err)
}

func (s *StatisticsTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *StatisticsTestSuite) TestDefaultPeriod() {
	opts := services.StatisticsOptions{
		Period: services.PeriodLastMonth,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.Equal("last_month", stats.Period)
	s.NotNil(stats.ByStatus)
	s.Equal(int64(5), stats.InvoiceCount)
	s.Equal(955.00, stats.TotalAmount)
}

func (s *StatisticsTestSuite) TestLastWeekPeriod() {
	opts := services.StatisticsOptions{
		Period: services.PeriodLastWeek,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.Equal("last_week", stats.Period)
	s.Equal(int64(5), stats.InvoiceCount)
}

func (s *StatisticsTestSuite) TestLastDayPeriod() {
	opts := services.StatisticsOptions{
		Period: services.PeriodLastDay,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.Equal("last_day", stats.Period)
	// Only the invoice from today should be included
	s.GreaterOrEqual(stats.InvoiceCount, int64(1))
}

func (s *StatisticsTestSuite) TestCustomDays() {
	opts := services.StatisticsOptions{
		Period: services.PeriodCustom,
		Days:   3,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.Equal("custom", stats.Period)
	// Invoices from last 3 days
	s.GreaterOrEqual(stats.InvoiceCount, int64(3))
}

func (s *StatisticsTestSuite) TestFilterByCategory() {
	opts := services.StatisticsOptions{
		Period:     services.PeriodLastMonth,
		CategoryID: &s.categoryID,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should only include Utilities category invoices (4 invoices)
	s.Equal(int64(4), stats.InvoiceCount)
	s.Equal(s.categoryID, *stats.Filters.CategoryID)
}

func (s *StatisticsTestSuite) TestFilterByCompany() {
	opts := services.StatisticsOptions{
		Period:    services.PeriodLastMonth,
		CompanyID: &s.companyID,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should only include Electric Co invoices (3 invoices)
	s.Equal(int64(3), stats.InvoiceCount)
	s.Equal(s.companyID, *stats.Filters.CompanyID)
}

func (s *StatisticsTestSuite) TestFilterByStatus() {
	paidStatus := models.InvoiceStatusPaid
	opts := services.StatisticsOptions{
		Period: services.PeriodLastMonth,
		Status: &paidStatus,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should only include paid invoices (3 invoices)
	s.Equal(int64(3), stats.InvoiceCount)
}

func (s *StatisticsTestSuite) TestFilterByKeyword() {
	opts := services.StatisticsOptions{
		Period:  services.PeriodLastMonth,
		Keyword: "Electricity",
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should only include invoices with "Electricity" in title
	s.Equal(int64(2), stats.InvoiceCount)
	s.Equal("Electricity", stats.Filters.Keyword)
}

func (s *StatisticsTestSuite) TestCombinedFilters() {
	opts := services.StatisticsOptions{
		Period:     services.PeriodLastMonth,
		CategoryID: &s.categoryID,
		Keyword:    "Electricity",
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should only include Utilities category invoices with "Electricity" keyword
	s.Equal(int64(2), stats.InvoiceCount)
}

func (s *StatisticsTestSuite) TestStatusBreakdown() {
	opts := services.StatisticsOptions{
		Period: services.PeriodLastMonth,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.NotNil(stats.ByStatus)
	s.Equal(int64(3), stats.ByStatus.Paid.Count)
	s.Equal(280.00, stats.ByStatus.Paid.Amount) // 150 + 50 + 80
	s.Equal(int64(1), stats.ByStatus.Unpaid.Count)
	s.Equal(175.00, stats.ByStatus.Unpaid.Amount)
	s.Equal(int64(1), stats.ByStatus.Overdue.Count)
	s.Equal(500.00, stats.ByStatus.Overdue.Amount)
}

func (s *StatisticsTestSuite) TestGroupByDay() {
	opts := services.StatisticsOptions{
		Period:  services.PeriodLastWeek,
		GroupBy: services.GroupByDay,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should have breakdown by day
	s.NotNil(stats.Breakdown)
	s.GreaterOrEqual(len(stats.Breakdown), 7) // At least 7 days

	// Verify breakdown items have date field
	for _, item := range stats.Breakdown {
		s.NotEmpty(item.Date)
	}
}

func (s *StatisticsTestSuite) TestGroupByMonth() {
	opts := services.StatisticsOptions{
		Period:  services.PeriodLastYear,
		GroupBy: services.GroupByMonth,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should have breakdown by month
	s.NotNil(stats.Breakdown)
	// All our test data is in the same month
	s.GreaterOrEqual(len(stats.Breakdown), 1)
}

func (s *StatisticsTestSuite) TestGroupByCategory() {
	opts := services.StatisticsOptions{
		Period:  services.PeriodLastMonth,
		GroupBy: services.GroupByCategory,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should have breakdown by category
	s.NotNil(stats.Breakdown)
	s.GreaterOrEqual(len(stats.Breakdown), 2) // Utilities and Services

	// Verify breakdown items have name field
	for _, item := range stats.Breakdown {
		s.NotEmpty(item.Name)
	}
}

func (s *StatisticsTestSuite) TestGroupByCompany() {
	opts := services.StatisticsOptions{
		Period:  services.PeriodLastMonth,
		GroupBy: services.GroupByCompany,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should have breakdown by company
	s.NotNil(stats.Breakdown)
	s.GreaterOrEqual(len(stats.Breakdown), 1) // Electric Co and No Company
}

func (s *StatisticsTestSuite) TestAggregations() {
	opts := services.StatisticsOptions{
		Period:              services.PeriodLastMonth,
		IncludeAggregations: true,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.NotNil(stats.Aggregations)
	s.Equal(500.00, stats.Aggregations.MaxAmount) // Consulting Fee
	s.Equal(50.00, stats.Aggregations.MinAmount)  // Water Bill
	s.Greater(stats.Aggregations.AvgAmount, 0.0)
}

func (s *StatisticsTestSuite) TestAggregationsWithMaxInvoice() {
	opts := services.StatisticsOptions{
		Period:              services.PeriodLastMonth,
		IncludeAggregations: true,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.NotNil(stats.Aggregations)
	s.NotNil(stats.Aggregations.MaxInvoice)
	s.Equal("Consulting Fee", stats.Aggregations.MaxInvoice.Title)
}

func (s *StatisticsTestSuite) TestGroupByDayWithAggregations() {
	opts := services.StatisticsOptions{
		Period:              services.PeriodLastWeek,
		GroupBy:             services.GroupByDay,
		IncludeAggregations: true,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.NotNil(stats.Breakdown)
	s.NotNil(stats.Aggregations)
	s.NotNil(stats.Aggregations.MaxDay)
}

func (s *StatisticsTestSuite) TestNoInvoicesInPeriod() {
	opts := services.StatisticsOptions{
		Period: services.PeriodCustom,
		Days:   0, // This should default to last month but we'll use a specific user with no invoices
	}

	// Use a different user ID with no invoices
	stats, err := s.setup.AnalyticsService.GetStatistics("non-existent-user", opts)
	s.Require().NoError(err)

	s.Equal(int64(0), stats.InvoiceCount)
	s.Equal(0.0, stats.TotalAmount)
}

func (s *StatisticsTestSuite) TestUserIsolation() {
	// Create an invoice for a different user
	otherUserID := "other-user-456"

	// Get stats for the test user
	opts := services.StatisticsOptions{
		Period: services.PeriodLastMonth,
	}
	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)
	s.Equal(int64(5), stats.InvoiceCount)

	// Get stats for the other user (should be 0)
	otherStats, err := s.setup.AnalyticsService.GetStatistics(otherUserID, opts)
	s.Require().NoError(err)
	s.Equal(int64(0), otherStats.InvoiceCount)
}

func (s *StatisticsTestSuite) TestElectricityLastMonthQuery() {
	// Test the example query: "search invoices last month for electricity"
	opts := services.StatisticsOptions{
		Period:  services.PeriodLastMonth,
		Keyword: "electricity",
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.Equal(int64(2), stats.InvoiceCount)
	s.Equal(325.00, stats.TotalAmount) // 150 + 175
}

func (s *StatisticsTestSuite) TestBarChartDataForSevenDays() {
	// Test: "I want to see a bar chart showing invoice spending for 7 days"
	opts := services.StatisticsOptions{
		Period:  services.PeriodLastWeek,
		GroupBy: services.GroupByDay,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	// Should have 7+ days of data (including all days even if no invoices)
	s.GreaterOrEqual(len(stats.Breakdown), 7)

	// Each breakdown item should have date and amount for chart
	for _, item := range stats.Breakdown {
		s.NotEmpty(item.Date, "Bar chart needs date field")
		s.GreaterOrEqual(item.Amount, 0.0, "Bar chart needs amount field")
	}
}

func (s *StatisticsTestSuite) TestMaxSpendLastWeek() {
	// Test: "What is the max spend that happened last week?"
	opts := services.StatisticsOptions{
		Period:              services.PeriodLastWeek,
		IncludeAggregations: true,
	}

	stats, err := s.setup.AnalyticsService.GetStatistics(s.setup.TestUserID, opts)
	s.Require().NoError(err)

	s.NotNil(stats.Aggregations)
	s.Equal(500.00, stats.Aggregations.MaxAmount)
	s.NotNil(stats.Aggregations.MaxInvoice)
}

func TestStatisticsSuite(t *testing.T) {
	suite.Run(t, new(StatisticsTestSuite))
}
