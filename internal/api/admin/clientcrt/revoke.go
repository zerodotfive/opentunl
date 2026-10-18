package clientcrt

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelClientCrtRevokeResponse struct {
	Status int
	Body   *ModelClientCrtRevokeResponseBody
}

type ModelClientCrtRevokeResponseBody struct {
	Status bool `json:"status"`
}
type ModelClientCrtRevokeRequest struct {
	RootID uuid.UUID `path:"root_id" json:"root_id"`
	ID     uuid.UUID `path:"id" json:"id"`
}

var respDeleteError = &ModelClientCrtRevokeResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelClientCrtRevokeResponseBody{
		Status: false,
	},
}

func (h *Handler) revoke(ctx context.Context, request *ModelClientCrtRevokeRequest) (*ModelClientCrtRevokeResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	rootCert, err := provider.GetRootCertificate(request.RootID)
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respDeleteError, errors.New("error getting root certificate")
	}

	cert, err := rootCert.GetClientCertificate(request.ID)
	if err != nil {
		h.logger.Printf("error getting client certificate: %s", err.Error())
		return respDeleteError, errors.New("error getting client certificate")
	}

	if err := cert.RevokeClientCertificate(request.ID); err != nil {
		h.logger.Printf("error revoking client certificate: %s", err.Error())
		return respDeleteError, errors.New("error revoking client certificate")
	}

	return &ModelClientCrtRevokeResponse{
		Status: http.StatusOK,
		Body: &ModelClientCrtRevokeResponseBody{
			Status: true,
		},
	}, nil
}
