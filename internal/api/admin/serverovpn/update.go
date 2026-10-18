package serverovpn

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type ModelServerUpdateResponse struct {
	Status int
	Body   *ModelServerUpdateResponseBody
}

type ModelServerUpdateResponseBody struct {
	Status bool `json:"status"`
}
type ModelServerUpdateRequest struct {
	ID   uuid.UUID `path:"id"`
	Body struct {
		Name                 string    `json:"name"`
		ServerCertificateID  uuid.UUID `json:"server_certificate_id"`
		DHID                 uuid.UUID `json:"dh_id"`
		ServerConfigTemplate string    `json:"server_config_template"`
		ClientConfigTemplate string    `json:"client_config_template"`
		GroupsAllowed        []string  `json:"groups_allowed"`
	}
}

type ModelServerSetStateRequest struct {
	ID       uuid.UUID `path:"id"`
	NodeName string    `path:"node_name"`
}
type ModelServerScheduleReloadRequest struct {
	ID uuid.UUID `path:"id"`
}

var respUpdateError = &ModelServerUpdateResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelServerUpdateResponseBody{
		Status: false,
	},
}

func (h *Handler) update(ctx context.Context, request *ModelServerUpdateRequest) (*ModelServerUpdateResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("Error getting server config: %s", err.Error())
		return respUpdateError, nil
	}

	status, err := server.Update(
		request.Body.Name,
		request.Body.ServerCertificateID,
		request.Body.DHID,
		request.Body.ServerConfigTemplate,
		request.Body.ClientConfigTemplate,
		request.Body.GroupsAllowed,
	)
	if err != nil {
		h.logger.Printf("Error updating server: %s", err.Error())
		return respUpdateError, nil
	}

	return &ModelServerUpdateResponse{
		Status: http.StatusOK,
		Body: &ModelServerUpdateResponseBody{
			Status: status,
		},
	}, nil
}

func (h *Handler) enable(ctx context.Context, request *ModelServerSetStateRequest) (*ModelServerUpdateResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("Error getting server config: %s", err.Error())
		return respUpdateError, nil
	}

	state, err := server.GetState(request.NodeName)
	if err != nil {
		h.logger.Printf("Error getting server state: %s", err.Error())
		return respUpdateError, nil
	}

	if state.Enabled {
		return &ModelServerUpdateResponse{
			Status: http.StatusOK,
			Body: &ModelServerUpdateResponseBody{
				Status: false,
			},
		}, nil
	}

	if err := state.Enable(); err != nil {
		h.logger.Printf("Error enabling server: %s", err.Error())
		return respUpdateError, nil
	}

	return &ModelServerUpdateResponse{
		Status: http.StatusOK,
		Body: &ModelServerUpdateResponseBody{
			Status: true,
		},
	}, nil
}

func (h *Handler) disable(ctx context.Context, request *ModelServerSetStateRequest) (*ModelServerUpdateResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("Error getting server config: %s", err.Error())
		return respUpdateError, nil
	}

	state, err := server.GetState(request.NodeName)
	if err != nil {
		h.logger.Printf("Error getting server state: %s", err.Error())
		return respUpdateError, nil
	}

	if !state.Enabled {
		return &ModelServerUpdateResponse{
			Status: http.StatusOK,
			Body: &ModelServerUpdateResponseBody{
				Status: false,
			},
		}, nil
	}

	if err := state.Disable(); err != nil {
		h.logger.Printf("Error disabling server: %s", err.Error())
		return respUpdateError, nil
	}

	return &ModelServerUpdateResponse{
		Status: http.StatusOK,
		Body: &ModelServerUpdateResponseBody{
			Status: true,
		},
	}, nil
}

func (h *Handler) reload(ctx context.Context, request *ModelServerScheduleReloadRequest) (*ModelServerUpdateResponse, error) {
	provider := ovpn.NewServerProvider(ctx, h.postgres)

	server, err := provider.GetServerConfig(request.ID)
	if err != nil {
		h.logger.Printf("Error getting server config: %s", err.Error())
		return respUpdateError, errors.New("error getting server config")
	}

	if err := server.ScheduleReload(); err != nil {
		h.logger.Printf("Error reloading server: %s", err.Error())
		return respUpdateError, errors.New("error reloading server")
	}

	return &ModelServerUpdateResponse{
		Status: http.StatusOK,
		Body: &ModelServerUpdateResponseBody{
			Status: true,
		},
	}, nil
}
