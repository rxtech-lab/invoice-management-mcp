package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/models"
)

// CreateReceiverRequest represents the request body for creating a receiver
type CreateReceiverRequest struct {
	Name           string `json:"name" validate:"required"`
	IsOrganization bool   `json:"is_organization"`
}

// UpdateReceiverRequest represents the request body for updating a receiver
type UpdateReceiverRequest struct {
	Name           string `json:"name"`
	IsOrganization *bool  `json:"is_organization"`
}

// handleCreateReceiver creates a new receiver
func (s *APIServer) handleCreateReceiver(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	var req CreateReceiverRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	receiver := &models.InvoiceReceiver{
		Name:           req.Name,
		IsOrganization: req.IsOrganization,
	}

	if err := s.receiverService.CreateReceiver(userID, receiver); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(receiver)
}

// handleListReceivers lists receivers with optional search
func (s *APIServer) handleListReceivers(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	keyword := c.Query("keyword", "")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	receivers, total, err := s.receiverService.ListReceivers(userID, keyword, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":   receivers,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// handleGetReceiver retrieves a receiver by ID
func (s *APIServer) handleGetReceiver(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid receiver ID"})
	}

	receiver, err := s.receiverService.GetReceiverByID(userID, uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Receiver not found"})
	}

	return c.JSON(receiver)
}

// handleUpdateReceiver updates a receiver
func (s *APIServer) handleUpdateReceiver(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid receiver ID"})
	}

	var req UpdateReceiverRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Get existing receiver to preserve fields not being updated
	existing, err := s.receiverService.GetReceiverByID(userID, uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Receiver not found"})
	}

	// Update fields if provided
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.IsOrganization != nil {
		existing.IsOrganization = *req.IsOrganization
	}

	if err := s.receiverService.UpdateReceiver(userID, existing); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch updated receiver
	updated, _ := s.receiverService.GetReceiverByID(userID, uint(id))
	return c.JSON(updated)
}

// handleDeleteReceiver deletes a receiver
func (s *APIServer) handleDeleteReceiver(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid receiver ID"})
	}

	if err := s.receiverService.DeleteReceiver(userID, uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
