package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/models"
)

// CreateCompanyRequest represents the request body for creating a company
type CreateCompanyRequest struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	TaxID   string `json:"tax_id"`
	Notes   string `json:"notes"`
}

// UpdateCompanyRequest represents the request body for updating a company
type UpdateCompanyRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	TaxID   string `json:"tax_id"`
	Notes   string `json:"notes"`
}

// handleCreateCompany creates a new company
func (s *APIServer) handleCreateCompany(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	var req CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	company := &models.InvoiceCompany{
		Name:    req.Name,
		Address: req.Address,
		Email:   req.Email,
		Phone:   req.Phone,
		Website: req.Website,
		TaxID:   req.TaxID,
		Notes:   req.Notes,
	}

	if err := s.companyService.CreateCompany(userID, company); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(company)
}

// handleListCompanies lists companies with optional search
func (s *APIServer) handleListCompanies(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	keyword := c.Query("keyword", "")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	companies, total, err := s.companyService.ListCompanies(userID, keyword, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":   companies,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// handleGetCompany retrieves a company by ID
func (s *APIServer) handleGetCompany(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid company ID"})
	}

	company, err := s.companyService.GetCompanyByID(userID, uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Company not found"})
	}

	return c.JSON(company)
}

// handleUpdateCompany updates a company
func (s *APIServer) handleUpdateCompany(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid company ID"})
	}

	var req UpdateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	company := &models.InvoiceCompany{
		ID:      uint(id),
		Name:    req.Name,
		Address: req.Address,
		Email:   req.Email,
		Phone:   req.Phone,
		Website: req.Website,
		TaxID:   req.TaxID,
		Notes:   req.Notes,
	}

	if err := s.companyService.UpdateCompany(userID, company); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch updated company
	updated, _ := s.companyService.GetCompanyByID(userID, uint(id))
	return c.JSON(updated)
}

// handleDeleteCompany deletes a company
func (s *APIServer) handleDeleteCompany(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid company ID"})
	}

	if err := s.companyService.DeleteCompany(userID, uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
