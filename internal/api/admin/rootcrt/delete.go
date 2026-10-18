package rootcrt

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelRootCrtDeleteResponse struct {
	Status int
	Body   *ModelRootCrtDeleteResponseBody
}

type ModelRootCrtDeleteResponseBody struct {
	Status bool `json:"status"`
}
type ModelRootCrtDeleteRequest struct {
	ID uuid.UUID `path:"id" json:"id"`
}

var respDeleteError = &ModelRootCrtDeleteResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelRootCrtDeleteResponseBody{
		Status: false,
	},
}

func (h *Handler) delete(ctx context.Context, request *ModelRootCrtDeleteRequest) (*ModelRootCrtDeleteResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	cert, err := provider.GetRootCertificate(request.ID)
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respDeleteError, errors.New("error getting root certificate")
	}

	if err := cert.Delete(); err != nil {
		h.logger.Printf("error deleting root certificate: %s", err.Error())
		return respDeleteError, errors.New("error deleting root certificate")
	}

	return &ModelRootCrtDeleteResponse{
		Status: http.StatusOK,
		Body: &ModelRootCrtDeleteResponseBody{
			Status: true,
		},
	}, nil
}
