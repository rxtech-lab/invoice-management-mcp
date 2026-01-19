package services

import (
	"fmt"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// ReceiverService handles invoice receiver business logic
type ReceiverService interface {
	CreateReceiver(userID string, receiver *models.InvoiceReceiver) error
	GetReceiverByID(userID string, id uint) (*models.InvoiceReceiver, error)
	ListReceivers(userID string, keyword string, limit, offset int) ([]models.InvoiceReceiver, int64, error)
	UpdateReceiver(userID string, receiver *models.InvoiceReceiver) error
	DeleteReceiver(userID string, id uint) error
	SearchReceivers(userID string, query string) ([]models.InvoiceReceiver, error)
}

type receiverService struct {
	db *gorm.DB
}

// NewReceiverService creates a new ReceiverService instance
func NewReceiverService(db *gorm.DB) ReceiverService {
	return &receiverService{db: db}
}

// CreateReceiver creates a new invoice receiver
func (s *receiverService) CreateReceiver(userID string, receiver *models.InvoiceReceiver) error {
	receiver.UserID = userID
	return s.db.Create(receiver).Error
}

// GetReceiverByID retrieves a receiver by ID for a specific user
func (s *receiverService) GetReceiverByID(userID string, id uint) (*models.InvoiceReceiver, error) {
	var receiver models.InvoiceReceiver
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&receiver).Error
	if err != nil {
		return nil, err
	}
	return &receiver, nil
}

// ListReceivers lists receivers with optional keyword search and pagination
func (s *receiverService) ListReceivers(userID string, keyword string, limit, offset int) ([]models.InvoiceReceiver, int64, error) {
	var receivers []models.InvoiceReceiver
	var total int64

	query := s.db.Model(&models.InvoiceReceiver{}).Where("user_id = ?", userID)

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

	if err := query.Find(&receivers).Error; err != nil {
		return nil, 0, err
	}

	return receivers, total, nil
}

// UpdateReceiver updates an existing receiver
func (s *receiverService) UpdateReceiver(userID string, receiver *models.InvoiceReceiver) error {
	// Verify ownership
	existing, err := s.GetReceiverByID(userID, receiver.ID)
	if err != nil {
		return fmt.Errorf("receiver not found: %w", err)
	}

	// Update fields
	existing.Name = receiver.Name
	existing.IsOrganization = receiver.IsOrganization

	return s.db.Save(existing).Error
}

// DeleteReceiver soft-deletes a receiver
func (s *receiverService) DeleteReceiver(userID string, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.InvoiceReceiver{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("receiver not found")
	}
	return nil
}

// SearchReceivers performs a text search on receivers
func (s *receiverService) SearchReceivers(userID string, query string) ([]models.InvoiceReceiver, error) {
	var receivers []models.InvoiceReceiver
	searchPattern := "%" + query + "%"

	err := s.db.Where("user_id = ? AND name LIKE ?",
		userID, searchPattern).
		Order("name ASC").
		Find(&receivers).Error

	return receivers, err
}
