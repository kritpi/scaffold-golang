package handler

import (
	"context"

	"github.com/kritpi/scaffold-golang/internal/handler/dto"
)

// ListWidgets godoc
// @Summary  List widgets
// @Tags     Widget
// @Produce  json
// @Param    q   query    dto.ListWidgetsRequest false "paging"
// @Success  200 {object} dto.ListWidgetsResponse
// @Failure  400 {object} dto.ErrorResponse
// @Failure  503 {object} dto.ErrorResponse
// @Security BearerAuth
// @Failure  401 {object} dto.ErrorResponse
// @Router   /v1/widgets [get]
func (h *Handler) ListWidgets(ctx context.Context, req dto.ListWidgetsRequest) (dto.ListWidgetsResponse, error) {
	r, err := h.svc.Widget.List(ctx, req.ToDomain())
	if err != nil {
		return dto.ListWidgetsResponse{}, err
	}
	return dto.ListWidgetsFromDomain(r, h.loc), nil
}
