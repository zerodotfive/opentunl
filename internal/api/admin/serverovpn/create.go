package serverovpn

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type ModelServerCreateResponse struct {
	Status int
	Body   *ModelServerCreateResponseBody
}

type ModelServerCreateResponseBody struct {
	Status bool `json:"status"`
}
type ModelServerCreateRequest struct {
	Body struct {
		Name                 string    `json:"name"`
		ServerCertificateID  uuid.UUID `json:"server_certificate_id"`
		DHID                 uuid.UUID `json:"dh_id"`
		ServerConfigTemplate string    `json:"server_config_template"`
		ClientConfigTemplate string    `json:"client_config_template"`
		GroupsAllowed        []string  `json:"groups_allowed"`
	}
}

var respCreateError = &ModelServerCreateResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelServerCreateResponseBody{
		Status: false,
	},
}

func (h *Handler) create(ctx context.Context, request *ModelServerCreateRequest) (*ModelServerCreateResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	_, err := provider.NewServer(
		request.Body.Name,
		request.Body.ServerCertificateID,
		request.Body.DHID,
		request.Body.ServerConfigTemplate,
		request.Body.ClientConfigTemplate,
		request.Body.GroupsAllowed,
	)
	if err != nil {
		h.logger.Printf("error creating server: %s", err.Error())
		return respCreateError, errors.New("error creating server")
	}

	return &ModelServerCreateResponse{
		Status: http.StatusOK,
		Body: &ModelServerCreateResponseBody{
			Status: true,
		},
	}, nil
}
