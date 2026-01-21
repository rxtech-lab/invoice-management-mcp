package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// FileUnlinkService handles unlinking files from an external file server
type FileUnlinkService interface {
	UnlinkInvoiceFile(ctx context.Context, invoiceID uint, authToken string) error
}

type fileUnlinkService struct {
	fileServerURL string
	httpClient    *http.Client
}

// FileUnlinkConfig holds configuration for the file unlink service
type FileUnlinkConfig struct {
	FileServerURL string
	Timeout       time.Duration
}

// NewFileUnlinkService creates a new FileUnlinkService
func NewFileUnlinkService(cfg FileUnlinkConfig) FileUnlinkService {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	return &fileUnlinkService{
		fileServerURL: cfg.FileServerURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// UnlinkInvoiceFile attempts to unlink a file from the file server by invoice ID
// It implements retry logic with exponential backoff (3 attempts: 100ms, 200ms, 400ms)
// If the file server URL is not configured or auth token is empty, it returns nil
func (s *fileUnlinkService) UnlinkInvoiceFile(ctx context.Context, invoiceID uint, authToken string) error {
	// Skip if file server URL is not configured
	if s.fileServerURL == "" {
		return nil
	}

	// Skip if no auth token (user authenticated via other means)
	if authToken == "" {
		return nil
	}

	// Build request URL with invoice_id query parameter
	reqURL, err := url.Parse(s.fileServerURL + "/api/files/invoice")
	if err != nil {
		return fmt.Errorf("invalid file server URL: %w", err)
	}
	query := reqURL.Query()
	query.Set("invoice_id", strconv.FormatUint(uint64(invoiceID), 10))
	reqURL.RawQuery = query.Encode()

	// Retry configuration: 3 attempts with exponential backoff
	maxAttempts := 3
	backoffDurations := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Create DELETE request
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Add Authorization header
		req.Header.Set("Authorization", authToken)

		// Execute request
		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: request failed: %w", attempt+1, err)
			// Retry on network errors
			if attempt < maxAttempts-1 {
				select {
				case <-time.After(backoffDurations[attempt]):
					continue
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				}
			}
			continue
		}

		// Close response body
		resp.Body.Close()

		// Check response status
		switch resp.StatusCode {
		case http.StatusNoContent:
			// Success
			return nil
		case http.StatusNotFound:
			// File not found - consider it already unlinked
			return nil
		case http.StatusUnauthorized:
			// Auth error - don't retry
			return fmt.Errorf("unauthorized: invalid or expired token")
		case http.StatusBadRequest:
			// Bad request - don't retry
			return fmt.Errorf("bad request: status %d", resp.StatusCode)
		default:
			// Server error - retry
			lastErr = fmt.Errorf("attempt %d: unexpected status %d", attempt+1, resp.StatusCode)
			if attempt < maxAttempts-1 {
				select {
				case <-time.After(backoffDurations[attempt]):
					continue
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				}
			}
		}
	}

	// All retries exhausted
	return fmt.Errorf("failed to unlink file after %d attempts: %w", maxAttempts, lastErr)
}
