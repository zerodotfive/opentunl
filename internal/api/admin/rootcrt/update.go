package rootcrt

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelRootCrtUpdateResponse struct {
	Status int
	Body   *ModelRootCrtUpdateResponseBody
}

type ModelRootCrtUpdateResponseBody struct {
	Status bool `json:"status"`
}
type ModelRootCrtUpdateRequest struct {
	ID   uuid.UUID `path:"id"`
	Body struct {
		ClientDefaultValidDays int `json:"client_default_valid_days"`
	}
}

var respUpdateError = &ModelRootCrtUpdateResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelRootCrtUpdateResponseBody{
		Status: false,
	},
}

func (h *Handler) update(ctx context.Context, request *ModelRootCrtUpdateRequest) (*ModelRootCrtUpdateResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	cert, err := provider.GetRootCertificate(request.ID)
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respUpdateError, errors.New("error getting root certificate")
	}

	status, err := cert.UpdateRootCertificate(request.Body.ClientDefaultValidDays)
	if err != nil {
		h.logger.Printf("error updating root certificate: %s", err.Error())
		return respUpdateError, errors.New("error updating root certificate")
	}

	return &ModelRootCrtUpdateResponse{
		Status: http.StatusOK,
		Body: &ModelRootCrtUpdateResponseBody{
			Status: status,
		},
	}, nil
}
