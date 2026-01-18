package models

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceCompany represents a company/vendor associated with invoices
type InvoiceCompany struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    string         `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	Name      string         `gorm:"not null;type:varchar(255)" json:"name"`
	Address   string         `gorm:"type:text" json:"address"`
	Email     string         `gorm:"type:varchar(255)" json:"email"`
	Phone     string         `gorm:"type:varchar(50)" json:"phone"`
	Website   string         `gorm:"type:varchar(255)" json:"website"`
	TaxID     string         `gorm:"type:varchar(100)" json:"tax_id"`
	Notes     string         `gorm:"type:text" json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for InvoiceCompany
func (InvoiceCompany) TableName() string {
	return "invoice_companies"
}
