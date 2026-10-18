package server

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/auth"
	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type ModelServerGetResponseClientPlain struct {
	Status      int
	ContentType string
	Body        []byte
}

type ModelServerClientGetRenderedClientRequest struct {
	ID   uuid.UUID `path:"id"`
	Body struct {
		ForceNewCert bool `json:"force_new"`
	}
}

var respGetErrorPlain = &ModelServerGetResponseClientPlain{
	Status:      http.StatusInternalServerError,
	ContentType: "text/plain; charset=utf-8",
	Body:        []byte("error getting client configuration"),
}

func (h *Handler) getClientRendered(ctx context.Context, request *ModelServerClientGetRenderedClientRequest) (*ModelServerGetResponseClientPlain, error) {
	userInfo, ok := ctx.Value("userInfo").(*auth.UserInfo)
	if !ok {
		h.logger.Printf("user info not in context")
		return respGetErrorPlain, nil
	}

	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("user %s: %s", userInfo.Login, err.Error())
		return respGetErrorPlain, nil
	}

	configRendered, err := server.GetClientConfigRenderedClient(userInfo.Login, userInfo.Groups, request.Body.ForceNewCert, true)
	if err != nil {
		h.logger.Printf("user %s: %s", userInfo.Login, err.Error())

		if errors.Is(err, ovpn.ClientIsNotAllowed) {
			respGetErrorPlain.Body = []byte(ovpn.ClientIsNotAllowed.Error())
		}

		if errors.Is(err, ovpn.MaxCertsReached) {
			respGetErrorPlain.Body = []byte(ovpn.MaxCertsReached.Error())
		}

		return respGetErrorPlain, nil
	}

	return &ModelServerGetResponseClientPlain{
		Status:      http.StatusOK,
		ContentType: "text/plain; charset=utf-8",
		Body:        []byte(configRendered),
	}, nil
}
