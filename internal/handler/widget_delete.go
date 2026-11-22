package handler

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/handler/dto"
)

// DeleteWidget godoc
// @Summary  Delete widget
// @Tags     Widget
// @Produce  json
// @Param    widget-id path     string true "widget id"
// @Success  200       {object} dto.SuccessResponse
// @Failure  404       {object} dto.ErrorResponse
// @Failure  503       {object} dto.ErrorResponse
// @Security BearerAuth
// @Failure  401 {object} dto.ErrorResponse
// @Router   /v1/widgets/{widget-id} [delete]
func (h *Handler) DeleteWidget(ctx context.Context, req dto.WidgetIDRequest) (dto.SuccessResponse, error) {
	if err := h.svc.Widget.Delete(ctx, req.WidgetID); err != nil {
		return dto.SuccessResponse{}, err
	}
	return dto.Deleted, nil
}
