package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
)

// ListCompanies implements generated.StrictServerInterface
func (h *StrictHandlers) ListCompanies(
	ctx context.Context,
	request generated.ListCompaniesRequestObject,
) (generated.ListCompaniesResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ListCompanies401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	keyword := deref(request.Params.Keyword)
	limit := derefInt(request.Params.Limit, 50)
	offset := derefInt(request.Params.Offset, 0)

	companies, total, err := h.companyService.ListCompanies(userID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}

	data := companyListToGenerated(companies)

	return generated.ListCompanies200JSONResponse{
		Data:   &data,
		Total:  ptr(int(total)),
		Limit:  ptr(limit),
		Offset: ptr(offset),
	}, nil
}

// CreateCompany implements generated.StrictServerInterface
func (h *StrictHandlers) CreateCompany(
	ctx context.Context,
	request generated.CreateCompanyRequestObject,
) (generated.CreateCompanyResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.CreateCompany401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body.Name == "" {
		return generated.CreateCompany400JSONResponse{BadRequestJSONResponse: badRequest("Name is required")}, nil
	}

	company := &models.InvoiceCompany{
		Name:    request.Body.Name,
		Address: deref(request.Body.Address),
		Phone:   deref(request.Body.Phone),
		Website: deref(request.Body.Website),
		TaxID:   deref(request.Body.TaxId),
		Notes:   deref(request.Body.Notes),
	}

	// Handle email conversion
	if request.Body.Email != nil {
		company.Email = string(*request.Body.Email)
	}

	if err := h.companyService.CreateCompany(userID, company); err != nil {
		return generated.CreateCompany400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.CreateCompany201JSONResponse(companyModelToGenerated(company)), nil
}

// GetCompany implements generated.StrictServerInterface
func (h *StrictHandlers) GetCompany(
	ctx context.Context,
	request generated.GetCompanyRequestObject,
) (generated.GetCompanyResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetCompany401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	company, err := h.companyService.GetCompanyByID(userID, uint(request.Id))
	if err != nil {
		return generated.GetCompany404JSONResponse{NotFoundJSONResponse: notFound("Company not found")}, nil
	}

	return generated.GetCompany200JSONResponse(companyModelToGenerated(company)), nil
}

// UpdateCompany implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateCompany(
	ctx context.Context,
	request generated.UpdateCompanyRequestObject,
) (generated.UpdateCompanyResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateCompany401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get existing company first
	existing, err := h.companyService.GetCompanyByID(userID, uint(request.Id))
	if err != nil {
		return generated.UpdateCompany400JSONResponse{BadRequestJSONResponse: badRequest("Company not found")}, nil
	}

	// Update fields if provided
	if request.Body.Name != nil {
		existing.Name = *request.Body.Name
	}
	if request.Body.Address != nil {
		existing.Address = *request.Body.Address
	}
	if request.Body.Email != nil {
		existing.Email = string(*request.Body.Email)
	}
	if request.Body.Phone != nil {
		existing.Phone = *request.Body.Phone
	}
	if request.Body.Website != nil {
		existing.Website = *request.Body.Website
	}
	if request.Body.TaxId != nil {
		existing.TaxID = *request.Body.TaxId
	}
	if request.Body.Notes != nil {
		existing.Notes = *request.Body.Notes
	}

	if err := h.companyService.UpdateCompany(userID, existing); err != nil {
		return generated.UpdateCompany400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	updated, _ := h.companyService.GetCompanyByID(userID, uint(request.Id))
	return generated.UpdateCompany200JSONResponse(companyModelToGenerated(updated)), nil
}

// DeleteCompany implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteCompany(
	ctx context.Context,
	request generated.DeleteCompanyRequestObject,
) (generated.DeleteCompanyResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteCompany401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.companyService.DeleteCompany(userID, uint(request.Id)); err != nil {
		return generated.DeleteCompany404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.DeleteCompany204Response{}, nil
}
