package serverovpn

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type ModelServerDeleteResponse struct {
	Status int
	Body   *ModelServerDeleteResponseBody
}

type ModelServerDeleteResponseBody struct {
	Status bool `json:"status"`
}

type ModelServerDeleteRequest struct {
	ID uuid.UUID `path:"id"`
}

var respDeleteError = &ModelServerDeleteResponse{
	Status: http.StatusInternalServerError,
	Body:   nil,
}

func (h *Handler) delete(ctx context.Context, request *ModelServerDeleteRequest) (*ModelServerDeleteResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("error getting server config: %s", err.Error())
		return respDeleteError, errors.New("error getting server config")
	}

	if err := server.Delete(); err != nil {
		h.logger.Printf("error deleting server: %s", err.Error())
		return respDeleteError, errors.New("error deleting server")
	}

	return &ModelServerDeleteResponse{
		Status: http.StatusOK,
		Body: &ModelServerDeleteResponseBody{
			Status: true,
		},
	}, nil
}
