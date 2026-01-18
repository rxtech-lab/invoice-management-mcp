package api

import (
	"github.com/gofiber/fiber/v2"
)

// PresignedURLRequest represents the request for getting a presigned URL
type PresignedURLRequest struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
}

// handleUploadFile handles direct file upload
func (s *APIServer) handleUploadFile(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
	}

	// Open the file
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer f.Close()

	// Read file content
	content := make([]byte, file.Size)
	if _, err := f.Read(content); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read file"})
	}

	// Get content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to S3
	key, err := s.uploadService.UploadFile(c.Context(), userID, file.Filename, content, contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Generate download URL
	downloadURL, err := s.uploadService.GetPresignedDownloadURL(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"key":          key,
		"download_url": downloadURL,
		"filename":     file.Filename,
		"size":         file.Size,
		"content_type": contentType,
	})
}

// handleGetPresignedURL generates a presigned URL for direct upload
func (s *APIServer) handleGetPresignedURL(c *fiber.Ctx) error {
	userID, err := s.getUserID(c)
	if err != nil {
		return err
	}

	filename := c.Query("filename")
	contentType := c.Query("content_type")

	if filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Filename is required"})
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploadURL, key, err := s.uploadService.GetPresignedUploadURL(c.Context(), userID, filename, contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"upload_url":   uploadURL,
		"key":          key,
		"content_type": contentType,
	})
}
