package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
)

// ListCategories implements generated.StrictServerInterface
func (h *StrictHandlers) ListCategories(
	ctx context.Context,
	request generated.ListCategoriesRequestObject,
) (generated.ListCategoriesResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ListCategories401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	keyword := deref(request.Params.Keyword)
	limit := derefInt(request.Params.Limit, 50)
	offset := derefInt(request.Params.Offset, 0)

	categories, total, err := h.categoryService.ListCategories(userID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}

	data := categoryListToGenerated(categories)

	return generated.ListCategories200JSONResponse{
		Data:   &data,
		Total:  ptr(int(total)),
		Limit:  ptr(limit),
		Offset: ptr(offset),
	}, nil
}

// CreateCategory implements generated.StrictServerInterface
func (h *StrictHandlers) CreateCategory(
	ctx context.Context,
	request generated.CreateCategoryRequestObject,
) (generated.CreateCategoryResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.CreateCategory401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	category := &models.InvoiceCategory{
		Name:        request.Body.Name,
		Description: deref(request.Body.Description),
		Color:       deref(request.Body.Color),
	}

	if err := h.categoryService.CreateCategory(userID, category); err != nil {
		return generated.CreateCategory400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.CreateCategory201JSONResponse(categoryModelToGenerated(category)), nil
}

// GetCategory implements generated.StrictServerInterface
func (h *StrictHandlers) GetCategory(
	ctx context.Context,
	request generated.GetCategoryRequestObject,
) (generated.GetCategoryResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetCategory401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	category, err := h.categoryService.GetCategoryByID(userID, uint(request.Id))
	if err != nil {
		return generated.GetCategory404JSONResponse{NotFoundJSONResponse: notFound("Category not found")}, nil
	}

	return generated.GetCategory200JSONResponse(categoryModelToGenerated(category)), nil
}

// UpdateCategory implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateCategory(
	ctx context.Context,
	request generated.UpdateCategoryRequestObject,
) (generated.UpdateCategoryResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateCategory401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get existing category first
	existing, err := h.categoryService.GetCategoryByID(userID, uint(request.Id))
	if err != nil {
		return generated.UpdateCategory400JSONResponse{BadRequestJSONResponse: badRequest("Category not found")}, nil
	}

	// Update fields if provided
	if request.Body.Name != nil {
		existing.Name = *request.Body.Name
	}
	if request.Body.Description != nil {
		existing.Description = *request.Body.Description
	}
	if request.Body.Color != nil {
		existing.Color = *request.Body.Color
	}

	if err := h.categoryService.UpdateCategory(userID, existing); err != nil {
		return generated.UpdateCategory400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	updated, _ := h.categoryService.GetCategoryByID(userID, uint(request.Id))
	return generated.UpdateCategory200JSONResponse(categoryModelToGenerated(updated)), nil
}

// DeleteCategory implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteCategory(
	ctx context.Context,
	request generated.DeleteCategoryRequestObject,
) (generated.DeleteCategoryResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteCategory401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.categoryService.DeleteCategory(userID, uint(request.Id)); err != nil {
		return generated.DeleteCategory404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.DeleteCategory204Response{}, nil
}
