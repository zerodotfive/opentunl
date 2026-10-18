package serverovpn

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type ModelServerGetResponse struct {
	Status int
	Body   *ModelServerGetResponseBody
}
type ModelServerGetStateResponse struct {
	Status int
	Body   *ModelServerGetStateResponseBody
}

type ModelServerGetResponsePlain struct {
	Status      int
	ContentType string
	Body        []byte
}

type ModelServerGetResponseBody struct {
	Status bool               `json:"status"`
	Server *ovpn.ServerConfig `json:"server"`
}

type ModelServerGetStateResponseBody struct {
	Status bool         `json:"status"`
	State  []ovpn.State `json:"state"`
}
type ModelServerGetRequest struct {
	ID uuid.UUID `path:"id"`
}

type ModelServerClientGetRenderedRequest struct {
	ID   uuid.UUID `path:"id"`
	Name string    `path:"name"`
	Body struct {
		ValidDays    int  `json:"valid_days"`
		ForceNewCert bool `json:"force_new"`
	}
}

type ModelServerGetStateRequest struct {
	ID uuid.UUID `path:"id"`
}

var respGetError = &ModelServerGetResponse{
	Status: http.StatusInternalServerError,
	Body:   nil,
}

var respGetStateError = &ModelServerGetStateResponse{
	Status: http.StatusInternalServerError,
	Body:   nil,
}

var respGetErrorPlain = &ModelServerGetResponsePlain{
	Status:      http.StatusInternalServerError,
	ContentType: "text/plain; charset=utf-8",
	Body:        []byte("error getting server configuration"),
}

func (h *Handler) getConfigs(ctx context.Context, request *ModelServerGetRequest) (*ModelServerGetResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("error getting server configuration: %s", err.Error())
		return respGetError, errors.New("error getting server configuration")
	}

	return &ModelServerGetResponse{
		Status: http.StatusOK,
		Body: &ModelServerGetResponseBody{
			Status: true,
			Server: server,
		},
	}, nil
}

func (h *Handler) getRendered(ctx context.Context, request *ModelServerGetRequest) (*ModelServerGetResponsePlain, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("error getting server configuration: %s", err.Error())
		return respGetErrorPlain, errors.New("error getting server configuration")
	}

	configRendered, err := server.GetRendered()
	if err != nil {
		h.logger.Printf("error getting server configuration: %s", err.Error())
		return respGetErrorPlain, errors.New("error getting server configuration")
	}

	return &ModelServerGetResponsePlain{
		Status:      http.StatusOK,
		ContentType: "text/plain; charset=utf-8",
		Body:        []byte(configRendered),
	}, nil
}

func (h *Handler) getClientRendered(ctx context.Context, request *ModelServerClientGetRenderedRequest) (*ModelServerGetResponsePlain, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("error getting server configuration: %s", err.Error())
		return respGetErrorPlain, nil
	}

	configRendered, err := server.GetClientConfigRendered(request.Name, request.Body.ValidDays, request.Body.ForceNewCert, false)
	if err != nil {
		h.logger.Printf("error getting server configuration: %s", err.Error())
		return respGetErrorPlain, errors.New("error getting server configuration")
	}

	return &ModelServerGetResponsePlain{
		Status:      http.StatusOK,
		ContentType: "text/plain; charset=utf-8",
		Body:        []byte(configRendered),
	}, nil
}
