package services

import (
	"fmt"
	"time"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// InvoiceListOptions contains options for listing invoices
type InvoiceListOptions struct {
	Keyword    string
	CategoryID *uint
	CompanyID  *uint
	ReceiverID *uint
	Status     *models.InvoiceStatus
	Tags       []string
	StartDate  *time.Time
	EndDate    *time.Time
	SortBy     string // "created_at", "amount", "due_date", "title"
	SortOrder  string // "asc", "desc"
	Limit      int
	Offset     int
}

// InvoiceService handles invoice business logic
type InvoiceService interface {
	// Invoice CRUD
	CreateInvoice(userID string, invoice *models.Invoice) error
	GetInvoiceByID(userID string, id uint) (*models.Invoice, error)
	ListInvoices(userID string, opts InvoiceListOptions) ([]models.Invoice, int64, error)
	UpdateInvoice(userID string, invoice *models.Invoice) error
	DeleteInvoice(userID string, id uint) error
	SearchInvoices(userID string, query string) ([]models.Invoice, error)

	// Invoice Items
	AddInvoiceItem(userID string, invoiceID uint, item *models.InvoiceItem) error
	UpdateInvoiceItem(userID string, itemID uint, item *models.InvoiceItem) error
	DeleteInvoiceItem(userID string, itemID uint) error
	GetInvoiceItem(userID string, itemID uint) (*models.InvoiceItem, error)

	// Status management
	UpdateInvoiceStatus(userID string, id uint, status models.InvoiceStatus) error
	GetOverdueInvoices(userID string) ([]models.Invoice, error)
}

type invoiceService struct {
	db *gorm.DB
}

// NewInvoiceService creates a new InvoiceService instance
func NewInvoiceService(db *gorm.DB) InvoiceService {
	return &invoiceService{db: db}
}

// CreateInvoice creates a new invoice with optional items
// Amount is always calculated from items (0 if no items)
func (s *invoiceService) CreateInvoice(userID string, invoice *models.Invoice) error {
	invoice.UserID = userID

	// Calculate item amounts and total - amount is always calculated from items
	var totalAmount float64
	for i := range invoice.Items {
		invoice.Items[i].CalculateAmount()
		totalAmount += invoice.Items[i].Amount
	}
	invoice.Amount = totalAmount

	return s.db.Create(invoice).Error
}

// GetInvoiceByID retrieves an invoice by ID with related data
func (s *invoiceService) GetInvoiceByID(userID string, id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := s.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Category").
		Preload("Company").
		Preload("Receiver").
		Preload("Items").
		First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// ListInvoices lists invoices with filtering, sorting, and pagination
func (s *invoiceService) ListInvoices(userID string, opts InvoiceListOptions) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := s.db.Model(&models.Invoice{}).Where("user_id = ?", userID)

	// Apply filters
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

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

	if opts.StartDate != nil {
		query = query.Where("created_at >= ?", *opts.StartDate)
	}

	if opts.EndDate != nil {
		query = query.Where("created_at <= ?", *opts.EndDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	if opts.SortBy != "" {
		switch opts.SortBy {
		case "created_at", "amount", "due_date", "title":
			sortBy = opts.SortBy
		}
	}

	sortOrder := "DESC"
	if opts.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	// Preload relationships
	query = query.Preload("Category").Preload("Company").Preload("Receiver").Preload("Items")

	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// UpdateInvoice updates an existing invoice
// Note: Amount is NOT updated here - it's calculated from items
func (s *invoiceService) UpdateInvoice(userID string, invoice *models.Invoice) error {
	// Verify ownership
	existing, err := s.GetInvoiceByID(userID, invoice.ID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Update fields (amount is NOT updated - it's calculated from items)
	existing.Title = invoice.Title
	existing.Description = invoice.Description
	existing.InvoiceStartedAt = invoice.InvoiceStartedAt
	existing.InvoiceEndedAt = invoice.InvoiceEndedAt
	existing.ReceiverID = invoice.ReceiverID
	existing.Currency = invoice.Currency
	existing.CategoryID = invoice.CategoryID
	existing.CompanyID = invoice.CompanyID
	existing.OriginalDownloadLink = invoice.OriginalDownloadLink
	existing.Tags = invoice.Tags
	existing.Status = invoice.Status
	existing.DueDate = invoice.DueDate

	return s.db.Save(existing).Error
}

// DeleteInvoice soft-deletes an invoice and its items
func (s *invoiceService) DeleteInvoice(userID string, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify ownership
		var invoice models.Invoice
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&invoice).Error; err != nil {
			return fmt.Errorf("invoice not found: %w", err)
		}

		// Delete items first
		if err := tx.Where("invoice_id = ?", id).Delete(&models.InvoiceItem{}).Error; err != nil {
			return err
		}

		// Delete invoice
		return tx.Delete(&invoice).Error
	})
}

// SearchInvoices performs a text search on invoices
func (s *invoiceService) SearchInvoices(userID string, query string) ([]models.Invoice, error) {
	var invoices []models.Invoice
	searchPattern := "%" + query + "%"

	err := s.db.Where("user_id = ? AND (title LIKE ? OR description LIKE ?)",
		userID, searchPattern, searchPattern).
		Preload("Category").
		Preload("Company").
		Preload("Receiver").
		Preload("Items").
		Order("created_at DESC").
		Find(&invoices).Error

	return invoices, err
}

// AddInvoiceItem adds an item to an invoice
func (s *invoiceService) AddInvoiceItem(userID string, invoiceID uint, item *models.InvoiceItem) error {
	// Verify invoice ownership
	_, err := s.GetInvoiceByID(userID, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	item.InvoiceID = invoiceID
	item.CalculateAmount()

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create item
		if err := tx.Create(item).Error; err != nil {
			return err
		}

		// Update invoice total
		return s.updateInvoiceTotal(tx, invoiceID)
	})
}

// UpdateInvoiceItem updates an invoice item
func (s *invoiceService) UpdateInvoiceItem(userID string, itemID uint, item *models.InvoiceItem) error {
	// Get existing item and verify ownership
	existing, err := s.GetInvoiceItem(userID, itemID)
	if err != nil {
		return err
	}

	// Update fields
	existing.Description = item.Description
	existing.Quantity = item.Quantity
	existing.UnitPrice = item.UnitPrice
	existing.CalculateAmount()

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(existing).Error; err != nil {
			return err
		}
		return s.updateInvoiceTotal(tx, existing.InvoiceID)
	})
}

