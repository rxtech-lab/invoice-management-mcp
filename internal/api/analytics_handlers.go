package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// handleGetAnalyticsSummary returns invoice summary analytics
func (s *APIServer) handleGetAnalyticsSummary(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	period := services.AnalyticsPeriod(c.Query("period", "1m"))

	summary, err := s.analyticsService.GetSummary(userID, period)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(summary)
}

// handleGetAnalyticsByCategory returns analytics grouped by category
func (s *APIServer) handleGetAnalyticsByCategory(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	period := services.AnalyticsPeriod(c.Query("period", "1m"))

	result, err := s.analyticsService.GetByCategory(userID, period)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// handleGetAnalyticsByCompany returns analytics grouped by company
func (s *APIServer) handleGetAnalyticsByCompany(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	period := services.AnalyticsPeriod(c.Query("period", "1m"))

	result, err := s.analyticsService.GetByCompany(userID, period)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// handleGetAnalyticsByReceiver returns analytics grouped by receiver
func (s *APIServer) handleGetAnalyticsByReceiver(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	period := services.AnalyticsPeriod(c.Query("period", "1m"))

	result, err := s.analyticsService.GetByReceiver(userID, period)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}
