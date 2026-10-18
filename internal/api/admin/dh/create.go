package dh

import (
	"context"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/dh"
)

type ModelDHCreateResponse struct {
	Status int
	Body   *ModelDHCreateResponseBody
}

type ModelDHCreateResponseBody struct {
	Status bool `json:"status"`
}
type ModelDHCreateRequest struct {
	Body struct {
		Name string `json:"name"`
	}
}

var respCreateError = &ModelDHCreateResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelDHCreateResponseBody{
		Status: false,
	},
}

func (h *Handler) create(ctx context.Context, request *ModelDHCreateRequest) (*ModelDHCreateResponse, error) {
	provider := dh.NewDHProvider(ctx, h.postgres)

	if err := provider.NewDH(request.Body.Name); err != nil {
		h.logger.Printf("error creating client: %s", err.Error())
		return respCreateError, errors.New("error creating client")
	}

	return &ModelDHCreateResponse{
		Status: http.StatusOK,
		Body: &ModelDHCreateResponseBody{
			Status: true,
		},
	}, nil
}
