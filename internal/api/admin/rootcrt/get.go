package rootcrt

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
)

type ModelCRLGetResponse struct {
	Status      int
	ContentType string
	Body        []byte
}

type ModelCRLGetRequest struct {
	ID uuid.UUID `path:"id"`
}

var respGetErrorPlain = &ModelCRLGetResponse{
	Status:      http.StatusInternalServerError,
	ContentType: "text/plain; charset=utf-8",
	Body:        []byte("error getting crl"),
}

func (h *Handler) getCRL(ctx context.Context, request *ModelCRLGetRequest) (*ModelCRLGetResponse, error) {
	provider := ca.NewProvider(ctx, h.postgres)

	cert, err := provider.GetRootCertificate(request.ID)
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respGetErrorPlain, errors.New("error getting root certificate")
	}

	crl, err := cert.GetCRLPEM()
	if err != nil {
		h.logger.Printf("error getting root certificate: %s", err.Error())
		return respGetErrorPlain, errors.New("error getting root certificate")
	}

	return &ModelCRLGetResponse{
		Status:      http.StatusOK,
		ContentType: "text/plain; charset=utf-8",
		Body:        []byte(crl),
	}, nil
}
