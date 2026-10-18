package dh

import (
	"context"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/dh"
)

type ModelDHListResponse struct {
	Status int
	Body   *ModelDHListResponseBody
}

type ModelDHListResponseBody struct {
	Status bool         `json:"status"`
	DH     []dh.DHShort `json:"dh"`
}
type ModelDHListRequest struct{}

var respListError = &ModelDHListResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelDHListResponseBody{
		Status: false,
	},
}

func (h *Handler) list(ctx context.Context, request *ModelDHListRequest) (*ModelDHListResponse, error) {
	provider := dh.NewDHProvider(ctx, h.postgres)

	ds, err := provider.ListDH()
	if err != nil {
		h.logger.Printf("error listing DH: %s", err.Error())
		return respListError, errors.New("error listing DH")
	}

	return &ModelDHListResponse{
		Status: http.StatusOK,
		Body: &ModelDHListResponseBody{
			Status: true,
			DH:     ds,
		},
	}, nil
}
