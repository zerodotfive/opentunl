package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/auth"
	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type ModelServerListClientResponse struct {
	Status int
	Body   *ModelServerListClientResponseBody
}

type ModelServerListClientResponseBody struct {
	Status       bool                     `json:"status"`
	Certificates []ovpn.ServerConfigShort `json:"servers"`
}
type ModelServerListRequest struct{}

var respListError = &ModelServerListClientResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelServerListClientResponseBody{
		Status: false,
	},
}

func (h *Handler) listServers(ctx context.Context, request *ModelServerListRequest) (*ModelServerListClientResponse, error) {
	userInfo, ok := ctx.Value("userInfo").(*auth.UserInfo)
	if !ok {
		h.logger.Printf("user info not in context")
		return respListError, nil
	}

	provider := ovpn.NewServerProvider(ctx, h.postgres)

	servers, err := provider.ListAllowedServers(userInfo.Groups)
	if err != nil {
		h.logger.Printf("user %s: %s", userInfo.Login, err.Error())
		return respListError, errors.New("servers not found")
	}

	return &ModelServerListClientResponse{
		Status: http.StatusOK,
		Body: &ModelServerListClientResponseBody{
			Status:       true,
			Certificates: servers,
		},
	}, nil
}
