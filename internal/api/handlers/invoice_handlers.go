package handlers

import (
	"context"
	"time"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// ListInvoices implements generated.StrictServerInterface
func (h *StrictHandlers) ListInvoices(
	ctx context.Context,
	request generated.ListInvoicesRequestObject,
) (generated.ListInvoicesResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ListInvoices401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	opts := services.InvoiceListOptions{
		Keyword:   deref(request.Params.Keyword),
		Limit:     derefInt(request.Params.Limit, 50),
		Offset:    derefInt(request.Params.Offset, 0),
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if request.Params.CategoryId != nil {
		id := uint(*request.Params.CategoryId)
		opts.CategoryID = &id
	}
	if request.Params.CompanyId != nil {
		id := uint(*request.Params.CompanyId)
		opts.CompanyID = &id
	}
	if request.Params.ReceiverId != nil {
		id := uint(*request.Params.ReceiverId)
		opts.ReceiverID = &id
	}
	if request.Params.Status != nil {
		status := models.InvoiceStatus(*request.Params.Status)
		opts.Status = &status
	}
	if request.Params.SortBy != nil {
		opts.SortBy = string(*request.Params.SortBy)
	}
	if request.Params.SortOrder != nil {
		opts.SortOrder = string(*request.Params.SortOrder)
	}

	invoices, total, err := h.invoiceService.ListInvoices(userID, opts)
	if err != nil {
		return nil, err
	}

	data := invoiceListToGenerated(invoices)

	return generated.ListInvoices200JSONResponse{
		Data:   &data,
		Total:  ptr(int(total)),
		Limit:  ptr(opts.Limit),
		Offset: ptr(opts.Offset),
	}, nil
}

// CreateInvoice implements generated.StrictServerInterface
func (h *StrictHandlers) CreateInvoice(
	ctx context.Context,
	request generated.CreateInvoiceRequestObject,
) (generated.CreateInvoiceResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.CreateInvoice401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body.Title == "" {
		return generated.CreateInvoice400JSONResponse{BadRequestJSONResponse: badRequest("Title is required")}, nil
	}

	invoice := &models.Invoice{
		Title:       request.Body.Title,
		Description: deref(request.Body.Description),
		Currency:    deref(request.Body.Currency),
	}

	if invoice.Currency == "" {
		invoice.Currency = "USD"
	}

	if request.Body.InvoiceStartedAt != nil {
		invoice.InvoiceStartedAt = request.Body.InvoiceStartedAt
	}
	if request.Body.InvoiceEndedAt != nil {
		invoice.InvoiceEndedAt = request.Body.InvoiceEndedAt
	}
	if request.Body.CategoryId != nil {
		id := uint(*request.Body.CategoryId)
		invoice.CategoryID = &id
	}
	if request.Body.CompanyId != nil {
		id := uint(*request.Body.CompanyId)
		invoice.CompanyID = &id
	}
	if request.Body.ReceiverId != nil {
		id := uint(*request.Body.ReceiverId)
		invoice.ReceiverID = &id
	}
	if request.Body.OriginalDownloadLink != nil {
		invoice.OriginalDownloadLink = *request.Body.OriginalDownloadLink
	}
	if request.Body.Tags != nil {
		invoice.Tags = models.StringArray(*request.Body.Tags)
	}
	if request.Body.Status != nil {
		invoice.Status = models.InvoiceStatus(*request.Body.Status)
	} else {
		invoice.Status = models.InvoiceStatusUnpaid
	}
	if request.Body.DueDate != nil {
		invoice.DueDate = request.Body.DueDate
	}

	// Convert items if provided
	if request.Body.Items != nil {
		for _, item := range *request.Body.Items {
			invoiceItem := models.InvoiceItem{
				Description: item.Description,
				Quantity:    deref(item.Quantity),
				UnitPrice:   deref(item.UnitPrice),
			}
			if invoiceItem.Quantity == 0 {
				invoiceItem.Quantity = 1
			}
			invoice.Items = append(invoice.Items, invoiceItem)
		}
	}

	if err := h.invoiceService.CreateInvoice(userID, invoice); err != nil {
		return generated.CreateInvoice400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Reload to get relationships
	created, _ := h.invoiceService.GetInvoiceByID(userID, invoice.ID)
	return generated.CreateInvoice201JSONResponse(invoiceModelToGenerated(created)), nil
}

// GetInvoice implements generated.StrictServerInterface
func (h *StrictHandlers) GetInvoice(
	ctx context.Context,
	request generated.GetInvoiceRequestObject,
) (generated.GetInvoiceResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetInvoice401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	invoice, err := h.invoiceService.GetInvoiceByID(userID, uint(request.Id))
	if err != nil {
		return generated.GetInvoice404JSONResponse{NotFoundJSONResponse: notFound("Invoice not found")}, nil
	}

	return generated.GetInvoice200JSONResponse(invoiceModelToGenerated(invoice)), nil
}

// UpdateInvoice implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateInvoice(
	ctx context.Context,
	request generated.UpdateInvoiceRequestObject,
) (generated.UpdateInvoiceResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateInvoice401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get existing invoice first
	existing, err := h.invoiceService.GetInvoiceByID(userID, uint(request.Id))
	if err != nil {
		return generated.UpdateInvoice400JSONResponse{BadRequestJSONResponse: badRequest("Invoice not found")}, nil
	}

	// Update fields if provided
	if request.Body.Title != nil {
		existing.Title = *request.Body.Title
	}
	if request.Body.Description != nil {
		existing.Description = *request.Body.Description
	}
	if request.Body.InvoiceStartedAt != nil {
		existing.InvoiceStartedAt = request.Body.InvoiceStartedAt
	}
	if request.Body.InvoiceEndedAt != nil {
		existing.InvoiceEndedAt = request.Body.InvoiceEndedAt
	}
	if request.Body.Currency != nil {
		existing.Currency = *request.Body.Currency
	}
	if request.Body.CategoryId != nil {
		id := uint(*request.Body.CategoryId)
		existing.CategoryID = &id
	}
	if request.Body.CompanyId != nil {
		id := uint(*request.Body.CompanyId)
		existing.CompanyID = &id
	}
	if request.Body.ReceiverId != nil {
		id := uint(*request.Body.ReceiverId)
		existing.ReceiverID = &id
	}
	if request.Body.OriginalDownloadLink != nil {
		existing.OriginalDownloadLink = *request.Body.OriginalDownloadLink
	}
	if request.Body.Tags != nil {
		existing.Tags = models.StringArray(*request.Body.Tags)
	}
	if request.Body.Status != nil {
		existing.Status = models.InvoiceStatus(*request.Body.Status)
	}
	if request.Body.DueDate != nil {
		existing.DueDate = request.Body.DueDate
	}

	if err := h.invoiceService.UpdateInvoice(userID, existing); err != nil {
		return generated.UpdateInvoice400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	updated, _ := h.invoiceService.GetInvoiceByID(userID, uint(request.Id))
	return generated.UpdateInvoice200JSONResponse(invoiceModelToGenerated(updated)), nil
}

// DeleteInvoice implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteInvoice(
	ctx context.Context,
	request generated.DeleteInvoiceRequestObject,
) (generated.DeleteInvoiceResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteInvoice401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.invoiceService.DeleteInvoice(userID, uint(request.Id)); err != nil {
		return generated.DeleteInvoice404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.DeleteInvoice204Response{}, nil
}

// UpdateInvoiceStatus implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateInvoiceStatus(
	ctx context.Context,
	request generated.UpdateInvoiceStatusRequestObject,
) (generated.UpdateInvoiceStatusResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateInvoiceStatus401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	status := models.InvoiceStatus(request.Body.Status)
	if err := h.invoiceService.UpdateInvoiceStatus(userID, uint(request.Id), status); err != nil {
		return generated.UpdateInvoiceStatus400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	invoice, _ := h.invoiceService.GetInvoiceByID(userID, uint(request.Id))
	return generated.UpdateInvoiceStatus200JSONResponse(invoiceModelToGenerated(invoice)), nil
}

// Helper to convert string to time if needed
func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}
