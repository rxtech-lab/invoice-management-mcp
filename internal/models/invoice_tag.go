package models

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceTag represents a tag for organizing invoices
type InvoiceTag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    string         `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	Name      string         `gorm:"not null;type:varchar(100)" json:"name"`
	Color     string         `gorm:"type:varchar(7)" json:"color"` // Hex color code (e.g., #FF5733)
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for InvoiceTag
func (InvoiceTag) TableName() string {
	return "invoice_tags"
}

// InvoiceTagMapping is the join table for many-to-many relationship between Invoice and InvoiceTag
type InvoiceTagMapping struct {
	InvoiceID uint `gorm:"primaryKey" json:"invoice_id"`
	TagID     uint `gorm:"primaryKey;column:invoice_tag_id" json:"invoice_tag_id"`
}

// TableName returns the table name for InvoiceTagMapping
func (InvoiceTagMapping) TableName() string {
	return "invoice_tag_mappings"
}
