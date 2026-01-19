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

// AnalyticsService handles analytics business logic
type AnalyticsService interface {
	GetSummary(userID string, period AnalyticsPeriod) (*AnalyticsSummary, error)
	GetByCategory(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error)
	GetByCompany(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error)
	GetByReceiver(userID string, period AnalyticsPeriod) (*AnalyticsByGroup, error)
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
