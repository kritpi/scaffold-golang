package handler

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/handler/dto"
)

// GetWidget godoc
// @Summary  Get widget by id
// @Tags     Widget
// @Produce  json
// @Param    widget-id path     string true "widget id"
// @Success  200       {object} dto.WidgetDetailResponse
// @Failure  404       {object} dto.ErrorResponse
// @Failure  503       {object} dto.ErrorResponse
// @Security BearerAuth
// @Failure  401 {object} dto.ErrorResponse
// @Router   /v1/widgets/{widget-id} [get]
func (h *Handler) GetWidget(ctx context.Context, req dto.WidgetIDRequest) (dto.WidgetDetailResponse, error) {
	v, err := h.svc.Widget.Get(ctx, req.WidgetID)
	if err != nil {
		return dto.WidgetDetailResponse{}, err
	}
	return dto.WidgetDetailFromDomain(v, h.loc), nil
}
