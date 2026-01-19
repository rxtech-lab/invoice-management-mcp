package handlers

import (
	"context"
	"io"
	"time"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
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

	// Return only the key - download URL is obtained via /api/files/{key}/download
	return generated.UploadFile201JSONResponse{
		Key:         ptr(key),
		Filename:    ptr(filename),
		Size:        ptr(len(content)),
		ContentType: ptr(contentType),
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
