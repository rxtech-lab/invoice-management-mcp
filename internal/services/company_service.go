package services

import (
	"fmt"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// CompanyService handles invoice company business logic
type CompanyService interface {
	CreateCompany(userID string, company *models.InvoiceCompany) error
	GetCompanyByID(userID string, id uint) (*models.InvoiceCompany, error)
	ListCompanies(userID string, keyword string, limit, offset int) ([]models.InvoiceCompany, int64, error)
	UpdateCompany(userID string, company *models.InvoiceCompany) error
	DeleteCompany(userID string, id uint) error
	SearchCompanies(userID string, query string) ([]models.InvoiceCompany, error)
}

type companyService struct {
	db *gorm.DB
}

// NewCompanyService creates a new CompanyService instance
func NewCompanyService(db *gorm.DB) CompanyService {
	return &companyService{db: db}
}

// CreateCompany creates a new invoice company
func (s *companyService) CreateCompany(userID string, company *models.InvoiceCompany) error {
	company.UserID = userID
	return s.db.Create(company).Error
}

// GetCompanyByID retrieves a company by ID for a specific user
func (s *companyService) GetCompanyByID(userID string, id uint) (*models.InvoiceCompany, error) {
	var company models.InvoiceCompany
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// ListCompanies lists companies with optional keyword search and pagination
func (s *companyService) ListCompanies(userID string, keyword string, limit, offset int) ([]models.InvoiceCompany, int64, error) {
	var companies []models.InvoiceCompany
	var total int64

	query := s.db.Model(&models.InvoiceCompany{}).Where("user_id = ?", userID)

	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR email LIKE ? OR address LIKE ?", searchPattern, searchPattern, searchPattern)
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

	if err := query.Find(&companies).Error; err != nil {
		return nil, 0, err
	}

	return companies, total, nil
}

// UpdateCompany updates an existing company
func (s *companyService) UpdateCompany(userID string, company *models.InvoiceCompany) error {
	// Verify ownership
	existing, err := s.GetCompanyByID(userID, company.ID)
	if err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Update fields
	existing.Name = company.Name
	existing.Address = company.Address
	existing.Email = company.Email
	existing.Phone = company.Phone
	existing.Website = company.Website
	existing.TaxID = company.TaxID
	existing.Notes = company.Notes

	return s.db.Save(existing).Error
}

// DeleteCompany soft-deletes a company
func (s *companyService) DeleteCompany(userID string, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.InvoiceCompany{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("company not found")
	}
	return nil
}

// SearchCompanies performs a text search on companies
func (s *companyService) SearchCompanies(userID string, query string) ([]models.InvoiceCompany, error) {
	var companies []models.InvoiceCompany
	searchPattern := "%" + query + "%"

	err := s.db.Where("user_id = ? AND (name LIKE ? OR email LIKE ? OR address LIKE ? OR notes LIKE ?)",
		userID, searchPattern, searchPattern, searchPattern, searchPattern).
		Order("name ASC").
		Find(&companies).Error

	return companies, err
}
