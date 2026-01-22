package models

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceReceiver represents a receiver entity for invoices
type InvoiceReceiver struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         string         `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	Name           string         `gorm:"not null;type:varchar(255)" json:"name"`
	OtherNames     StringArray    `gorm:"type:text" json:"other_names"`
	IsOrganization bool           `gorm:"not null;default:false" json:"is_organization"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for InvoiceReceiver
func (InvoiceReceiver) TableName() string {
	return "invoice_receivers"
}
