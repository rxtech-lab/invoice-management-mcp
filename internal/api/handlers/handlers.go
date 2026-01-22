package handlers

import (
	"context"
	"errors"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
)

// Ensure StrictHandlers implements the StrictServerInterface
var _ generated.StrictServerInterface = (*StrictHandlers)(nil)

// Common errors
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
)

// StrictHandlers implements the generated StrictServerInterface
type StrictHandlers struct {
	categoryService   services.CategoryService
	companyService    services.CompanyService
	receiverService   services.ReceiverService
	tagService        services.TagService
	invoiceService    services.InvoiceService
	uploadService     services.UploadService
	fileUploadService services.FileUploadService
	analyticsService  services.AnalyticsService
	fileUnlinkService services.FileUnlinkService
}

// NewStrictHandlers creates a new StrictHandlers instance
func NewStrictHandlers(
	categoryService services.CategoryService,
	companyService services.CompanyService,
	receiverService services.ReceiverService,
	tagService services.TagService,
	invoiceService services.InvoiceService,
	uploadService services.UploadService,
	fileUploadService services.FileUploadService,
	analyticsService services.AnalyticsService,
	fileUnlinkService services.FileUnlinkService,
) *StrictHandlers {
	return &StrictHandlers{
		categoryService:   categoryService,
		companyService:    companyService,
		receiverService:   receiverService,
		tagService:        tagService,
		invoiceService:    invoiceService,
		uploadService:     uploadService,
		fileUploadService: fileUploadService,
		analyticsService:  analyticsService,
		fileUnlinkService: fileUnlinkService,
	}
}

// getUserID extracts user ID from context (set by authentication middleware)
func getUserID(ctx context.Context) (string, error) {
	user, ok := utils.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		return "", ErrUnauthorized
	}
	return user.Sub, nil
}

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

// ptrIfNotEmpty returns a pointer to the given string, or nil if the string is empty
func ptrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// deref safely dereferences a pointer, returning zero value if nil
func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// derefInt safely dereferences an int pointer with a default value
func derefInt(p *int, defaultVal int) int {
	if p == nil {
		return defaultVal
	}
	return *p
}

// Error response helpers

func unauthorized() generated.UnauthorizedJSONResponse {
	return generated.UnauthorizedJSONResponse{Error: ptr("Unauthorized")}
}

func badRequest(msg string) generated.BadRequestJSONResponse {
	return generated.BadRequestJSONResponse{Error: ptr(msg)}
}

func notFound(msg string) generated.NotFoundJSONResponse {
	return generated.NotFoundJSONResponse{Error: ptr(msg)}
}
