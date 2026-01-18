package services

import (
	"fmt"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// CategoryService handles invoice category business logic
type CategoryService interface {
	CreateCategory(userID string, category *models.InvoiceCategory) error
	GetCategoryByID(userID string, id uint) (*models.InvoiceCategory, error)
	ListCategories(userID string, keyword string, limit, offset int) ([]models.InvoiceCategory, int64, error)
	UpdateCategory(userID string, category *models.InvoiceCategory) error
	DeleteCategory(userID string, id uint) error
	SearchCategories(userID string, query string) ([]models.InvoiceCategory, error)
}

type categoryService struct {
	db *gorm.DB
}

// NewCategoryService creates a new CategoryService instance
func NewCategoryService(db *gorm.DB) CategoryService {
	return &categoryService{db: db}
}

// CreateCategory creates a new invoice category
func (s *categoryService) CreateCategory(userID string, category *models.InvoiceCategory) error {
	category.UserID = userID
	return s.db.Create(category).Error
}

// GetCategoryByID retrieves a category by ID for a specific user
func (s *categoryService) GetCategoryByID(userID string, id uint) (*models.InvoiceCategory, error) {
	var category models.InvoiceCategory
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// ListCategories lists categories with optional keyword search and pagination
func (s *categoryService) ListCategories(userID string, keyword string, limit, offset int) ([]models.InvoiceCategory, int64, error) {
	var categories []models.InvoiceCategory
	var total int64

	query := s.db.Model(&models.InvoiceCategory{}).Where("user_id = ?", userID)

	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
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

	if err := query.Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

// UpdateCategory updates an existing category
func (s *categoryService) UpdateCategory(userID string, category *models.InvoiceCategory) error {
	// Verify ownership
	existing, err := s.GetCategoryByID(userID, category.ID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Update fields
	existing.Name = category.Name
	existing.Description = category.Description
	existing.Color = category.Color

	return s.db.Save(existing).Error
}

// DeleteCategory soft-deletes a category
func (s *categoryService) DeleteCategory(userID string, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.InvoiceCategory{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

// SearchCategories performs a text search on categories
func (s *categoryService) SearchCategories(userID string, query string) ([]models.InvoiceCategory, error) {
	var categories []models.InvoiceCategory
	searchPattern := "%" + query + "%"

	err := s.db.Where("user_id = ? AND (name LIKE ? OR description LIKE ?)",
		userID, searchPattern, searchPattern).
		Order("name ASC").
		Find(&categories).Error

	return categories, err
}
