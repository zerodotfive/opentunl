package servercrt

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelServerCrtCreateResponse struct {
	Status int
	Body   *ModelServerCrtCreateResponseBody
}

type ModelServerCrtCreateResponseBody struct {
	Status bool `json:"status"`
}
type ModelServerCrtCreateRequest struct {
	RootID uuid.UUID `path:"root_id" json:"root_id"`
	Body   struct {
		Name      string `json:"name"`
		ValidDays int    `json:"valid_days"`
	}
}

var respCreateError = &ModelServerCrtCreateResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelServerCrtCreateResponseBody{
		Status: false,
	},
}

func (h *Handler) create(ctx context.Context, request *ModelServerCrtCreateRequest) (*ModelServerCrtCreateResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	rootCert, err := provider.GetRootCertificate(request.RootID)
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respCreateError, errors.New("error getting root certificate")
	}

	_, err = rootCert.NewServerCertificate(rootCert.Name, request.Body.Name, request.Body.ValidDays)
	if err != nil {
		h.logger.Printf("error creating server certificate: %s", err.Error())
		return respCreateError, errors.New("error creating server certificate")
	}

	return &ModelServerCrtCreateResponse{
		Status: http.StatusOK,
		Body: &ModelServerCrtCreateResponseBody{
			Status: true,
		},
	}, nil
}
