package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
)

// AddInvoiceItem implements generated.StrictServerInterface
func (h *StrictHandlers) AddInvoiceItem(
	ctx context.Context,
	request generated.AddInvoiceItemRequestObject,
) (generated.AddInvoiceItemResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.AddInvoiceItem401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	item := &models.InvoiceItem{
		Description: request.Body.Description,
		Quantity:    deref(request.Body.Quantity),
		UnitPrice:   deref(request.Body.UnitPrice),
	}

	if item.Quantity == 0 {
		item.Quantity = 1
	}

	if err := h.invoiceService.AddInvoiceItem(userID, uint(request.Id), item); err != nil {
		return generated.AddInvoiceItem400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.AddInvoiceItem201JSONResponse(invoiceItemModelToGenerated(item)), nil
}

// UpdateInvoiceItem implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateInvoiceItem(
	ctx context.Context,
	request generated.UpdateInvoiceItemRequestObject,
) (generated.UpdateInvoiceItemResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateInvoiceItem401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get existing item first
	existing, err := h.invoiceService.GetInvoiceItem(userID, uint(request.ItemId))
	if err != nil {
		return generated.UpdateInvoiceItem400JSONResponse{BadRequestJSONResponse: badRequest("Item not found")}, nil
	}

	// Update fields if provided
	if request.Body.Description != nil {
		existing.Description = *request.Body.Description
	}
	if request.Body.Quantity != nil {
		existing.Quantity = *request.Body.Quantity
	}
	if request.Body.UnitPrice != nil {
		existing.UnitPrice = *request.Body.UnitPrice
	}

	// Determine if we should force recalculation
	forceRecalculate := request.Body.AutoCalculateTargetCurrency != nil && *request.Body.AutoCalculateTargetCurrency

	// Handle manual target_amount override (only if not forcing recalculation)
	var targetAmountOverride *float64
	if !forceRecalculate && request.Body.TargetAmount != nil {
		targetAmountOverride = request.Body.TargetAmount
	}

	if err := h.invoiceService.UpdateInvoiceItem(userID, uint(request.ItemId), existing, targetAmountOverride, forceRecalculate); err != nil {
		return generated.UpdateInvoiceItem400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	updated, _ := h.invoiceService.GetInvoiceItem(userID, uint(request.ItemId))
	return generated.UpdateInvoiceItem200JSONResponse(invoiceItemModelToGenerated(updated)), nil
}

// DeleteInvoiceItem implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteInvoiceItem(
	ctx context.Context,
	request generated.DeleteInvoiceItemRequestObject,
) (generated.DeleteInvoiceItemResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteInvoiceItem401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.invoiceService.DeleteInvoiceItem(userID, uint(request.ItemId)); err != nil {
		return generated.DeleteInvoiceItem401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	return generated.DeleteInvoiceItem204Response{}, nil
}
