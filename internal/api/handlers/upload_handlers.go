package handlers

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// UploadFile implements generated.StrictServerInterface
func (h *StrictHandlers) UploadFile(
	ctx context.Context,
	request generated.UploadFileRequestObject,
) (generated.UploadFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UploadFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Read file from multipart request
	file, err := request.Body.NextPart()
	if err != nil {
		return generated.UploadFile400JSONResponse{BadRequestJSONResponse: badRequest("No file provided")}, nil
	}
	defer file.Close()

	filename := file.FileName()
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return generated.UploadFile400JSONResponse{BadRequestJSONResponse: badRequest("Failed to read file")}, nil
	}

	// Upload to S3 - returns the key
	key, err := h.uploadService.UploadFile(ctx, userID, filename, content, contentType)
	if err != nil {
		return generated.UploadFile400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Save file metadata to database for ownership tracking
	_, err = h.fileUploadService.CreateFileUpload(userID, key, filename, contentType, int64(len(content)))
	if err != nil {
		// Try to clean up S3 file if DB save fails
		_ = h.uploadService.DeleteFile(ctx, key)
		return generated.UploadFile400JSONResponse{BadRequestJSONResponse: badRequest("Failed to save file metadata")}, nil
	}

	// Generate presigned download URL
	downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, key)
	if err != nil {
		// If we can't generate presigned URL, fail the upload
		return generated.UploadFile400JSONResponse{BadRequestJSONResponse: badRequest("Failed to generate download URL")}, nil
	}

	return generated.UploadFile201JSONResponse{
		Key:         key,
		DownloadUrl: downloadURL,
		Filename:    filename,
		Size:        len(content),
		ContentType: contentType,
	}, nil
}

// GetPresignedURL implements generated.StrictServerInterface
func (h *StrictHandlers) GetPresignedURL(
	ctx context.Context,
	request generated.GetPresignedURLRequestObject,
) (generated.GetPresignedURLResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetPresignedURL401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	filename := request.Params.Filename
	if filename == "" {
		return generated.GetPresignedURL400JSONResponse{BadRequestJSONResponse: badRequest("Filename is required")}, nil
	}

	contentType := "application/octet-stream"
	if request.Params.ContentType != nil {
		contentType = *request.Params.ContentType
	}

	uploadURL, key, err := h.uploadService.GetPresignedUploadURL(ctx, userID, filename, contentType)
	if err != nil {
		return generated.GetPresignedURL400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.GetPresignedURL200JSONResponse{
		UploadUrl:   ptr(uploadURL),
		Key:         ptr(key),
		ContentType: ptr(contentType),
	}, nil
}

// ConfirmPresignedUpload implements generated.StrictServerInterface
func (h *StrictHandlers) ConfirmPresignedUpload(
	ctx context.Context,
	request generated.ConfirmPresignedUploadRequestObject,
) (generated.ConfirmPresignedUploadResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ConfirmPresignedUpload401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.ConfirmPresignedUpload400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	key := request.Body.Key
	filename := request.Body.Filename
	contentType := request.Body.ContentType
	size := request.Body.Size

	// Check if file already registered
	existingFile, _ := h.fileUploadService.GetFileUploadByKey(key)
	if existingFile != nil {
		// File already registered - return success with existing data
		downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, key)
		if err != nil {
			return generated.ConfirmPresignedUpload400JSONResponse{BadRequestJSONResponse: badRequest("Failed to generate download URL")}, nil
		}
		return generated.ConfirmPresignedUpload201JSONResponse{
			Key:         key,
			DownloadUrl: downloadURL,
			Filename:    existingFile.Filename,
			Size:        int(existingFile.Size),
			ContentType: existingFile.ContentType,
		}, nil
	}

	// Save file metadata to database for ownership tracking
	_, err = h.fileUploadService.CreateFileUpload(userID, key, filename, contentType, size)
	if err != nil {
		return generated.ConfirmPresignedUpload400JSONResponse{BadRequestJSONResponse: badRequest("Failed to save file metadata")}, nil
	}

	// Generate presigned download URL
	downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, key)
	if err != nil {
		return generated.ConfirmPresignedUpload400JSONResponse{BadRequestJSONResponse: badRequest("Failed to generate download URL")}, nil
	}

	return generated.ConfirmPresignedUpload201JSONResponse{
		Key:         key,
		DownloadUrl: downloadURL,
		Filename:    filename,
		Size:        int(size),
		ContentType: contentType,
	}, nil
}

