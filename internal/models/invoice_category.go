package models

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceCategory represents a category for organizing invoices
type InvoiceCategory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      string         `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	Name        string         `gorm:"not null;type:varchar(255)" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"type:varchar(7)" json:"color"` // Hex color for UI (e.g., #FF5733)
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for InvoiceCategory
func (InvoiceCategory) TableName() string {
	return "invoice_categories"
}
