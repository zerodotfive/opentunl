package rootcrt

import (
	"context"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelRootCrtListResponse struct {
	Status int
	Body   *ModelRootCrtListResponseBody
}

type ModelRootCrtListResponseBody struct {
	Status       bool                      `json:"status"`
	Certificates []ca.RootCertificateShort `json:"certificates"`
}
type ModelRootCrtListRequest struct{}

var respListError = &ModelRootCrtListResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelRootCrtListResponseBody{
		Status: false,
	},
}

func (h *Handler) list(ctx context.Context, request *ModelRootCrtListRequest) (*ModelRootCrtListResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	certs, err := provider.ListRootCertificates()
	if err != nil {
		h.logger.Printf("error listing root certificates: %s", err.Error())
		return respListError, errors.New("error listing root certificates")
	}

	return &ModelRootCrtListResponse{
		Status: http.StatusOK,
		Body: &ModelRootCrtListResponseBody{
			Status:       true,
			Certificates: certs,
		},
	}, nil
}
