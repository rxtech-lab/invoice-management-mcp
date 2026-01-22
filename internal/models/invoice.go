package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// InvoiceStatus represents the payment status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusUnpaid  InvoiceStatus = "unpaid"
	InvoiceStatusOverdue InvoiceStatus = "overdue"
)

// StringArray is a custom type for storing string arrays in SQLite/databases
type StringArray []string

// Value implements driver.Valuer for database storage
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner for database retrieval
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion to []byte failed for StringArray")
	}

	if len(bytes) == 0 {
		*s = nil
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// Invoice represents a billing invoice
type Invoice struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	UserID      string `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	Title       string `gorm:"not null;type:varchar(255)" json:"title"`
	Description string `gorm:"type:text" json:"description"`

	// Billing cycle dates
	InvoiceStartedAt *time.Time `json:"invoice_started_at"`
	InvoiceEndedAt   *time.Time `json:"invoice_ended_at"`

	// Financial details
	Amount   float64 `gorm:"not null;default:0" json:"amount"`
	Currency string  `gorm:"not null;type:varchar(3);default:'USD'" json:"currency"`

	// Note: target_amount column exists in DB but is deprecated.
	// Analytics now calculate USD-normalized amounts from invoice_items.target_amount

	// Relationships
	CategoryID *uint            `gorm:"index" json:"category_id"`
	Category   *InvoiceCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CompanyID  *uint            `gorm:"index" json:"company_id"`
	Company    *InvoiceCompany  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	ReceiverID *uint            `gorm:"index" json:"receiver_id"`
	Receiver   *InvoiceReceiver `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`

	// Items (one-to-many)
	Items []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`

	// File attachment
	OriginalDownloadLink string `gorm:"type:text" json:"original_download_link"`

	// Tags - many-to-many relationship
	Tags []InvoiceTag `gorm:"many2many:invoice_tag_mappings" json:"tags,omitempty"`

	// LegacyTags stored as JSON (deprecated, kept for migration)
	LegacyTags StringArray `gorm:"column:tags;type:text" json:"-"`

	// Status and due date
	Status  InvoiceStatus `gorm:"type:varchar(20);default:'unpaid'" json:"status"`
	DueDate *time.Time    `json:"due_date"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for Invoice
func (Invoice) TableName() string {
	return "invoices"
}

// CalculateTotalFromItems calculates and updates the invoice amount from its items
func (i *Invoice) CalculateTotalFromItems() {
	var total float64
	for _, item := range i.Items {
		total += item.Amount
	}
	i.Amount = total
}
