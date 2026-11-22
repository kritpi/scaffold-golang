package handler

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/auth"
	"github.com/kritpi/scaffold-golang/internal/handler/dto"
)

// UpdateWidget godoc
// @Summary  Update widget
// @Tags     Widget
// @Accept   json
// @Produce  json
// @Param    widget-id path     string                  true "widget id"
// @Param    data      body     dto.UpdateWidgetRequest true "request body"
// @Success  200       {object} dto.SuccessResponse
// @Failure  400       {object} dto.ErrorResponse
// @Failure  404       {object} dto.ErrorResponse
// @Failure  503       {object} dto.ErrorResponse
// @Security BearerAuth
// @Failure  401 {object} dto.ErrorResponse
// @Router   /v1/widgets/{widget-id} [put]
func (h *Handler) UpdateWidget(ctx context.Context, req dto.UpdateWidgetRequest) (dto.SuccessResponse, error) {
	in, err := req.ToDomain(h.loc)
	in.Actor = auth.ActorFrom(ctx)
	if err != nil {
		return dto.SuccessResponse{}, err
	}
	if err := h.svc.Widget.Update(ctx, req.WidgetID, in); err != nil {
		return dto.SuccessResponse{}, err
	}
	return dto.Updated, nil
}
