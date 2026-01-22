package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rxtech-lab/invoice-management/internal/models"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBService handles database connection and lifecycle management
type DBService interface {
	GetDB() *gorm.DB
	Close() error
}

type dbService struct {
	db *gorm.DB
}

// createGormLogger creates a configured GORM logger
func createGormLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	)
}

// NewSqliteDBService creates a new DBService with SQLite connection
func NewSqliteDBService(dbPath string) (DBService, error) {
	// Handle in-memory database for testing
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: createGormLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	service := &dbService{db: db}
	if err := service.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return service, nil
}

// NewTursoDBService creates a new DBService with Turso (libsql) connection
func NewTursoDBService(databaseURL, authToken string) (DBService, error) {
	// Build connection string for libsql
	connStr := databaseURL
	if authToken != "" {
		connStr = fmt.Sprintf("%s?authToken=%s", databaseURL, authToken)
	}

	// Open libsql connection
	sqlDB, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open libsql connection: %w", err)
	}

	// Create GORM DB from sql.DB
	db, err := gorm.Open(sqlite.Dialector{
		Conn: sqlDB,
	}, &gorm.Config{
		Logger: createGormLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM connection: %w", err)
	}

	service := &dbService{db: db}
	if err := service.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return service, nil
}

// NewDBServiceFromDB creates a DBService from an existing GORM database connection
func NewDBServiceFromDB(db *gorm.DB) DBService {
	service := &dbService{db: db}
	return service
}

// GetDB returns the underlying GORM database instance
func (s *dbService) GetDB() *gorm.DB {
	return s.db
}

// migrate runs database migrations for invoice management models
func (s *dbService) migrate() error {
	if err := s.db.AutoMigrate(
		&models.InvoiceCategory{},
		&models.InvoiceCompany{},
		&models.InvoiceReceiver{},
		&models.InvoiceTag{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.FileUpload{},
	); err != nil {
		return err
	}

	// Explicitly create invoice_tag_mappings table if it doesn't exist
	// AutoMigrate may not properly handle join tables on some databases
	if err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS invoice_tag_mappings (
			invoice_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (invoice_id, tag_id),
			FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES invoice_tags(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create invoice_tag_mappings table: %w", err)
	}

	// Migrate legacy tags from JSON array to many-to-many relationship
	return s.migrateLegacyTags()
}

// migrateLegacyTags migrates existing JSON tags to the new many-to-many relationship
func (s *dbService) migrateLegacyTags() error {
	// Check if the tags column exists by querying the schema
	type columnInfo struct {
		Name string `gorm:"column:name"`
	}
	var columns []columnInfo
	if err := s.db.Raw("PRAGMA table_info(invoices)").Scan(&columns).Error; err != nil {
		// Can't query schema, skip migration
		return nil
	}

	// Check if 'tags' column exists
	hasTagsColumn := false
	for _, col := range columns {
		if col.Name == "tags" {
			hasTagsColumn = true
			break
		}
	}
	if !hasTagsColumn {
		// No legacy tags column, skip migration
		return nil
	}

	// Use raw query to get invoices with legacy tags
	type invoiceWithTags struct {
		ID     uint   `gorm:"column:id"`
		UserID string `gorm:"column:user_id"`
		Tags   string `gorm:"column:tags"`
	}

	var invoices []invoiceWithTags
	if err := s.db.Raw(`
		SELECT id, user_id, tags
		FROM invoices
		WHERE tags IS NOT NULL AND tags != '[]' AND tags != ''
	`).Scan(&invoices).Error; err != nil {
		// Query failed, skip migration
		return nil
	}

	if len(invoices) == 0 {
		return nil
	}

	// Build tag map per user: userID -> tagName -> tagID
	tagMap := make(map[string]map[string]uint)

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, inv := range invoices {
			// Parse the JSON tags string manually
			var tagNames []string
			if inv.Tags == "" {
				continue
			}
			if err := json.Unmarshal([]byte(inv.Tags), &tagNames); err != nil {
				// Invalid JSON, skip this invoice
				continue
			}
			if len(tagNames) == 0 {
				continue
			}

			if tagMap[inv.UserID] == nil {
				tagMap[inv.UserID] = make(map[string]uint)
			}

			for _, tagName := range tagNames {
				if tagName == "" {
					continue
				}

				// Check if tag already exists for this user
				tagID, exists := tagMap[inv.UserID][tagName]
				if !exists {
					// Check if it exists in the database
					var existingTag models.InvoiceTag
					err := tx.Where("user_id = ? AND name = ?", inv.UserID, tagName).First(&existingTag).Error
					if err == nil {
						tagID = existingTag.ID
					} else {
						// Create new tag
						newTag := models.InvoiceTag{
							UserID: inv.UserID,
							Name:   tagName,
							Color:  "#6B7280", // Default gray color
						}
						if err := tx.Create(&newTag).Error; err != nil {
							return err
						}
						tagID = newTag.ID
					}
					tagMap[inv.UserID][tagName] = tagID
				}

				// Create mapping if it doesn't exist
				var existingMapping models.InvoiceTagMapping
				err := tx.Where("invoice_id = ? AND tag_id = ?", inv.ID, tagID).First(&existingMapping).Error
				if err != nil {
					mapping := models.InvoiceTagMapping{
						InvoiceID: inv.ID,
						TagID:     tagID,
					}
					if err := tx.Create(&mapping).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

// Close closes the database connection
func (s *dbService) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
