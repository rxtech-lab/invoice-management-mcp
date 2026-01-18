package models

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceItem represents a line item within an invoice
type InvoiceItem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	InvoiceID   uint           `gorm:"index;not null" json:"invoice_id"`
	Description string         `gorm:"not null;type:varchar(255)" json:"description"`
	Quantity    float64        `gorm:"not null;default:1" json:"quantity"`
	UnitPrice   float64        `gorm:"not null;default:0" json:"unit_price"`
	Amount      float64        `gorm:"not null;default:0" json:"amount"` // Computed: Quantity * UnitPrice
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for InvoiceItem
func (InvoiceItem) TableName() string {
	return "invoice_items"
}

// CalculateAmount calculates and sets the amount based on quantity and unit price
func (i *InvoiceItem) CalculateAmount() {
	i.Amount = i.Quantity * i.UnitPrice
}

// BeforeCreate hook to calculate amount before saving
func (i *InvoiceItem) BeforeCreate(tx *gorm.DB) error {
	i.CalculateAmount()
	return nil
}

// BeforeUpdate hook to calculate amount before updating
func (i *InvoiceItem) BeforeUpdate(tx *gorm.DB) error {
	i.CalculateAmount()
	return nil
}
