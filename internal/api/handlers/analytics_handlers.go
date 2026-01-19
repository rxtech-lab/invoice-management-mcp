package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
)

// GetAnalyticsSummary implements generated.StrictServerInterface
func (h *StrictHandlers) GetAnalyticsSummary(
	ctx context.Context,
	request generated.GetAnalyticsSummaryRequestObject,
) (generated.GetAnalyticsSummaryResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetAnalyticsSummary401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	period := "1m"
	if request.Params.Period != nil {
		period = string(*request.Params.Period)
	}

	summary, err := h.analyticsService.GetSummary(userID, periodParamToService(period))
	if err != nil {
		return nil, err
	}

	return generated.GetAnalyticsSummary200JSONResponse(analyticsSummaryToGenerated(summary)), nil
}

// GetAnalyticsByCategory implements generated.StrictServerInterface
func (h *StrictHandlers) GetAnalyticsByCategory(
	ctx context.Context,
	request generated.GetAnalyticsByCategoryRequestObject,
) (generated.GetAnalyticsByCategoryResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetAnalyticsByCategory401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	period := "1m"
	if request.Params.Period != nil {
		period = string(*request.Params.Period)
	}

	result, err := h.analyticsService.GetByCategory(userID, periodParamToService(period))
	if err != nil {
		return nil, err
	}

	return generated.GetAnalyticsByCategory200JSONResponse(analyticsByGroupToGenerated(result)), nil
}

// GetAnalyticsByCompany implements generated.StrictServerInterface
func (h *StrictHandlers) GetAnalyticsByCompany(
	ctx context.Context,
	request generated.GetAnalyticsByCompanyRequestObject,
) (generated.GetAnalyticsByCompanyResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetAnalyticsByCompany401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	period := "1m"
	if request.Params.Period != nil {
		period = string(*request.Params.Period)
	}

	result, err := h.analyticsService.GetByCompany(userID, periodParamToService(period))
	if err != nil {
		return nil, err
	}

	return generated.GetAnalyticsByCompany200JSONResponse(analyticsByGroupToGenerated(result)), nil
}

// GetAnalyticsByReceiver implements generated.StrictServerInterface
func (h *StrictHandlers) GetAnalyticsByReceiver(
	ctx context.Context,
	request generated.GetAnalyticsByReceiverRequestObject,
) (generated.GetAnalyticsByReceiverResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetAnalyticsByReceiver401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	period := "1m"
	if request.Params.Period != nil {
		period = string(*request.Params.Period)
	}

	result, err := h.analyticsService.GetByReceiver(userID, periodParamToService(period))
	if err != nil {
		return nil, err
	}

	return generated.GetAnalyticsByReceiver200JSONResponse(analyticsByGroupToGenerated(result)), nil
}
