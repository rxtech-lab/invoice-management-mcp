package services

import (
	"fmt"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// TagService handles invoice tag business logic
type TagService interface {
	// Tag CRUD
	CreateTag(userID string, tag *models.InvoiceTag) error
	GetTagByID(userID string, id uint) (*models.InvoiceTag, error)
	GetTagByName(userID string, name string) (*models.InvoiceTag, error)
	ListTags(userID string, keyword string, limit, offset int) ([]models.InvoiceTag, int64, error)
	UpdateTag(userID string, tag *models.InvoiceTag) error
	DeleteTag(userID string, id uint) error
	SearchTags(userID string, query string) ([]models.InvoiceTag, error)

	// Tag-Invoice relationships
	AddTagToInvoice(userID string, invoiceID, tagID uint) error
	RemoveTagFromInvoice(userID string, invoiceID, tagID uint) error
	GetInvoicesByTagID(userID string, tagID uint, limit, offset int) ([]models.Invoice, int64, error)
	GetOrCreateTagByName(userID string, name string) (*models.InvoiceTag, error)
}

type tagService struct {
	db *gorm.DB
}

// NewTagService creates a new TagService instance
func NewTagService(db *gorm.DB) TagService {
	return &tagService{db: db}
}

// CreateTag creates a new tag
func (s *tagService) CreateTag(userID string, tag *models.InvoiceTag) error {
	tag.UserID = userID

	// Check if tag with same name already exists for this user
	var existing models.InvoiceTag
	err := s.db.Where("user_id = ? AND name = ?", userID, tag.Name).First(&existing).Error
	if err == nil {
		return fmt.Errorf("tag with name '%s' already exists", tag.Name)
	}

	return s.db.Create(tag).Error
}

// GetTagByID retrieves a tag by ID for a specific user
func (s *tagService) GetTagByID(userID string, id uint) (*models.InvoiceTag, error) {
	var tag models.InvoiceTag
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetTagByName retrieves a tag by name for a specific user
func (s *tagService) GetTagByName(userID string, name string) (*models.InvoiceTag, error) {
	var tag models.InvoiceTag
	err := s.db.Where("user_id = ? AND name = ?", userID, name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// ListTags lists tags with optional keyword search and pagination
func (s *tagService) ListTags(userID string, keyword string, limit, offset int) ([]models.InvoiceTag, int64, error) {
	var tags []models.InvoiceTag
	var total int64

	query := s.db.Model(&models.InvoiceTag{}).Where("user_id = ?", userID)

	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where("name LIKE ?", searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Order by name
	query = query.Order("name ASC")

	if err := query.Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// UpdateTag updates an existing tag
func (s *tagService) UpdateTag(userID string, tag *models.InvoiceTag) error {
	// Verify ownership
	existing, err := s.GetTagByID(userID, tag.ID)
	if err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	// Check if new name conflicts with another tag
	if tag.Name != existing.Name {
		var conflict models.InvoiceTag
		err := s.db.Where("user_id = ? AND name = ? AND id != ?", userID, tag.Name, tag.ID).First(&conflict).Error
		if err == nil {
			return fmt.Errorf("tag with name '%s' already exists", tag.Name)
		}
	}

	// Update fields
	existing.Name = tag.Name
	existing.Color = tag.Color

	return s.db.Save(existing).Error
}

// DeleteTag soft-deletes a tag and removes all its mappings
func (s *tagService) DeleteTag(userID string, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify ownership
		var tag models.InvoiceTag
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&tag).Error; err != nil {
			return fmt.Errorf("tag not found: %w", err)
		}

		// Delete all mappings for this tag
		if err := tx.Where("invoice_tag_id = ?", id).Delete(&models.InvoiceTagMapping{}).Error; err != nil {
			return err
		}

		// Delete the tag
		return tx.Delete(&tag).Error
	})
}

// SearchTags performs a text search on tags
func (s *tagService) SearchTags(userID string, query string) ([]models.InvoiceTag, error) {
	var tags []models.InvoiceTag
	searchPattern := "%" + query + "%"

	err := s.db.Where("user_id = ? AND name LIKE ?", userID, searchPattern).
		Order("name ASC").
		Find(&tags).Error

	return tags, err
}

// AddTagToInvoice adds a tag to an invoice
func (s *tagService) AddTagToInvoice(userID string, invoiceID, tagID uint) error {
	// Verify invoice ownership
	var invoice models.Invoice
	if err := s.db.Where("id = ? AND user_id = ?", invoiceID, userID).First(&invoice).Error; err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Verify tag ownership
	var tag models.InvoiceTag
	if err := s.db.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	// Check if mapping already exists
	var existing models.InvoiceTagMapping
	err := s.db.Where("invoice_id = ? AND invoice_tag_id = ?", invoiceID, tagID).First(&existing).Error
	if err == nil {
		// Mapping already exists, nothing to do
		return nil
	}

	// Create mapping
	mapping := models.InvoiceTagMapping{
		InvoiceID: invoiceID,
		TagID:     tagID,
	}
	return s.db.Create(&mapping).Error
}

// RemoveTagFromInvoice removes a tag from an invoice
func (s *tagService) RemoveTagFromInvoice(userID string, invoiceID, tagID uint) error {
	// Verify invoice ownership
	var invoice models.Invoice
	if err := s.db.Where("id = ? AND user_id = ?", invoiceID, userID).First(&invoice).Error; err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Verify tag ownership
	var tag models.InvoiceTag
	if err := s.db.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	// Delete the mapping
	result := s.db.Where("invoice_id = ? AND invoice_tag_id = ?", invoiceID, tagID).Delete(&models.InvoiceTagMapping{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetInvoicesByTagID retrieves invoices that have a specific tag
func (s *tagService) GetInvoicesByTagID(userID string, tagID uint, limit, offset int) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	// Verify tag ownership
	var tag models.InvoiceTag
	if err := s.db.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		return nil, 0, fmt.Errorf("tag not found: %w", err)
	}

	// Build query for invoices with this tag
	query := s.db.Model(&models.Invoice{}).
		Where("user_id = ?", userID).
		Where("id IN (SELECT invoice_id FROM invoice_tag_mappings WHERE invoice_tag_id = ?)", tagID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Preload relationships
	query = query.Preload("Category").
		Preload("Company").
		Preload("Receiver").
		Preload("Items").
		Preload("Tags").
		Order("created_at DESC")

	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// GetOrCreateTagByName gets an existing tag by name or creates a new one
func (s *tagService) GetOrCreateTagByName(userID string, name string) (*models.InvoiceTag, error) {
	var tag models.InvoiceTag
	err := s.db.Where("user_id = ? AND name = ?", userID, name).First(&tag).Error
	if err == nil {
		return &tag, nil
	}

	// Create new tag
	tag = models.InvoiceTag{
		UserID: userID,
		Name:   name,
		Color:  "#6B7280", // Default gray color
	}
	if err := s.db.Create(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}
