package services

import (
	"time"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// AnalyticsPeriod represents a time period for analytics
type AnalyticsPeriod string

const (
	Period7Days  AnalyticsPeriod = "7d"
	Period1Month AnalyticsPeriod = "1m"
	Period1Year  AnalyticsPeriod = "1y"
)

// AnalyticsSummary represents aggregated invoice statistics
type AnalyticsSummary struct {
	Period        string    `json:"period"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	TotalAmount   float64   `json:"total_amount"`
	PaidAmount    float64   `json:"paid_amount"`
	UnpaidAmount  float64   `json:"unpaid_amount"`
	OverdueAmount float64   `json:"overdue_amount"`
	InvoiceCount  int64     `json:"invoice_count"`
	PaidCount     int64     `json:"paid_count"`
	UnpaidCount   int64     `json:"unpaid_count"`
	OverdueCount  int64     `json:"overdue_count"`
}

// AnalyticsGroupItem represents a single group's statistics
type AnalyticsGroupItem struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Color        string  `json:"color,omitempty"`
	TotalAmount  float64 `json:"total_amount"`
	PaidAmount   float64 `json:"paid_amount"`
	UnpaidAmount float64 `json:"unpaid_amount"`
	InvoiceCount int64   `json:"invoice_count"`
}

// AnalyticsByGroup represents grouped analytics
type AnalyticsByGroup struct {
	Period        string               `json:"period"`
	StartDate     time.Time            `json:"start_date"`
	EndDate       time.Time            `json:"end_date"`
	Items         []AnalyticsGroupItem `json:"items"`
	Uncategorized *AnalyticsGroupItem  `json:"uncategorized,omitempty"`
}

// StatisticsPeriod represents a natural time period for statistics
type StatisticsPeriod string

const (
	PeriodLastDay   StatisticsPeriod = "last_day"
	PeriodLastWeek  StatisticsPeriod = "last_week"
	PeriodLastMonth StatisticsPeriod = "last_month"
	PeriodLastYear  StatisticsPeriod = "last_year"
	PeriodCustom    StatisticsPeriod = "custom"
)

// StatisticsGroupBy represents how to group statistics results
type StatisticsGroupBy string

const (
	GroupByNone     StatisticsGroupBy = ""
	GroupByDay      StatisticsGroupBy = "day"
	GroupByWeek     StatisticsGroupBy = "week"
	GroupByMonth    StatisticsGroupBy = "month"
	GroupByCategory StatisticsGroupBy = "category"
	GroupByCompany  StatisticsGroupBy = "company"
	GroupByReceiver StatisticsGroupBy = "receiver"
)

// StatisticsOptions contains filtering and grouping options for statistics
type StatisticsOptions struct {
	Period              StatisticsPeriod
	Days                int
	CategoryID          *uint
	CompanyID           *uint
	ReceiverID          *uint
	Status              *models.InvoiceStatus
	Keyword             string
	GroupBy             StatisticsGroupBy
	IncludeAggregations bool
}

// StatusStats represents count and amount for a status
type StatusStats struct {
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
}

// StatusBreakdown represents breakdown by invoice status
type StatusBreakdown struct {
	Paid    StatusStats `json:"paid"`
	Unpaid  StatusStats `json:"unpaid"`
	Overdue StatusStats `json:"overdue"`
}

// BreakdownItem represents a single item in the breakdown
type BreakdownItem struct {
	// For time-based grouping
	Date string `json:"date,omitempty"`

	// For entity-based grouping
	ID   uint   `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
}

// InvoiceReference represents a reference to an invoice
type InvoiceReference struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

// DayReference represents a reference to a specific day
type DayReference struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

// EntityReference represents a reference to an entity (category, company, receiver)
type EntityReference struct {
	ID     uint    `json:"id"`
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

// AggregationStats represents aggregation statistics
type AggregationStats struct {
	MaxAmount   float64           `json:"max_amount"`
	MinAmount   float64           `json:"min_amount"`
	AvgAmount   float64           `json:"avg_amount"`
	MaxInvoice  *InvoiceReference `json:"max_invoice,omitempty"`
	MaxDay      *DayReference     `json:"max_day,omitempty"`
	MaxCategory *EntityReference  `json:"max_category,omitempty"`
	MaxCompany  *EntityReference  `json:"max_company,omitempty"`
}

// StatisticsFilters represents the applied filters
type StatisticsFilters struct {
	CategoryID *uint                 `json:"category_id,omitempty"`
	CompanyID  *uint                 `json:"company_id,omitempty"`
	ReceiverID *uint                 `json:"receiver_id,omitempty"`
	Status     *models.InvoiceStatus `json:"status,omitempty"`
	Keyword    string                `json:"keyword,omitempty"`
}

// InvoiceStatistics represents aggregated invoice statistics with optional grouping
type InvoiceStatistics struct {
	Period       string            `json:"period"`
	StartDate    time.Time         `json:"start_date"`
	EndDate      time.Time         `json:"end_date"`
	TotalAmount  float64           `json:"total_amount"`
	InvoiceCount int64             `json:"invoice_count"`
	ByStatus     *StatusBreakdown  `json:"by_status,omitempty"`
	Breakdown    []BreakdownItem   `json:"breakdown,omitempty"`
	Aggregations *AggregationStats `json:"aggregations,omitempty"`
	Filters      StatisticsFilters `json:"filters"`
}

// AnalyticsService handles analytics business logic
type AnalyticsService interface {
	GetSummary(userID string, period AnalyticsPeriod) (*AnalyticsSummary, error)
	GetByCategory(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error)
	GetByCompany(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error)
	GetByReceiver(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error)
	GetStatistics(userID string, opts StatisticsOptions) (*InvoiceStatistics, error)
}

type analyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService creates a new AnalyticsService instance
func NewAnalyticsService(db *gorm.DB) AnalyticsService {
	return &analyticsService{db: db}
}

// getDateRange returns start and end dates for a period
func (s *analyticsService) getDateRange(period AnalyticsPeriod) (time.Time, time.Time) {
	now := time.Now()
	end := now
	var start time.Time

	switch period {
	case Period7Days:
		start = now.AddDate(0, 0, -7)
	case Period1Month:
		start = now.AddDate(0, -1, 0)
	case Period1Year:
		start = now.AddDate(-1, 0, 0)
	default:
		start = now.AddDate(0, -1, 0)
	}

	return start, end
}

// GetSummary returns aggregated invoice statistics for a period
func (s *analyticsService) GetSummary(userID string, period AnalyticsPeriod) (*AnalyticsSummary, error) {
	start, end := s.getDateRange(period)

	summary := &AnalyticsSummary{
		Period:    string(period),
		StartDate: start,
		EndDate:   end,
	}

	// Base query for invoices in the period
	baseQuery := s.db.Model(&models.Invoice{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, start, end)

	// Get total count and amount
	var result struct {
		Count  int64
		Amount float64
	}
	if err := baseQuery.Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as amount").Scan(&result).Error; err != nil {
		return nil, err
	}
	summary.InvoiceCount = result.Count
	summary.TotalAmount = result.Amount

	// Get paid count and amount
	if err := s.db.Model(&models.Invoice{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND status = ?",
			userID, start, end, models.InvoiceStatusPaid).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as amount").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	summary.PaidCount = result.Count
	summary.PaidAmount = result.Amount

	// Get unpaid count and amount
	if err := s.db.Model(&models.Invoice{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND status = ?",
			userID, start, end, models.InvoiceStatusUnpaid).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as amount").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	summary.UnpaidCount = result.Count
	summary.UnpaidAmount = result.Amount

	// Get overdue count and amount
	if err := s.db.Model(&models.Invoice{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND status = ?",
			userID, start, end, models.InvoiceStatusOverdue).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as amount").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	summary.OverdueCount = result.Count
	summary.OverdueAmount = result.Amount

	return summary, nil
}

// GetByCategory returns invoice analytics grouped by category
func (s *analyticsService) GetByCategory(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error) {
	start, end := s.getDateRange(period)

	response := &AnalyticsByGroup{
		Period:    string(period),
		StartDate: start,
		EndDate:   end,
		Items:     []AnalyticsGroupItem{},
	}

	// Query for grouped analytics by category
	type groupResult struct {
		ID           uint
		Name         string
		Color        string
		TotalAmount  float64
		PaidAmount   float64
		UnpaidAmount float64
		InvoiceCount int64
	}

	var results []groupResult
	err := s.db.Table("invoices").
		Select(`
			invoice_categories.id,
			invoice_categories.name,
			invoice_categories.color,
			COUNT(invoices.id) as invoice_count,
			COALESCE(SUM(invoices.amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN invoices.status = 'paid' THEN invoices.amount ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN invoices.status IN ('unpaid', 'overdue') THEN invoices.amount ELSE 0 END), 0) as unpaid_amount
		`).
		Joins("INNER JOIN invoice_categories ON invoices.category_id = invoice_categories.id").
		Where("invoices.user_id = ? AND invoices.created_at >= ? AND invoices.created_at <= ? AND invoices.deleted_at IS NULL",
			userID, start, end).
		Group("invoice_categories.id, invoice_categories.name, invoice_categories.color").
		Order("total_amount DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for _, r := range results {
		response.Items = append(response.Items, AnalyticsGroupItem{
			ID:           r.ID,
			Name:         r.Name,
			Color:        r.Color,
			TotalAmount:  r.TotalAmount,
			PaidAmount:   r.PaidAmount,
			UnpaidAmount: r.UnpaidAmount,
			InvoiceCount: r.InvoiceCount,
		})
	}

	// Get uncategorized invoices
	var uncategorized groupResult
	err = s.db.Table("invoices").
		Select(`
			COUNT(id) as invoice_count,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN status IN ('unpaid', 'overdue') THEN amount ELSE 0 END), 0) as unpaid_amount
		`).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND category_id IS NULL AND deleted_at IS NULL",
			userID, start, end).
		Scan(&uncategorized).Error

	if err != nil {
		return nil, err
	}

	if uncategorized.InvoiceCount > 0 {
		response.Uncategorized = &AnalyticsGroupItem{
			ID:           0,
			Name:         "Uncategorized",
			TotalAmount:  uncategorized.TotalAmount,
			PaidAmount:   uncategorized.PaidAmount,
			UnpaidAmount: uncategorized.UnpaidAmount,
			InvoiceCount: uncategorized.InvoiceCount,
		}
	}

	return response, nil
}

// GetByCompany returns invoice analytics grouped by company
func (s *analyticsService) GetByCompany(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error) {
	start, end := s.getDateRange(period)

	response := &AnalyticsByGroup{
		Period:    string(period),
		StartDate: start,
		EndDate:   end,
		Items:     []AnalyticsGroupItem{},
	}

	type groupResult struct {
		ID           uint
		Name         string
		TotalAmount  float64
		PaidAmount   float64
		UnpaidAmount float64
		InvoiceCount int64
	}

	var results []groupResult
	err := s.db.Table("invoices").
		Select(`
			invoice_companies.id,
			invoice_companies.name,
			COUNT(invoices.id) as invoice_count,
			COALESCE(SUM(invoices.amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN invoices.status = 'paid' THEN invoices.amount ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN invoices.status IN ('unpaid', 'overdue') THEN invoices.amount ELSE 0 END), 0) as unpaid_amount
		`).
		Joins("INNER JOIN invoice_companies ON invoices.company_id = invoice_companies.id").
		Where("invoices.user_id = ? AND invoices.created_at >= ? AND invoices.created_at <= ? AND invoices.deleted_at IS NULL",
			userID, start, end).
		Group("invoice_companies.id, invoice_companies.name").
		Order("total_amount DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for _, r := range results {
		response.Items = append(response.Items, AnalyticsGroupItem{
			ID:           r.ID,
			Name:         r.Name,
			TotalAmount:  r.TotalAmount,
			PaidAmount:   r.PaidAmount,
			UnpaidAmount: r.UnpaidAmount,
			InvoiceCount: r.InvoiceCount,
		})
	}

	// Get invoices without company
	var uncategorized groupResult
	err = s.db.Table("invoices").
		Select(`
			COUNT(id) as invoice_count,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN status IN ('unpaid', 'overdue') THEN amount ELSE 0 END), 0) as unpaid_amount
		`).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND company_id IS NULL AND deleted_at IS NULL",
			userID, start, end).
		Scan(&uncategorized).Error

	if err != nil {
		return nil, err
	}

	if uncategorized.InvoiceCount > 0 {
		response.Uncategorized = &AnalyticsGroupItem{
			ID:           0,
			Name:         "No Company",
			TotalAmount:  uncategorized.TotalAmount,
			PaidAmount:   uncategorized.PaidAmount,
			UnpaidAmount: uncategorized.UnpaidAmount,
			InvoiceCount: uncategorized.InvoiceCount,
		}
	}

	return response, nil
}

// GetByReceiver returns invoice analytics grouped by receiver
func (s *analyticsService) GetByReceiver(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error) {
	start, end := s.getDateRange(period)

	response := &AnalyticsByGroup{
		Period:    string(period),
		StartDate: start,
		EndDate:   end,
		Items:     []AnalyticsGroupItem{},
	}

	type groupResult struct {
		ID           uint
		Name         string
		TotalAmount  float64
		PaidAmount   float64
		UnpaidAmount float64
		InvoiceCount int64
	}

	var results []groupResult
	err := s.db.Table("invoices").
		Select(`
			invoice_receivers.id,
			invoice_receivers.name,
			COUNT(invoices.id) as invoice_count,
			COALESCE(SUM(invoices.amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN invoices.status = 'paid' THEN invoices.amount ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN invoices.status IN ('unpaid', 'overdue') THEN invoices.amount ELSE 0 END), 0) as unpaid_amount
		`).
		Joins("INNER JOIN invoice_receivers ON invoices.receiver_id = invoice_receivers.id").
		Where("invoices.user_id = ? AND invoices.created_at >= ? AND invoices.created_at <= ? AND invoices.deleted_at IS NULL",
			userID, start, end).
		Group("invoice_receivers.id, invoice_receivers.name").
		Order("total_amount DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for _, r := range results {
		response.Items = append(response.Items, AnalyticsGroupItem{
			ID:           r.ID,
			Name:         r.Name,
			TotalAmount:  r.TotalAmount,
			PaidAmount:   r.PaidAmount,
			UnpaidAmount: r.UnpaidAmount,
			InvoiceCount: r.InvoiceCount,
		})
	}

	// Get invoices without receiver
	var uncategorized groupResult
	err = s.db.Table("invoices").
		Select(`
			COUNT(id) as invoice_count,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN status IN ('unpaid', 'overdue') THEN amount ELSE 0 END), 0) as unpaid_amount
		`).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND receiver_id IS NULL AND deleted_at IS NULL",
			userID, start, end).
		Scan(&uncategorized).Error

	if err != nil {
		return nil, err
	}

	if uncategorized.InvoiceCount > 0 {
		response.Uncategorized = &AnalyticsGroupItem{
			ID:           0,
			Name:         "No Receiver",
			TotalAmount:  uncategorized.TotalAmount,
			PaidAmount:   uncategorized.PaidAmount,
			UnpaidAmount: uncategorized.UnpaidAmount,
			InvoiceCount: uncategorized.InvoiceCount,
		}
	}

	return response, nil
}

// getStatisticsDateRange returns start and end dates for a statistics period
func (s *analyticsService) getStatisticsDateRange(opts StatisticsOptions) (time.Time, time.Time) {
	now := time.Now()
	end := now
	var start time.Time

	switch opts.Period {
	case PeriodLastDay:
		start = now.AddDate(0, 0, -1)
	case PeriodLastWeek:
		start = now.AddDate(0, 0, -7)
	case PeriodLastMonth:
		start = now.AddDate(0, -1, 0)
	case PeriodLastYear:
		start = now.AddDate(-1, 0, 0)
	case PeriodCustom:
		if opts.Days > 0 {
			start = now.AddDate(0, 0, -opts.Days)
		} else {
			start = now.AddDate(0, -1, 0) // Default to last month
		}
	default:
		start = now.AddDate(0, -1, 0) // Default to last month
	}

	return start, end
}

// buildStatisticsQuery builds a filtered query for statistics
func (s *analyticsService) buildStatisticsQuery(userID string, start, end time.Time, opts StatisticsOptions) *gorm.DB {
	query := s.db.Model(&models.Invoice{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL",
			userID, start, end)

	if opts.CategoryID != nil {
		query = query.Where("category_id = ?", *opts.CategoryID)
	}
	if opts.CompanyID != nil {
		query = query.Where("company_id = ?", *opts.CompanyID)
	}
	if opts.ReceiverID != nil {
		query = query.Where("receiver_id = ?", *opts.ReceiverID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("(title LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
	}

	return query
}

// GetStatistics returns aggregated invoice statistics with optional grouping and filters
func (s *analyticsService) GetStatistics(userID string, opts StatisticsOptions) (*InvoiceStatistics, error) {
	start, end := s.getStatisticsDateRange(opts)

	stats := &InvoiceStatistics{
		Period:    string(opts.Period),
		StartDate: start,
		EndDate:   end,
		Filters: StatisticsFilters{
			CategoryID: opts.CategoryID,
			CompanyID:  opts.CompanyID,
			ReceiverID: opts.ReceiverID,
			Status:     opts.Status,
			Keyword:    opts.Keyword,
		},
	}

	// Get total count and amount
	var result struct {
		Count  int64
		Amount float64
	}
	if err := s.buildStatisticsQuery(userID, start, end, opts).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as amount").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	stats.InvoiceCount = result.Count
	stats.TotalAmount = result.Amount

	// Handle grouping
	switch opts.GroupBy {
	case GroupByDay:
		breakdown, err := s.getGroupedByDay(userID, start, end, opts)
		if err != nil {
			return nil, err
		}
		stats.Breakdown = breakdown
	case GroupByWeek:
		breakdown, err := s.getGroupedByWeek(userID, start, end, opts)
		if err != nil {
			return nil, err
		}
		stats.Breakdown = breakdown
	case GroupByMonth:
		breakdown, err := s.getGroupedByMonth(userID, start, end, opts)
		if err != nil {
			return nil, err
		}
		stats.Breakdown = breakdown
	case GroupByCategory:
		breakdown, err := s.getGroupedByCategory(userID, start, end, opts)
		if err != nil {
			return nil, err
		}
		stats.Breakdown = breakdown
	case GroupByCompany:
		breakdown, err := s.getGroupedByCompany(userID, start, end, opts)
		if err != nil {
			return nil, err
		}
		stats.Breakdown = breakdown
	case GroupByReceiver:
		breakdown, err := s.getGroupedByReceiver(userID, start, end, opts)
		if err != nil {
			return nil, err
		}
		stats.Breakdown = breakdown
	default:
		// No grouping - include status breakdown
		byStatus, err := s.getStatusBreakdown(userID, start, end, opts)
		if err != nil {
			return nil, err
		}
		stats.ByStatus = byStatus
	}

	// Include aggregations if requested
	if opts.IncludeAggregations {
		aggs, err := s.getAggregations(userID, start, end, opts)
		if err != nil {
			return nil, err
		}
		stats.Aggregations = aggs
	}

	return stats, nil
}

// getStatusBreakdown returns breakdown by status
func (s *analyticsService) getStatusBreakdown(userID string, start, end time.Time, opts StatisticsOptions) (*StatusBreakdown, error) {
	breakdown := &StatusBreakdown{}

	var result struct {
		Count  int64
		Amount float64
	}

	// Paid
	paidOpts := opts
	paidStatus := models.InvoiceStatusPaid
	paidOpts.Status = &paidStatus
	if err := s.buildStatisticsQuery(userID, start, end, paidOpts).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as amount").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	breakdown.Paid = StatusStats{Count: result.Count, Amount: result.Amount}

	// Unpaid
	unpaidOpts := opts
	unpaidStatus := models.InvoiceStatusUnpaid
	unpaidOpts.Status = &unpaidStatus
	if err := s.buildStatisticsQuery(userID, start, end, unpaidOpts).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as amount").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	breakdown.Unpaid = StatusStats{Count: result.Count, Amount: result.Amount}

	// Overdue
	overdueOpts := opts
	overdueStatus := models.InvoiceStatusOverdue
	overdueOpts.Status = &overdueStatus
	if err := s.buildStatisticsQuery(userID, start, end, overdueOpts).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as amount").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	breakdown.Overdue = StatusStats{Count: result.Count, Amount: result.Amount}

	return breakdown, nil
}

// getGroupedByDay returns statistics grouped by day
func (s *analyticsService) getGroupedByDay(userID string, start, end time.Time, opts StatisticsOptions) ([]BreakdownItem, error) {
	type dayResult struct {
		Date   string
		Amount float64
		Count  int64
	}

	var results []dayResult

	query := s.db.Table("invoices").
		Select("DATE(created_at) as date, COALESCE(SUM(amount), 0) as amount, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", userID, start, end)

	if opts.CategoryID != nil {
		query = query.Where("category_id = ?", *opts.CategoryID)
	}
	if opts.CompanyID != nil {
		query = query.Where("company_id = ?", *opts.CompanyID)
	}
	if opts.ReceiverID != nil {
		query = query.Where("receiver_id = ?", *opts.ReceiverID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("(title LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
	}

	if err := query.Group("DATE(created_at)").Order("date ASC").Scan(&results).Error; err != nil {
		return nil, err
	}

	// Create a map of existing dates
	dateMap := make(map[string]dayResult)
	for _, r := range results {
		dateMap[r.Date] = r
	}

	// Fill in all days in the range
	var breakdown []BreakdownItem
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if r, ok := dateMap[dateStr]; ok {
			breakdown = append(breakdown, BreakdownItem{
				Date:   r.Date,
				Amount: r.Amount,
				Count:  r.Count,
			})
		} else {
			breakdown = append(breakdown, BreakdownItem{
				Date:   dateStr,
				Amount: 0,
				Count:  0,
			})
		}
	}

	return breakdown, nil
}

// getGroupedByWeek returns statistics grouped by week
func (s *analyticsService) getGroupedByWeek(userID string, start, end time.Time, opts StatisticsOptions) ([]BreakdownItem, error) {
	type weekResult struct {
		Date   string
		Amount float64
		Count  int64
	}

	var results []weekResult

	// Use strftime to get week start (Monday)
	query := s.db.Table("invoices").
		Select("strftime('%Y-%W', created_at) as date, COALESCE(SUM(amount), 0) as amount, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", userID, start, end)

	if opts.CategoryID != nil {
		query = query.Where("category_id = ?", *opts.CategoryID)
	}
	if opts.CompanyID != nil {
		query = query.Where("company_id = ?", *opts.CompanyID)
	}
	if opts.ReceiverID != nil {
		query = query.Where("receiver_id = ?", *opts.ReceiverID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("(title LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
	}

	if err := query.Group("strftime('%Y-%W', created_at)").Order("date ASC").Scan(&results).Error; err != nil {
		return nil, err
	}

	var breakdown []BreakdownItem
	for _, r := range results {
		breakdown = append(breakdown, BreakdownItem{
			Date:   r.Date,
			Amount: r.Amount,
			Count:  r.Count,
		})
	}

	return breakdown, nil
}

// getGroupedByMonth returns statistics grouped by month
func (s *analyticsService) getGroupedByMonth(userID string, start, end time.Time, opts StatisticsOptions) ([]BreakdownItem, error) {
	type monthResult struct {
		Date   string
		Amount float64
		Count  int64
	}

	var results []monthResult

	query := s.db.Table("invoices").
		Select("strftime('%Y-%m', created_at) as date, COALESCE(SUM(amount), 0) as amount, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", userID, start, end)

	if opts.CategoryID != nil {
		query = query.Where("category_id = ?", *opts.CategoryID)
	}
	if opts.CompanyID != nil {
		query = query.Where("company_id = ?", *opts.CompanyID)
	}
	if opts.ReceiverID != nil {
		query = query.Where("receiver_id = ?", *opts.ReceiverID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("(title LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
	}

	if err := query.Group("strftime('%Y-%m', created_at)").Order("date ASC").Scan(&results).Error; err != nil {
		return nil, err
	}

	var breakdown []BreakdownItem
	for _, r := range results {
		breakdown = append(breakdown, BreakdownItem{
			Date:   r.Date,
			Amount: r.Amount,
			Count:  r.Count,
		})
	}

	return breakdown, nil
}

// getGroupedByCategory returns statistics grouped by category
func (s *analyticsService) getGroupedByCategory(userID string, start, end time.Time, opts StatisticsOptions) ([]BreakdownItem, error) {
	type categoryResult struct {
		ID     uint
		Name   string
		Amount float64
		Count  int64
	}

	var results []categoryResult

	query := s.db.Table("invoices").
		Select("invoice_categories.id, invoice_categories.name, COALESCE(SUM(invoices.amount), 0) as amount, COUNT(invoices.id) as count").
		Joins("LEFT JOIN invoice_categories ON invoices.category_id = invoice_categories.id").
		Where("invoices.user_id = ? AND invoices.created_at >= ? AND invoices.created_at <= ? AND invoices.deleted_at IS NULL", userID, start, end)

	if opts.CategoryID != nil {
		query = query.Where("invoices.category_id = ?", *opts.CategoryID)
	}
	if opts.CompanyID != nil {
		query = query.Where("invoices.company_id = ?", *opts.CompanyID)
	}
	if opts.ReceiverID != nil {
		query = query.Where("invoices.receiver_id = ?", *opts.ReceiverID)
	}
	if opts.Status != nil {
		query = query.Where("invoices.status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("(invoices.title LIKE ? OR invoices.description LIKE ?)", searchPattern, searchPattern)
	}

	if err := query.Group("invoice_categories.id, invoice_categories.name").Order("amount DESC").Scan(&results).Error; err != nil {
		return nil, err
	}

	var breakdown []BreakdownItem
	for _, r := range results {
		name := r.Name
		if name == "" {
			name = "Uncategorized"
		}
		breakdown = append(breakdown, BreakdownItem{
			ID:     r.ID,
			Name:   name,
			Amount: r.Amount,
			Count:  r.Count,
		})
	}

	return breakdown, nil
}

// getGroupedByCompany returns statistics grouped by company
func (s *analyticsService) getGroupedByCompany(userID string, start, end time.Time, opts StatisticsOptions) ([]BreakdownItem, error) {
	type companyResult struct {
		ID     uint
		Name   string
		Amount float64
		Count  int64
	}

	var results []companyResult

	query := s.db.Table("invoices").
		Select("invoice_companies.id, invoice_companies.name, COALESCE(SUM(invoices.amount), 0) as amount, COUNT(invoices.id) as count").
		Joins("LEFT JOIN invoice_companies ON invoices.company_id = invoice_companies.id").
		Where("invoices.user_id = ? AND invoices.created_at >= ? AND invoices.created_at <= ? AND invoices.deleted_at IS NULL", userID, start, end)

	if opts.CategoryID != nil {
		query = query.Where("invoices.category_id = ?", *opts.CategoryID)
	}
	if opts.CompanyID != nil {
		query = query.Where("invoices.company_id = ?", *opts.CompanyID)
	}
	if opts.ReceiverID != nil {
		query = query.Where("invoices.receiver_id = ?", *opts.ReceiverID)
	}
	if opts.Status != nil {
		query = query.Where("invoices.status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("(invoices.title LIKE ? OR invoices.description LIKE ?)", searchPattern, searchPattern)
	}

	if err := query.Group("invoice_companies.id, invoice_companies.name").Order("amount DESC").Scan(&results).Error; err != nil {
		return nil, err
	}

	var breakdown []BreakdownItem
	for _, r := range results {
		name := r.Name
		if name == "" {
			name = "No Company"
		}
		breakdown = append(breakdown, BreakdownItem{
			ID:     r.ID,
			Name:   name,
			Amount: r.Amount,
			Count:  r.Count,
		})
	}

	return breakdown, nil
}

// getGroupedByReceiver returns statistics grouped by receiver
func (s *analyticsService) getGroupedByReceiver(userID string, start, end time.Time, opts StatisticsOptions) ([]BreakdownItem, error) {
	type receiverResult struct {
		ID     uint
		Name   string
		Amount float64
		Count  int64
	}

	var results []receiverResult

	query := s.db.Table("invoices").
		Select("invoice_receivers.id, invoice_receivers.name, COALESCE(SUM(invoices.amount), 0) as amount, COUNT(invoices.id) as count").
		Joins("LEFT JOIN invoice_receivers ON invoices.receiver_id = invoice_receivers.id").
		Where("invoices.user_id = ? AND invoices.created_at >= ? AND invoices.created_at <= ? AND invoices.deleted_at IS NULL", userID, start, end)

	if opts.CategoryID != nil {
		query = query.Where("invoices.category_id = ?", *opts.CategoryID)
	}
	if opts.CompanyID != nil {
		query = query.Where("invoices.company_id = ?", *opts.CompanyID)
	}
	if opts.ReceiverID != nil {
		query = query.Where("invoices.receiver_id = ?", *opts.ReceiverID)
	}
	if opts.Status != nil {
		query = query.Where("invoices.status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("(invoices.title LIKE ? OR invoices.description LIKE ?)", searchPattern, searchPattern)
	}

	if err := query.Group("invoice_receivers.id, invoice_receivers.name").Order("amount DESC").Scan(&results).Error; err != nil {
		return nil, err
	}

	var breakdown []BreakdownItem
	for _, r := range results {
		name := r.Name
		if name == "" {
			name = "No Receiver"
		}
		breakdown = append(breakdown, BreakdownItem{
			ID:     r.ID,
			Name:   name,
			Amount: r.Amount,
			Count:  r.Count,
		})
	}

	return breakdown, nil
}

// getAggregations returns aggregation statistics
func (s *analyticsService) getAggregations(userID string, start, end time.Time, opts StatisticsOptions) (*AggregationStats, error) {
	aggs := &AggregationStats{}

	var result struct {
		MaxAmount float64
		MinAmount float64
		AvgAmount float64
	}

	if err := s.buildStatisticsQuery(userID, start, end, opts).
		Select("COALESCE(MAX(amount), 0) as max_amount, COALESCE(MIN(amount), 0) as min_amount, COALESCE(AVG(amount), 0) as avg_amount").
		Scan(&result).Error; err != nil {
		return nil, err
	}

	aggs.MaxAmount = result.MaxAmount
	aggs.MinAmount = result.MinAmount
	aggs.AvgAmount = result.AvgAmount

	// Get max invoice reference
	var maxInvoice models.Invoice
	if err := s.buildStatisticsQuery(userID, start, end, opts).
		Order("amount DESC").
		Limit(1).
		Find(&maxInvoice).Error; err != nil {
		return nil, err
	}
	if maxInvoice.ID != 0 {
		aggs.MaxInvoice = &InvoiceReference{
			ID:    maxInvoice.ID,
			Title: maxInvoice.Title,
		}
	}

	// If grouped by day, find max day
	if opts.GroupBy == GroupByDay {
		type dayResult struct {
			Date   string
			Amount float64
		}
		var maxDay dayResult

		query := s.db.Table("invoices").
			Select("DATE(created_at) as date, COALESCE(SUM(amount), 0) as amount").
			Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", userID, start, end)

		if opts.CategoryID != nil {
			query = query.Where("category_id = ?", *opts.CategoryID)
		}
		if opts.CompanyID != nil {
			query = query.Where("company_id = ?", *opts.CompanyID)
		}
		if opts.ReceiverID != nil {
			query = query.Where("receiver_id = ?", *opts.ReceiverID)
		}
		if opts.Status != nil {
			query = query.Where("status = ?", *opts.Status)
		}
		if opts.Keyword != "" {
			searchPattern := "%" + opts.Keyword + "%"
			query = query.Where("(title LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
		}

		if err := query.Group("DATE(created_at)").Order("amount DESC").Limit(1).Scan(&maxDay).Error; err != nil {
			return nil, err
		}
		if maxDay.Date != "" {
			aggs.MaxDay = &DayReference{
				Date:   maxDay.Date,
				Amount: maxDay.Amount,
			}
		}
	}

	// If grouped by category, find max category
	if opts.GroupBy == GroupByCategory {
		type catResult struct {
			ID     uint
			Name   string
			Amount float64
		}
		var maxCat catResult

		query := s.db.Table("invoices").
			Select("invoice_categories.id, invoice_categories.name, COALESCE(SUM(invoices.amount), 0) as amount").
			Joins("LEFT JOIN invoice_categories ON invoices.category_id = invoice_categories.id").
			Where("invoices.user_id = ? AND invoices.created_at >= ? AND invoices.created_at <= ? AND invoices.deleted_at IS NULL", userID, start, end)

		if opts.CompanyID != nil {
			query = query.Where("invoices.company_id = ?", *opts.CompanyID)
		}
		if opts.ReceiverID != nil {
			query = query.Where("invoices.receiver_id = ?", *opts.ReceiverID)
		}
		if opts.Status != nil {
			query = query.Where("invoices.status = ?", *opts.Status)
		}
		if opts.Keyword != "" {
			searchPattern := "%" + opts.Keyword + "%"
			query = query.Where("(invoices.title LIKE ? OR invoices.description LIKE ?)", searchPattern, searchPattern)
		}

		if err := query.Group("invoice_categories.id, invoice_categories.name").Order("amount DESC").Limit(1).Scan(&maxCat).Error; err != nil {
			return nil, err
		}
		if maxCat.Name != "" {
			aggs.MaxCategory = &EntityReference{
				ID:     maxCat.ID,
				Name:   maxCat.Name,
				Amount: maxCat.Amount,
			}
		}
	}

	// If grouped by company, find max company
	if opts.GroupBy == GroupByCompany {
		type compResult struct {
			ID     uint
			Name   string
			Amount float64
		}
		var maxComp compResult

		query := s.db.Table("invoices").
			Select("invoice_companies.id, invoice_companies.name, COALESCE(SUM(invoices.amount), 0) as amount").
			Joins("LEFT JOIN invoice_companies ON invoices.company_id = invoice_companies.id").
			Where("invoices.user_id = ? AND invoices.created_at >= ? AND invoices.created_at <= ? AND invoices.deleted_at IS NULL", userID, start, end)

		if opts.CategoryID != nil {
			query = query.Where("invoices.category_id = ?", *opts.CategoryID)
		}
		if opts.ReceiverID != nil {
			query = query.Where("invoices.receiver_id = ?", *opts.ReceiverID)
		}
		if opts.Status != nil {
			query = query.Where("invoices.status = ?", *opts.Status)
		}
		if opts.Keyword != "" {
			searchPattern := "%" + opts.Keyword + "%"
			query = query.Where("(invoices.title LIKE ? OR invoices.description LIKE ?)", searchPattern, searchPattern)
		}

		if err := query.Group("invoice_companies.id, invoice_companies.name").Order("amount DESC").Limit(1).Scan(&maxComp).Error; err != nil {
			return nil, err
		}
		if maxComp.Name != "" {
			aggs.MaxCompany = &EntityReference{
				ID:     maxComp.ID,
				Name:   maxComp.Name,
				Amount: maxComp.Amount,
			}
		}
	}

	return aggs, nil
}
