package dh

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/dh"
)

type ModelDHDeleteResponse struct {
	Status int
	Body   *ModelDHDeleteResponseBody
}

type ModelDHDeleteResponseBody struct {
	Status bool `json:"status"`
}
type ModelDHDeleteRequest struct {
	ID uuid.UUID `path:"id" json:"id"`
}

var respDeleteError = &ModelDHDeleteResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelDHDeleteResponseBody{
		Status: false,
	},
}

func (h *Handler) delete(ctx context.Context, request *ModelDHDeleteRequest) (*ModelDHDeleteResponse, error) {
	provider := dh.NewDHProvider(ctx, h.postgres)

	d, err := provider.GetDH(request.ID)
	if err != nil {
		h.logger.Printf("error getting DH: %s", err.Error())
		return respDeleteError, errors.New("error getting DH")
	}

	if err := d.Delete(); err != nil {
		h.logger.Printf("error deleting DH: %s", err.Error())
		return respDeleteError, errors.New("error deleting DH")
	}

	return &ModelDHDeleteResponse{
		Status: http.StatusOK,
		Body: &ModelDHDeleteResponseBody{
			Status: true,
		},
	}, nil
}
