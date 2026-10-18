package servercrt

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelServerCrtListResponse struct {
	Status int
	Body   *ModelServerCrtListResponseBody
}

type ModelServerCrtListResponseBody struct {
	Status       bool                                  `json:"status"`
	Certificates []ca.DefaultCertificateWithRevocation `json:"certificates"`
}
type ModelServerCrtListRequest struct {
	RootID uuid.UUID `path:"root_id" json:"root_id"`
}

var respListError = &ModelServerCrtListResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelServerCrtListResponseBody{
		Status: false,
	},
}

func (h *Handler) list(ctx context.Context, request *ModelServerCrtListRequest) (*ModelServerCrtListResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	rootCert, err := provider.GetRootCertificate(request.RootID)
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respListError, errors.New("error getting root certificate")
	}

	certs, err := rootCert.ListServerCertificates()
	if err != nil {
		h.logger.Printf("error listing server certificates: %s", err.Error())
		return respListError, errors.New("error listing server certificates")
	}

	return &ModelServerCrtListResponse{
		Status: http.StatusOK,
		Body: &ModelServerCrtListResponseBody{
			Status:       true,
			Certificates: certs,
		},
	}, nil
}