// DeleteInvoiceItem deletes an invoice item
func (s *invoiceService) DeleteInvoiceItem(userID string, itemID uint) error {
	// Get existing item and verify ownership
	existing, err := s.GetInvoiceItem(userID, itemID)
	if err != nil {
		return err
	}

	invoiceID := existing.InvoiceID

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(existing).Error; err != nil {
			return err
		}
		return s.updateInvoiceTotal(tx, invoiceID)
	})
}

// GetInvoiceItem retrieves an invoice item and verifies user ownership
func (s *invoiceService) GetInvoiceItem(userID string, itemID uint) (*models.InvoiceItem, error) {
	var item models.InvoiceItem
	err := s.db.First(&item, itemID).Error
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}

	// Verify invoice ownership
	var invoice models.Invoice
	err = s.db.Where("id = ? AND user_id = ?", item.InvoiceID, userID).First(&invoice).Error
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return &item, nil
}

// UpdateInvoiceStatus updates only the status of an invoice
func (s *invoiceService) UpdateInvoiceStatus(userID string, id uint, status models.InvoiceStatus) error {
	result := s.db.Model(&models.Invoice{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}
	return nil
}

// GetOverdueInvoices returns all overdue invoices for a user
func (s *invoiceService) GetOverdueInvoices(userID string) ([]models.Invoice, error) {
	var invoices []models.Invoice
	now := time.Now()

	err := s.db.Where("user_id = ? AND status = ? AND due_date < ?",
		userID, models.InvoiceStatusUnpaid, now).
		Preload("Category").
		Preload("Company").
		Preload("Receiver").
		Preload("Items").
		Order("due_date ASC").
		Find(&invoices).Error

	return invoices, err
}

// updateInvoiceTotal recalculates and updates the invoice total from items
func (s *invoiceService) updateInvoiceTotal(tx *gorm.DB, invoiceID uint) error {
	var total float64
	err := tx.Model(&models.InvoiceItem{}).
		Where("invoice_id = ?", invoiceID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return err
	}

	return tx.Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Update("amount", total).Error
}
