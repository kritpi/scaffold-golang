package handler

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/auth"
	"github.com/kritpi/scaffold-golang/internal/handler/dto"
)

// CreateWidget godoc
// @Summary  Create widget
// @Tags     Widget
// @Accept   json
// @Produce  json
// @Param    data body     dto.CreateWidgetRequest true "request body"
// @Success  201  {object} dto.CreateWidgetResponse
// @Failure  400  {object} dto.ErrorResponse
// @Failure  409  {object} dto.ErrorResponse
// @Failure  503  {object} dto.ErrorResponse
// @Security BearerAuth
// @Failure  401 {object} dto.ErrorResponse
// @Router   /v1/widgets [post]
func (h *Handler) CreateWidget(ctx context.Context, req dto.CreateWidgetRequest) (dto.CreateWidgetResponse, error) {
	in, err := req.ToDomain(h.loc)
	in.Actor = auth.ActorFrom(ctx)
	if err != nil {
		return dto.CreateWidgetResponse{}, err
	}
	id, err := h.svc.Widget.Create(ctx, in)
	if err != nil {
		return dto.CreateWidgetResponse{}, err
	}
	return dto.CreateWidgetResponse{WidgetID: id}, nil
}
