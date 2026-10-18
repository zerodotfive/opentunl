package serverovpn

import (
	"context"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type ModelServerListResponse struct {
	Status int
	Body   *ModelServerListResponseBody
}

type ModelServerListResponseBody struct {
	Status       bool                     `json:"status"`
	Certificates []ovpn.ServerConfigShort `json:"servers"`
}
type ModelServerListRequest struct{}

var respListError = &ModelServerListResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelServerListResponseBody{
		Status: false,
	},
}

func (h *Handler) listServers(ctx context.Context, request *ModelServerListRequest) (*ModelServerListResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	servers, err := provider.ListServerConfigs()
	if err != nil {
		h.logger.Printf("error listing server configs: %s", err.Error())
		return respListError, errors.New("error listing server configs")
	}

	return &ModelServerListResponse{
		Status: http.StatusOK,
		Body: &ModelServerListResponseBody{
			Status:       true,
			Certificates: servers,
		},
	}, nil
}

func (h *Handler) listStates(ctx context.Context, request *ModelServerGetStateRequest) (*ModelServerGetStateResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("error getting server config: %s", err.Error())
		return respGetStateError, errors.New("error getting server config")
	}

	states, err := server.ListStates()
	if err != nil {
		h.logger.Printf("error listing server states: %s", err.Error())
		return respGetStateError, errors.New("error listing server states")
	}

	return &ModelServerGetStateResponse{
		Status: http.StatusOK,
		Body: &ModelServerGetStateResponseBody{
			Status: true,
			State:  states,
		},
	}, nil
}
