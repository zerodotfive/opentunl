package servercrt

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelServerCrtRevokeResponse struct {
	Status int
	Body   *ModelServerCrtRevokeResponseBody
}

type ModelServerCrtRevokeResponseBody struct {
	Status bool `json:"status"`
}
type ModelServerCrtRevokeRequest struct {
	RootID uuid.UUID `path:"root_id" json:"root_id"`
	ID     uuid.UUID `path:"id" json:"id"`
}

var respDeleteError = &ModelServerCrtRevokeResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelServerCrtRevokeResponseBody{
		Status: false,
	},
}

func (h *Handler) revoke(ctx context.Context, request *ModelServerCrtRevokeRequest) (*ModelServerCrtRevokeResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	rootCert, err := provider.GetRootCertificate(request.RootID)
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respDeleteError, errors.New("error getting root certificate")
	}

	cert, err := rootCert.GetServerCertificate(request.ID)
	if err != nil {
		h.logger.Printf("error getting server certificate: %s", err.Error())
		return respDeleteError, errors.New("error getting server certificate")
	}

	if err := cert.RevokeServerCertificate(request.ID); err != nil {
		h.logger.Printf("error revoking server certificate: %s", err.Error())
		return respDeleteError, errors.New("error revoking server certificate")
	}

	return &ModelServerCrtRevokeResponse{
		Status: http.StatusOK,
		Body: &ModelServerCrtRevokeResponseBody{
			Status: true,
		},
	}, nil
}