// GetFileDownloadURL implements generated.StrictServerInterface
func (h *StrictHandlers) GetFileDownloadURL(
	ctx context.Context,
	request generated.GetFileDownloadURLRequestObject,
) (generated.GetFileDownloadURLResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetFileDownloadURL401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Look up file in database to verify ownership
	fileUpload, err := h.fileUploadService.GetFileUploadByKeyForUser(userID, request.Key)
	if err != nil {
		return nil, err
	}
	if fileUpload == nil {
		// Check if file exists but belongs to another user
		existingFile, _ := h.fileUploadService.GetFileUploadByKey(request.Key)
		if existingFile != nil {
			// File exists but user doesn't own it
			return generated.GetFileDownloadURL403JSONResponse{Error: ptr("Forbidden: you do not have access to this file")}, nil
		}
		// File doesn't exist at all
		return generated.GetFileDownloadURL404JSONResponse{NotFoundJSONResponse: notFound("File not found")}, nil
	}

	// Generate presigned download URL
	downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, request.Key)
	if err != nil {
		return nil, err
	}

	// Calculate expiration time (1 hour from now)
	expiresAt := time.Now().Add(1 * time.Hour)

	return generated.GetFileDownloadURL200JSONResponse{
		DownloadUrl: ptr(downloadURL),
		Key:         ptr(request.Key),
		Filename:    ptr(fileUpload.Filename),
		ExpiresAt:   ptr(expiresAt),
	}, nil
}

// UploadHtmlToPdf implements generated.StrictServerInterface
func (h *StrictHandlers) UploadHtmlToPdf(
	ctx context.Context,
	request generated.UploadHtmlToPdfRequestObject,
) (generated.UploadHtmlToPdfResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UploadHtmlToPdf401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil || request.Body.Html == "" {
		return generated.UploadHtmlToPdf400JSONResponse{BadRequestJSONResponse: badRequest("HTML content is required")}, nil
	}

	// Check if PDF service is available
	if h.pdfService == nil {
		return generated.UploadHtmlToPdf500JSONResponse{Error: ptr("PDF service not configured")}, nil
	}

	// Build PDF conversion options
	options := services.DefaultPDFOptions()
	if request.Body.PaperWidth != nil {
		options.PaperWidth = *request.Body.PaperWidth
	}
	if request.Body.PaperHeight != nil {
		options.PaperHeight = *request.Body.PaperHeight
	}
	if request.Body.MarginTop != nil {
		options.MarginTop = *request.Body.MarginTop
	}
	if request.Body.MarginBottom != nil {
		options.MarginBottom = *request.Body.MarginBottom
	}
	if request.Body.MarginLeft != nil {
		options.MarginLeft = *request.Body.MarginLeft
	}
	if request.Body.MarginRight != nil {
		options.MarginRight = *request.Body.MarginRight
	}
	if request.Body.Landscape != nil {
		options.Landscape = *request.Body.Landscape
	}

	// Convert HTML to PDF (sanitization happens inside the service)
	pdfContent, err := h.pdfService.ConvertHTMLToPDF(ctx, request.Body.Html, options)
	if err != nil {
		return generated.UploadHtmlToPdf500JSONResponse{Error: ptr("Failed to convert HTML to PDF: " + err.Error())}, nil
	}

	// Generate filename
	filename := "generated.pdf"
	if request.Body.Filename != nil && *request.Body.Filename != "" {
		filename = *request.Body.Filename
		if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
			filename += ".pdf"
		}
	}

	// Upload to S3
	contentType := "application/pdf"
	key, err := h.uploadService.UploadFile(ctx, userID, filename, pdfContent, contentType)
	if err != nil {
		return generated.UploadHtmlToPdf400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Save file metadata to database for ownership tracking
	_, err = h.fileUploadService.CreateFileUpload(userID, key, filename, contentType, int64(len(pdfContent)))
	if err != nil {
		// Try to clean up S3 file if DB save fails
		_ = h.uploadService.DeleteFile(ctx, key)
		return generated.UploadHtmlToPdf400JSONResponse{BadRequestJSONResponse: badRequest("Failed to save file metadata")}, nil
	}

	// Generate presigned download URL
	downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, key)
	if err != nil {
		return generated.UploadHtmlToPdf400JSONResponse{BadRequestJSONResponse: badRequest("Failed to generate download URL")}, nil
	}

	return generated.UploadHtmlToPdf201JSONResponse{
		Key:         key,
		DownloadUrl: downloadURL,
		Filename:    filename,
		Size:        len(pdfContent),
		ContentType: contentType,
	}, nil
}
