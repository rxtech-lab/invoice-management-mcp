package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// CreateInvoiceRequest represents the request body for creating an invoice
type CreateInvoiceRequest struct {
	Title                string               `json:"title" validate:"required"`
	Description          string               `json:"description"`
	InvoiceStartedAt     *time.Time           `json:"invoice_started_at"`
	InvoiceEndedAt       *time.Time           `json:"invoice_ended_at"`
	Amount               float64              `json:"amount"`
	Currency             string               `json:"currency"`
	CategoryID           *uint                `json:"category_id"`
	CompanyID            *uint                `json:"company_id"`
	OriginalDownloadLink string               `json:"original_download_link"`
	Tags                 []string             `json:"tags"`
	Status               models.InvoiceStatus `json:"status"`
	DueDate              *time.Time           `json:"due_date"`
	Items                []CreateItemRequest  `json:"items"`
}

// CreateItemRequest represents an item in the create invoice request
type CreateItemRequest struct {
	Description string  `json:"description" validate:"required"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// UpdateInvoiceRequest represents the request body for updating an invoice
type UpdateInvoiceRequest struct {
	Title                string               `json:"title"`
	Description          string               `json:"description"`
	InvoiceStartedAt     *time.Time           `json:"invoice_started_at"`
	InvoiceEndedAt       *time.Time           `json:"invoice_ended_at"`
	Amount               float64              `json:"amount"`
	Currency             string               `json:"currency"`
	CategoryID           *uint                `json:"category_id"`
	CompanyID            *uint                `json:"company_id"`
	OriginalDownloadLink string               `json:"original_download_link"`
	Tags                 []string             `json:"tags"`
	Status               models.InvoiceStatus `json:"status"`
	DueDate              *time.Time           `json:"due_date"`
}

// UpdateStatusRequest represents the request body for updating invoice status
type UpdateStatusRequest struct {
	Status models.InvoiceStatus `json:"status" validate:"required"`
}

// AddItemRequest represents the request body for adding an invoice item
type AddItemRequest struct {
	Description string  `json:"description" validate:"required"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// UpdateItemRequest represents the request body for updating an invoice item
type UpdateItemRequest struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// handleCreateInvoice creates a new invoice
func (s *APIServer) handleCreateInvoice(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	var req CreateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required"})
	}

	// Set defaults
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.Status == "" {
		req.Status = models.InvoiceStatusUnpaid
	}

	// Convert items
	var items []models.InvoiceItem
	for _, item := range req.Items {
		quantity := item.Quantity
		if quantity == 0 {
			quantity = 1
		}
		items = append(items, models.InvoiceItem{
			Description: item.Description,
			Quantity:    quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	invoice := &models.Invoice{
		Title:                req.Title,
		Description:          req.Description,
		InvoiceStartedAt:     req.InvoiceStartedAt,
		InvoiceEndedAt:       req.InvoiceEndedAt,
		Amount:               req.Amount,
		Currency:             req.Currency,
		CategoryID:           req.CategoryID,
		CompanyID:            req.CompanyID,
		OriginalDownloadLink: req.OriginalDownloadLink,
		Tags:                 req.Tags,
		Status:               req.Status,
		DueDate:              req.DueDate,
		Items:                items,
	}

	if err := s.invoiceService.CreateInvoice(userID, invoice); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch complete invoice with relationships
	created, _ := s.invoiceService.GetInvoiceByID(userID, invoice.ID)
	return c.Status(fiber.StatusCreated).JSON(created)
}

// handleListInvoices lists invoices with filtering and sorting
func (s *APIServer) handleListInvoices(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	// Parse query parameters
	opts := services.InvoiceListOptions{
		Keyword:   c.Query("keyword", ""),
		SortBy:    c.Query("sort_by", "created_at"),
		SortOrder: c.Query("sort_order", "desc"),
	}

	if limit, err := strconv.Atoi(c.Query("limit", "50")); err == nil {
		opts.Limit = limit
	}
	if offset, err := strconv.Atoi(c.Query("offset", "0")); err == nil {
		opts.Offset = offset
	}
	if categoryID, err := strconv.ParseUint(c.Query("category_id", ""), 10, 32); err == nil {
		id := uint(categoryID)
		opts.CategoryID = &id
	}
	if companyID, err := strconv.ParseUint(c.Query("company_id", ""), 10, 32); err == nil {
		id := uint(companyID)
		opts.CompanyID = &id
	}
	if status := c.Query("status", ""); status != "" {
		s := models.InvoiceStatus(status)
		opts.Status = &s
	}

	invoices, total, err := s.invoiceService.ListInvoices(userID, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":   invoices,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
}

// handleGetInvoice retrieves an invoice by ID
func (s *APIServer) handleGetInvoice(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid invoice ID"})
	}

	invoice, err := s.invoiceService.GetInvoiceByID(userID, uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invoice not found"})
	}

	return c.JSON(invoice)
}

// handleUpdateInvoice updates an invoice
func (s *APIServer) handleUpdateInvoice(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid invoice ID"})
	}

	var req UpdateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	invoice := &models.Invoice{
		ID:                   uint(id),
		Title:                req.Title,
		Description:          req.Description,
		InvoiceStartedAt:     req.InvoiceStartedAt,
		InvoiceEndedAt:       req.InvoiceEndedAt,
		Amount:               req.Amount,
		Currency:             req.Currency,
		CategoryID:           req.CategoryID,
		CompanyID:            req.CompanyID,
		OriginalDownloadLink: req.OriginalDownloadLink,
		Tags:                 req.Tags,
		Status:               req.Status,
		DueDate:              req.DueDate,
	}

	if err := s.invoiceService.UpdateInvoice(userID, invoice); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch updated invoice
	updated, _ := s.invoiceService.GetInvoiceByID(userID, uint(id))
	return c.JSON(updated)
}

// handleDeleteInvoice deletes an invoice
func (s *APIServer) handleDeleteInvoice(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid invoice ID"})
	}

	if err := s.invoiceService.DeleteInvoice(userID, uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// handleUpdateInvoiceStatus updates only the status of an invoice
func (s *APIServer) handleUpdateInvoiceStatus(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid invoice ID"})
	}

	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Status is required"})
	}

	if err := s.invoiceService.UpdateInvoiceStatus(userID, uint(id), req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch updated invoice
	updated, _ := s.invoiceService.GetInvoiceByID(userID, uint(id))
	return c.JSON(updated)
}

// handleAddInvoiceItem adds an item to an invoice
func (s *APIServer) handleAddInvoiceItem(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	invoiceID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid invoice ID"})
	}

	var req AddItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Description is required"})
	}

	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}

	item := &models.InvoiceItem{
		Description: req.Description,
		Quantity:    quantity,
		UnitPrice:   req.UnitPrice,
	}

	if err := s.invoiceService.AddInvoiceItem(userID, uint(invoiceID), item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// handleUpdateInvoiceItem updates an invoice item
func (s *APIServer) handleUpdateInvoiceItem(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	itemID, err := strconv.ParseUint(c.Params("item_id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid item ID"})
	}

	var req UpdateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	item := &models.InvoiceItem{
		Description: req.Description,
		Quantity:    req.Quantity,
		UnitPrice:   req.UnitPrice,
	}

	if err := s.invoiceService.UpdateInvoiceItem(userID, uint(itemID), item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch updated item
	updated, _ := s.invoiceService.GetInvoiceItem(userID, uint(itemID))
	return c.JSON(updated)
}

// handleDeleteInvoiceItem deletes an invoice item
func (s *APIServer) handleDeleteInvoiceItem(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	itemID, err := strconv.ParseUint(c.Params("item_id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid item ID"})
	}

	if err := s.invoiceService.DeleteInvoiceItem(userID, uint(itemID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
