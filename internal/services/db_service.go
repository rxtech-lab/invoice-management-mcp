package services

import (
	"database/sql"
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
	return s.db.AutoMigrate(
		&models.InvoiceCategory{},
		&models.InvoiceCompany{},
		&models.InvoiceReceiver{},
		&models.Invoice{},
		&models.InvoiceItem{},
	)
}

// Close closes the database connection
func (s *dbService) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
