package clientcrt

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelClientCrtListResponse struct {
	Status int
	Body   *ModelClientCrtListResponseBody
}

type ModelClientCrtListResponseBody struct {
	Status       bool                                  `json:"status"`
	Certificates []ca.DefaultCertificateWithRevocation `json:"certificates"`
}
type ModelClientCrtListRequest struct {
	RootID uuid.UUID `path:"root_id" json:"root_id"`
}

var respListError = &ModelClientCrtListResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelClientCrtListResponseBody{
		Status: false,
	},
}

func (h *Handler) list(ctx context.Context, request *ModelClientCrtListRequest) (*ModelClientCrtListResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	rootCert, err := provider.GetRootCertificate(request.RootID)
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respListError, errors.New("error getting root certificate")
	}

	certs, err := rootCert.ListClientCertificates()
	if err != nil {
		h.logger.Printf("error listing client certificates: %s", err.Error())
		return respListError, errors.New("error listing client certificates")
	}

	return &ModelClientCrtListResponse{
		Status: http.StatusOK,
		Body: &ModelClientCrtListResponseBody{
			Status:       true,
			Certificates: certs,
		},
	}, nil
}
