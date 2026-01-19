package models

import (
	"time"

	"gorm.io/gorm"
)

// FileUpload represents a file uploaded by a user
type FileUpload struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      string         `gorm:"index;not null" json:"user_id"`
	Key         string         `gorm:"uniqueIndex;not null" json:"key"`
	Filename    string         `gorm:"not null" json:"filename"`
	ContentType string         `json:"content_type"`
	Size        int64          `json:"size"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for FileUpload
func (FileUpload) TableName() string {
	return "file_uploads"
}
