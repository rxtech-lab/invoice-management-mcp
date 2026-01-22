package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
)

// ListTags implements generated.StrictServerInterface
func (h *StrictHandlers) ListTags(
	ctx context.Context,
	request generated.ListTagsRequestObject,
) (generated.ListTagsResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ListTags401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	keyword := deref(request.Params.Keyword)
	limit := derefInt(request.Params.Limit, 50)
	offset := derefInt(request.Params.Offset, 0)

	tags, total, err := h.tagService.ListTags(userID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}

	data := tagListToGenerated(tags)

	return generated.ListTags200JSONResponse{
		Data:   &data,
		Total:  ptr(int(total)),
		Limit:  ptr(limit),
		Offset: ptr(offset),
	}, nil
}

// CreateTag implements generated.StrictServerInterface
func (h *StrictHandlers) CreateTag(
	ctx context.Context,
	request generated.CreateTagRequestObject,
) (generated.CreateTagResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.CreateTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body.Name == "" {
		return generated.CreateTag400JSONResponse{BadRequestJSONResponse: badRequest("Name is required")}, nil
	}

	tag := &models.InvoiceTag{
		Name:  request.Body.Name,
		Color: deref(request.Body.Color),
	}

	if tag.Color == "" {
		tag.Color = "#6B7280" // Default gray color
	}

	if err := h.tagService.CreateTag(userID, tag); err != nil {
		return generated.CreateTag400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.CreateTag201JSONResponse(tagModelToGenerated(tag)), nil
}

// GetTag implements generated.StrictServerInterface
func (h *StrictHandlers) GetTag(
	ctx context.Context,
	request generated.GetTagRequestObject,
) (generated.GetTagResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	tag, err := h.tagService.GetTagByID(userID, uint(request.Id))
	if err != nil {
		return generated.GetTag404JSONResponse{NotFoundJSONResponse: notFound("Tag not found")}, nil
	}

	return generated.GetTag200JSONResponse(tagModelToGenerated(tag)), nil
}

// UpdateTag implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateTag(
	ctx context.Context,
	request generated.UpdateTagRequestObject,
) (generated.UpdateTagResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get existing tag
	existing, err := h.tagService.GetTagByID(userID, uint(request.Id))
	if err != nil {
		return generated.UpdateTag400JSONResponse{BadRequestJSONResponse: badRequest("Tag not found")}, nil
	}

	// Update fields if provided
	if request.Body.Name != nil {
		existing.Name = *request.Body.Name
	}
	if request.Body.Color != nil {
		existing.Color = *request.Body.Color
	}

	if err := h.tagService.UpdateTag(userID, existing); err != nil {
		return generated.UpdateTag400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	updated, _ := h.tagService.GetTagByID(userID, uint(request.Id))
	return generated.UpdateTag200JSONResponse(tagModelToGenerated(updated)), nil
}

// DeleteTag implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteTag(
	ctx context.Context,
	request generated.DeleteTagRequestObject,
) (generated.DeleteTagResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteTag401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.tagService.DeleteTag(userID, uint(request.Id)); err != nil {
		return generated.DeleteTag404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.DeleteTag204Response{}, nil
}

// AddTagToInvoice implements generated.StrictServerInterface
func (h *StrictHandlers) AddTagToInvoice(
	ctx context.Context,
	request generated.AddTagToInvoiceRequestObject,
) (generated.AddTagToInvoiceResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.AddTagToInvoice401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.tagService.AddTagToInvoice(userID, uint(request.Id), uint(request.Body.TagId)); err != nil {
		return generated.AddTagToInvoice400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Return updated invoice
	invoice, err := h.invoiceService.GetInvoiceByID(userID, uint(request.Id))
	if err != nil {
		return generated.AddTagToInvoice404JSONResponse{NotFoundJSONResponse: notFound("Invoice not found")}, nil
	}

	return generated.AddTagToInvoice200JSONResponse(invoiceModelToGenerated(invoice)), nil
}

// RemoveTagFromInvoice implements generated.StrictServerInterface
func (h *StrictHandlers) RemoveTagFromInvoice(
	ctx context.Context,
	request generated.RemoveTagFromInvoiceRequestObject,
) (generated.RemoveTagFromInvoiceResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.RemoveTagFromInvoice401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.tagService.RemoveTagFromInvoice(userID, uint(request.Id), uint(request.TagId)); err != nil {
		return generated.RemoveTagFromInvoice404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	// Return updated invoice
	invoice, err := h.invoiceService.GetInvoiceByID(userID, uint(request.Id))
	if err != nil {
		return generated.RemoveTagFromInvoice404JSONResponse{NotFoundJSONResponse: notFound("Invoice not found")}, nil
	}

	return generated.RemoveTagFromInvoice200JSONResponse(invoiceModelToGenerated(invoice)), nil
}

// Tag converters

func tagModelToGenerated(tag *models.InvoiceTag) generated.Tag {
	return generated.Tag{
		Id:        ptr(int(tag.ID)),
		UserId:    ptr(tag.UserID),
		Name:      ptr(tag.Name),
		Color:     ptr(tag.Color),
		CreatedAt: ptr(tag.CreatedAt),
		UpdatedAt: ptr(tag.UpdatedAt),
	}
}

func tagListToGenerated(tags []models.InvoiceTag) []generated.Tag {
	result := make([]generated.Tag, len(tags))
	for i, tag := range tags {
		result[i] = tagModelToGenerated(&tag)
	}
	return result
}
