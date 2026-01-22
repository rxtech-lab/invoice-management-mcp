package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
)

// ListReceivers implements generated.StrictServerInterface
func (h *StrictHandlers) ListReceivers(
	ctx context.Context,
	request generated.ListReceiversRequestObject,
) (generated.ListReceiversResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ListReceivers401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	keyword := deref(request.Params.Keyword)
	limit := derefInt(request.Params.Limit, 50)
	offset := derefInt(request.Params.Offset, 0)

	receivers, total, err := h.receiverService.ListReceivers(userID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}

	data := receiverListToGenerated(receivers)

	return generated.ListReceivers200JSONResponse{
		Data:   &data,
		Total:  ptr(int(total)),
		Limit:  ptr(limit),
		Offset: ptr(offset),
	}, nil
}

// CreateReceiver implements generated.StrictServerInterface
func (h *StrictHandlers) CreateReceiver(
	ctx context.Context,
	request generated.CreateReceiverRequestObject,
) (generated.CreateReceiverResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.CreateReceiver401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	receiver := &models.InvoiceReceiver{
		Name:           request.Body.Name,
		IsOrganization: deref(request.Body.IsOrganization),
	}

	if err := h.receiverService.CreateReceiver(userID, receiver); err != nil {
		return generated.CreateReceiver400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.CreateReceiver201JSONResponse(receiverModelToGenerated(receiver)), nil
}

// GetReceiver implements generated.StrictServerInterface
func (h *StrictHandlers) GetReceiver(
	ctx context.Context,
	request generated.GetReceiverRequestObject,
) (generated.GetReceiverResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetReceiver401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	receiver, err := h.receiverService.GetReceiverByID(userID, uint(request.Id))
	if err != nil {
		return generated.GetReceiver404JSONResponse{NotFoundJSONResponse: notFound("Receiver not found")}, nil
	}

	return generated.GetReceiver200JSONResponse(receiverModelToGenerated(receiver)), nil
}

// UpdateReceiver implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateReceiver(
	ctx context.Context,
	request generated.UpdateReceiverRequestObject,
) (generated.UpdateReceiverResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateReceiver401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get existing receiver first
	existing, err := h.receiverService.GetReceiverByID(userID, uint(request.Id))
	if err != nil {
		return generated.UpdateReceiver400JSONResponse{BadRequestJSONResponse: badRequest("Receiver not found")}, nil
	}

	// Update fields if provided
	if request.Body.Name != nil {
		existing.Name = *request.Body.Name
	}
	if request.Body.IsOrganization != nil {
		existing.IsOrganization = *request.Body.IsOrganization
	}

	if err := h.receiverService.UpdateReceiver(userID, existing); err != nil {
		return generated.UpdateReceiver400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	updated, _ := h.receiverService.GetReceiverByID(userID, uint(request.Id))
	return generated.UpdateReceiver200JSONResponse(receiverModelToGenerated(updated)), nil
}

// DeleteReceiver implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteReceiver(
	ctx context.Context,
	request generated.DeleteReceiverRequestObject,
) (generated.DeleteReceiverResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteReceiver401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.receiverService.DeleteReceiver(userID, uint(request.Id)); err != nil {
		return generated.DeleteReceiver404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.DeleteReceiver204Response{}, nil
}

// MergeReceivers implements generated.StrictServerInterface
func (h *StrictHandlers) MergeReceivers(
	ctx context.Context,
	request generated.MergeReceiversRequestObject,
) (generated.MergeReceiversResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.MergeReceivers401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.MergeReceivers400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	if request.Body.TargetId == 0 {
		return generated.MergeReceivers400JSONResponse{BadRequestJSONResponse: badRequest("Target ID is required")}, nil
	}

	if len(request.Body.SourceIds) == 0 {
		return generated.MergeReceivers400JSONResponse{BadRequestJSONResponse: badRequest("Source IDs are required")}, nil
	}

	// Convert source IDs to uint
	sourceIDs := make([]uint, len(request.Body.SourceIds))
	for i, id := range request.Body.SourceIds {
		sourceIDs[i] = uint(id)
	}

	receiver, affectedCount, err := h.receiverService.MergeReceivers(userID, uint(request.Body.TargetId), sourceIDs)
	if err != nil {
		return generated.MergeReceivers404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.MergeReceivers200JSONResponse{
		Receiver:        ptr(receiverModelToGenerated(receiver)),
		MergedCount:     ptr(len(sourceIDs)),
		InvoicesUpdated: ptr(int(affectedCount)),
	}, nil
}
