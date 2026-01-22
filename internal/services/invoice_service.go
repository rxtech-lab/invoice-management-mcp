package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// CreateInvoiceResult contains the result of creating an invoice
type CreateInvoiceResult struct {
	Invoice     *models.Invoice
	IsDuplicate bool
	Message     string
}

// InvoiceListOptions contains options for listing invoices
type InvoiceListOptions struct {
	Keyword    string
	CategoryID *uint
	CompanyID  *uint
	ReceiverID *uint
	Status     *models.InvoiceStatus
	Tags       []string // Deprecated: use TagIDs instead
	TagIDs     []uint   // Filter by tag IDs
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
	CreateInvoice(userID string, invoice *models.Invoice) (*CreateInvoiceResult, error)
	GetInvoiceByID(userID string, id uint) (*models.Invoice, error)
	ListInvoices(userID string, opts InvoiceListOptions) ([]models.Invoice, int64, error)
	UpdateInvoice(userID string, invoice *models.Invoice) error
	DeleteInvoice(userID string, id uint) error
	SearchInvoices(userID string, query string) ([]models.Invoice, error)

	// Invoice Items
	AddInvoiceItem(userID string, invoiceID uint, item *models.InvoiceItem) error
	UpdateInvoiceItem(userID string, itemID uint, item *models.InvoiceItem, targetAmountOverride *float64, forceRecalculate bool) error
	DeleteInvoiceItem(userID string, itemID uint) error
	GetInvoiceItem(userID string, itemID uint) (*models.InvoiceItem, error)

	// Status management
	UpdateInvoiceStatus(userID string, id uint, status models.InvoiceStatus) error
	GetOverdueInvoices(userID string) ([]models.Invoice, error)

	// Tag management
	SetInvoiceTags(userID string, invoiceID uint, tagNames []string) error
	SetInvoiceTagsByID(userID string, invoiceID uint, tagIDs []int) error
}

type invoiceService struct {
	db        *gorm.DB
	fxService FXService
}

// NewInvoiceService creates a new InvoiceService instance
// fxService can be nil (currency conversion will default to 1:1)
func NewInvoiceService(db *gorm.DB, fxService FXService) InvoiceService {
	return &invoiceService{db: db, fxService: fxService}
}

// CreateInvoice creates a new invoice with optional items
// Amount is always calculated from items (0 if no items)
// Returns existing invoice if a duplicate is found (same amount, dates, and receiver)
func (s *invoiceService) CreateInvoice(userID string, invoice *models.Invoice) (*CreateInvoiceResult, error) {
	invoice.UserID = userID

	// Calculate item amounts, target amounts, and totals
	var totalAmount float64
	var totalTargetAmount float64
	for i := range invoice.Items {
		invoice.Items[i].CalculateAmount()
		s.calculateItemTargetAmount(&invoice.Items[i], invoice.Currency)
		totalAmount += invoice.Items[i].Amount
		totalTargetAmount += invoice.Items[i].TargetAmount
	}
	invoice.Amount = totalAmount
	invoice.TargetAmount = totalTargetAmount

	// Check for duplicate invoice
	var existing models.Invoice
	query := s.db.Where("user_id = ? AND amount = ?", userID, totalAmount)

	// Handle nullable dates - null matches null
	if invoice.InvoiceStartedAt != nil {
		query = query.Where("invoice_started_at = ?", *invoice.InvoiceStartedAt)
	} else {
		query = query.Where("invoice_started_at IS NULL")
	}

	if invoice.InvoiceEndedAt != nil {
		query = query.Where("invoice_ended_at = ?", *invoice.InvoiceEndedAt)
	} else {
		query = query.Where("invoice_ended_at IS NULL")
	}

	// Handle nullable receiver_id
	if invoice.ReceiverID != nil {
		query = query.Where("receiver_id = ?", *invoice.ReceiverID)
	} else {
		query = query.Where("receiver_id IS NULL")
	}

	err := query.Preload("Category").Preload("Company").Preload("Receiver").Preload("Items").Preload("Tags").First(&existing).Error
	if err == nil {
		// Duplicate found - return existing invoice
		return &CreateInvoiceResult{
			Invoice:     &existing,
			IsDuplicate: true,
			Message:     "Duplicate invoice found with matching amount, dates, and receiver",
		}, nil
	}

	// No duplicate found, create new invoice
	if err := s.db.Create(invoice).Error; err != nil {
		return nil, err
	}

	return &CreateInvoiceResult{
		Invoice:     invoice,
		IsDuplicate: false,
	}, nil
}

