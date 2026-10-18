package rootcrt

import (
	"context"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelRootCrtCreateResponse struct {
	Status int
	Body   *ModelRootCrtCreateResponseBody
}

type ModelRootCrtCreateResponseBody struct {
	Status bool `json:"status"`
}
type ModelRootCrtCreateRequest struct {
	Body struct {
		Name                   string `json:"name"`
		CAValidDays            int    `json:"ca_valid_days"`
		ClientDefaultValidDays int    `json:"client_default_valid_days"`
	}
}

var respCreateError = &ModelRootCrtCreateResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelRootCrtCreateResponseBody{
		Status: false,
	},
}

func (h *Handler) create(ctx context.Context, request *ModelRootCrtCreateRequest) (*ModelRootCrtCreateResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	_, err := provider.NewRootCertificate(request.Body.Name, request.Body.CAValidDays, request.Body.ClientDefaultValidDays)
	if err != nil {
		h.logger.Printf("error creating root certificate: %s", err.Error())
		return respCreateError, errors.New("error creating root certificate")
	}

	return &ModelRootCrtCreateResponse{
		Status: http.StatusOK,
		Body: &ModelRootCrtCreateResponseBody{
			Status: true,
		},
	}, nil
}
