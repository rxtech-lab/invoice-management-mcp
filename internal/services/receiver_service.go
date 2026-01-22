package services

import (
	"errors"
	"fmt"
	"strings"

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
	MergeReceivers(userID string, targetID uint, sourceIDs []uint) (*models.InvoiceReceiver, int64, error)
	FindByNameOrAlias(userID string, name string) (*models.InvoiceReceiver, error)
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
	existing.OtherNames = receiver.OtherNames

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

// FindByNameOrAlias finds a receiver by name or any of its aliases (other_names)
func (s *receiverService) FindByNameOrAlias(userID string, name string) (*models.InvoiceReceiver, error) {
	normalizedName := strings.TrimSpace(strings.ToLower(name))
	if normalizedName == "" {
		return nil, errors.New("name cannot be empty")
	}

	// First check primary name (case-insensitive)
	var receiver models.InvoiceReceiver
	err := s.db.Where("user_id = ? AND LOWER(name) = ?", userID, normalizedName).First(&receiver).Error
	if err == nil {
		return &receiver, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Search in other_names JSON array
	// For SQLite/Turso, we need to fetch all receivers and check in Go
	var receivers []models.InvoiceReceiver
	err = s.db.Where("user_id = ?", userID).Find(&receivers).Error
	if err != nil {
		return nil, err
	}

	for i := range receivers {
		for _, alias := range receivers[i].OtherNames {
			if strings.ToLower(strings.TrimSpace(alias)) == normalizedName {
				return &receivers[i], nil
			}
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// MergeReceivers merges multiple receivers into a target receiver
// All invoices from source receivers are moved to the target receiver
// Source receiver names are preserved in target's other_names field
// Source receivers are then soft-deleted
// Returns the updated target receiver and count of affected invoices
func (s *receiverService) MergeReceivers(userID string, targetID uint, sourceIDs []uint) (*models.InvoiceReceiver, int64, error) {
	var target *models.InvoiceReceiver
	var affectedCount int64

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Verify target receiver ownership
		var targetReceiver models.InvoiceReceiver
		if err := tx.Where("id = ? AND user_id = ?", targetID, userID).First(&targetReceiver).Error; err != nil {
			return fmt.Errorf("target receiver not found: %w", err)
		}

		// Verify all source receivers belong to the user
		var sourceReceivers []models.InvoiceReceiver
		if err := tx.Where("id IN ? AND user_id = ?", sourceIDs, userID).Find(&sourceReceivers).Error; err != nil {
			return fmt.Errorf("error finding source receivers: %w", err)
		}

		if len(sourceReceivers) != len(sourceIDs) {
			return fmt.Errorf("some source receivers not found or don't belong to user")
		}

		// Collect names from source receivers to preserve as aliases
		existingNames := make(map[string]bool)
		existingNames[strings.ToLower(targetReceiver.Name)] = true
		for _, name := range targetReceiver.OtherNames {
			existingNames[strings.ToLower(name)] = true
		}

		newOtherNames := append([]string{}, targetReceiver.OtherNames...)
		for _, src := range sourceReceivers {
			// Add primary name if not already present
			if !existingNames[strings.ToLower(src.Name)] {
				newOtherNames = append(newOtherNames, src.Name)
				existingNames[strings.ToLower(src.Name)] = true
			}
			// Add other_names if not already present
			for _, alias := range src.OtherNames {
				if !existingNames[strings.ToLower(alias)] {
					newOtherNames = append(newOtherNames, alias)
					existingNames[strings.ToLower(alias)] = true
				}
			}
		}

		// Update target receiver's other_names
		targetReceiver.OtherNames = newOtherNames
		if err := tx.Save(&targetReceiver).Error; err != nil {
			return err
		}

		// Update all invoices from source receivers to target receiver
		result := tx.Model(&models.Invoice{}).
			Where("receiver_id IN ? AND user_id = ?", sourceIDs, userID).
			Update("receiver_id", targetID)
		if result.Error != nil {
			return result.Error
		}
		affectedCount = result.RowsAffected

		// Soft-delete source receivers
		if err := tx.Where("id IN ? AND user_id = ?", sourceIDs, userID).Delete(&models.InvoiceReceiver{}).Error; err != nil {
			return err
		}

		target = &targetReceiver
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return target, affectedCount, nil
}