// GetInvoiceByID retrieves an invoice by ID with related data
func (s *invoiceService) GetInvoiceByID(userID string, id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := s.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Category").
		Preload("Company").
		Preload("Receiver").
		Preload("Items").
		Preload("Tags").
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

	// Filter by tag IDs using subquery
	if len(opts.TagIDs) > 0 {
		query = query.Where("id IN (SELECT invoice_id FROM invoice_tag_mappings WHERE invoice_tag_id IN ?)", opts.TagIDs)
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
	query = query.Preload("Category").Preload("Company").Preload("Receiver").Preload("Items").Preload("Tags")

	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// UpdateInvoice updates an existing invoice
// Note: Amount is NOT updated here - it's calculated from items
// If currency changes, all item target_amounts are recalculated
func (s *invoiceService) UpdateInvoice(userID string, invoice *models.Invoice) error {
	// Verify ownership
	existing, err := s.GetInvoiceByID(userID, invoice.ID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	currencyChanged := existing.Currency != invoice.Currency

	// Update fields (amount is NOT updated - it's calculated from items)
	// Tags are updated separately via SetInvoiceTags
	existing.Title = invoice.Title
	existing.Description = invoice.Description
	existing.InvoiceStartedAt = invoice.InvoiceStartedAt
	existing.InvoiceEndedAt = invoice.InvoiceEndedAt
	existing.ReceiverID = invoice.ReceiverID
	existing.Currency = invoice.Currency
	existing.CategoryID = invoice.CategoryID
	existing.CompanyID = invoice.CompanyID
	existing.OriginalDownloadLink = invoice.OriginalDownloadLink
	existing.Status = invoice.Status
	existing.DueDate = invoice.DueDate

	// If currency changed, recalculate all item target_amounts
	if currencyChanged {
		return s.db.Transaction(func(tx *gorm.DB) error {
			// Save invoice first
			if err := tx.Save(existing).Error; err != nil {
				return err
			}
			// Recalculate all item FX and update invoice total
			return s.recalculateAllItemFX(tx, existing.ID, existing.Currency)
		})
	}

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
		Preload("Tags").
		Order("created_at DESC").
		Find(&invoices).Error

	return invoices, err
}

// AddInvoiceItem adds an item to an invoice
func (s *invoiceService) AddInvoiceItem(userID string, invoiceID uint, item *models.InvoiceItem) error {
	// Verify invoice ownership and get currency
	invoice, err := s.GetInvoiceByID(userID, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	item.InvoiceID = invoiceID
	item.CalculateAmount()
	s.calculateItemTargetAmount(item, invoice.Currency)

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
// targetAmountOverride allows manual override of the USD amount (nil = preserve existing)
// forceRecalculate forces recalculation of target_amount using latest FX rate
func (s *invoiceService) UpdateInvoiceItem(userID string, itemID uint, item *models.InvoiceItem, targetAmountOverride *float64, forceRecalculate bool) error {
	// Get existing item and verify ownership
	existing, err := s.GetInvoiceItem(userID, itemID)
	if err != nil {
		return err
	}

	// Get invoice to access currency
	var invoice models.Invoice
	if err := s.db.First(&invoice, existing.InvoiceID).Error; err != nil {
		return err
	}

	// Update fields
	existing.Description = item.Description
	existing.Quantity = item.Quantity
	existing.UnitPrice = item.UnitPrice
	existing.CalculateAmount()

	// Handle target_amount: forceRecalculate takes precedence, then override, then preserve existing
	if forceRecalculate {
		// Force recalculation using latest FX rate, ignoring any override
		s.calculateItemTargetAmount(existing, invoice.Currency)
	} else if targetAmountOverride != nil {
		// Manual override - use the provided value
		existing.TargetCurrency = "USD"
		existing.TargetAmount = *targetAmountOverride
		// Calculate the implied FX rate from the override
		if existing.Amount > 0 {
			existing.FXRateUsed = *targetAmountOverride / existing.Amount
		} else {
			existing.FXRateUsed = 1.0
		}
	}
	// else: preserve existing target_amount, target_currency, and fx_rate_used

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
		Preload("Tags").
		Order("due_date ASC").
		Find(&invoices).Error

	return invoices, err
}

// updateInvoiceTotal recalculates and updates the invoice total from items
// Sums both amount and target_amount from items
func (s *invoiceService) updateInvoiceTotal(tx *gorm.DB, invoiceID uint) error {
	var result struct {
		TotalAmount       float64
		TotalTargetAmount float64
	}
	err := tx.Model(&models.InvoiceItem{}).
		Where("invoice_id = ?", invoiceID).
		Select("COALESCE(SUM(amount), 0) as total_amount, COALESCE(SUM(target_amount), 0) as total_target_amount").
		Scan(&result).Error
	if err != nil {
		return err
	}

	// Save updates
	return tx.Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Updates(map[string]interface{}{
			"amount":        result.TotalAmount,
			"target_amount": result.TotalTargetAmount,
		}).Error
}

// recalculateAllItemFX recalculates FX for all items when currency changes
func (s *invoiceService) recalculateAllItemFX(tx *gorm.DB, invoiceID uint, currency string) error {
	// Get all items for this invoice
	var items []models.InvoiceItem
	if err := tx.Where("invoice_id = ?", invoiceID).Find(&items).Error; err != nil {
		return err
	}

	// Recalculate FX for each item
	var totalTargetAmount float64
	for i := range items {
		s.calculateItemTargetAmount(&items[i], currency)
		totalTargetAmount += items[i].TargetAmount

		// Update item
		if err := tx.Model(&models.InvoiceItem{}).
			Where("id = ?", items[i].ID).
			Updates(map[string]interface{}{
				"target_currency": items[i].TargetCurrency,
				"target_amount":   items[i].TargetAmount,
				"fx_rate_used":    items[i].FXRateUsed,
			}).Error; err != nil {
			return err
		}
	}

	// Update invoice target_amount
	return tx.Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Update("target_amount", totalTargetAmount).Error
}

// SetInvoiceTags sets the tags for an invoice by tag names
// It will look up existing tags or create new ones as needed
func (s *invoiceService) SetInvoiceTags(userID string, invoiceID uint, tagNames []string) error {
	// Verify invoice ownership
	_, err := s.GetInvoiceByID(userID, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing tag mappings for this invoice
		if err := tx.Where("invoice_id = ?", invoiceID).Delete(&models.InvoiceTagMapping{}).Error; err != nil {
			return err
		}

		if len(tagNames) == 0 {
			return nil
		}

		// For each tag name, find or create the tag and create the mapping
		for _, tagName := range tagNames {
			if tagName == "" {
				continue
			}

			// Try to find existing tag
			var tag models.InvoiceTag
			err := tx.Where("user_id = ? AND name = ?", userID, tagName).First(&tag).Error
			if err != nil {
				// Create new tag
				tag = models.InvoiceTag{
					UserID: userID,
					Name:   tagName,
					Color:  "#6B7280", // Default gray color
				}
				if err := tx.Create(&tag).Error; err != nil {
					return err
				}
			}

			// Create mapping
			mapping := models.InvoiceTagMapping{
				InvoiceID: invoiceID,
				TagID:     tag.ID,
			}
			if err := tx.Create(&mapping).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// SetInvoiceTagsByID sets tags for an invoice using tag IDs
func (s *invoiceService) SetInvoiceTagsByID(userID string, invoiceID uint, tagIDs []int) error {
	// Verify invoice ownership
	_, err := s.GetInvoiceByID(userID, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing tag mappings for this invoice
		if err := tx.Where("invoice_id = ?", invoiceID).Delete(&models.InvoiceTagMapping{}).Error; err != nil {
			return err
		}

		if len(tagIDs) == 0 {
			return nil
		}

		// For each tag ID, verify it exists and belongs to user, then create mapping
		for _, tagID := range tagIDs {
			var tag models.InvoiceTag
			err := tx.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error
			if err != nil {
				return fmt.Errorf("tag with ID %d not found or not owned by user", tagID)
			}

			// Create mapping
			mapping := models.InvoiceTagMapping{
				InvoiceID: invoiceID,
				TagID:     tag.ID,
			}
			if err := tx.Create(&mapping).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// calculateItemTargetAmount calculates and sets the target amount (USD) for an invoice item
func (s *invoiceService) calculateItemTargetAmount(item *models.InvoiceItem, invoiceCurrency string) {
	// Default target currency to USD
	item.TargetCurrency = "USD"

	// If no FX service, use 1:1 rate
	if s.fxService == nil {
		item.TargetAmount = item.Amount
		item.FXRateUsed = 1.0
		return
	}

	// If same currency, no conversion needed
	if invoiceCurrency == "USD" {
		item.TargetAmount = item.Amount
		item.FXRateUsed = 1.0
		return
	}

	// Convert to USD
	ctx := context.Background()
	convertedAmount, rate, _ := s.fxService.ConvertAmount(ctx, item.Amount, invoiceCurrency, "USD")
	item.TargetAmount = convertedAmount
	item.FXRateUsed = rate
}
