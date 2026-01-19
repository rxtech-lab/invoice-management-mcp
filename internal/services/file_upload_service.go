package services

import (
	"errors"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// FileUploadService handles file upload metadata storage
type FileUploadService interface {
	// CreateFileUpload stores file metadata in the database
	CreateFileUpload(userID string, key string, filename string, contentType string, size int64) (*models.FileUpload, error)
	// GetFileUploadByKey retrieves file metadata by key
	GetFileUploadByKey(key string) (*models.FileUpload, error)
	// GetFileUploadByKeyForUser retrieves file metadata by key, verifying user ownership
	GetFileUploadByKeyForUser(userID string, key string) (*models.FileUpload, error)
	// DeleteFileUpload soft deletes a file upload record
	DeleteFileUpload(userID string, key string) error
	// ListFileUploads lists all file uploads for a user
	ListFileUploads(userID string, limit, offset int) ([]models.FileUpload, int64, error)
}

type fileUploadService struct {
	db *gorm.DB
}

// NewFileUploadService creates a new FileUploadService
func NewFileUploadService(db *gorm.DB) FileUploadService {
	return &fileUploadService{db: db}
}

// CreateFileUpload stores file metadata in the database
func (s *fileUploadService) CreateFileUpload(userID string, key string, filename string, contentType string, size int64) (*models.FileUpload, error) {
	fileUpload := &models.FileUpload{
		UserID:      userID,
		Key:         key,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
	}

	if err := s.db.Create(fileUpload).Error; err != nil {
		return nil, err
	}

	return fileUpload, nil
}

// GetFileUploadByKey retrieves file metadata by key
func (s *fileUploadService) GetFileUploadByKey(key string) (*models.FileUpload, error) {
	var fileUpload models.FileUpload
	if err := s.db.Where("key = ?", key).First(&fileUpload).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fileUpload, nil
}

// GetFileUploadByKeyForUser retrieves file metadata by key, verifying user ownership
func (s *fileUploadService) GetFileUploadByKeyForUser(userID string, key string) (*models.FileUpload, error) {
	var fileUpload models.FileUpload
	if err := s.db.Where("user_id = ? AND key = ?", userID, key).First(&fileUpload).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fileUpload, nil
}

// DeleteFileUpload soft deletes a file upload record
func (s *fileUploadService) DeleteFileUpload(userID string, key string) error {
	result := s.db.Where("user_id = ? AND key = ?", userID, key).Delete(&models.FileUpload{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListFileUploads lists all file uploads for a user
func (s *fileUploadService) ListFileUploads(userID string, limit, offset int) ([]models.FileUpload, int64, error) {
	var files []models.FileUpload
	var total int64

	// Count total
	if err := s.db.Model(&models.FileUpload{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}
