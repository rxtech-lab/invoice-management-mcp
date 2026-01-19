package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
)

// HealthCheck implements generated.StrictServerInterface
func (h *StrictHandlers) HealthCheck(
	ctx context.Context,
	request generated.HealthCheckRequestObject,
) (generated.HealthCheckResponseObject, error) {
	return generated.HealthCheck200JSONResponse{
		Status: ptr("ok"),
	}, nil
}
