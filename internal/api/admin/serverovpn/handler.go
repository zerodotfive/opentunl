package serverovpn

import (
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zerodotfive/opentunl/internal/handler"
	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type Handler struct {
	postgres   *pgxpool.Pool
	supervisor *ovpn.Supervisor
	logger     *log.Logger
}

func Register(hapi huma.API, container handler.Container) {
	h := &Handler{
		postgres:   container.Postgres,
		supervisor: container.Supervisor,
		logger:     container.Logger,
	}

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-create",
		Method:      http.MethodPost,
		Path:        "/server",
		Summary:     "Create server configuration",
		Tags:        []string{"Server configuration"},
	}, h.create)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-delete",
		Method:      http.MethodDelete,
		Path:        "/server/{id}",
		Summary:     "Delete server configuration",
		Tags:        []string{"Server configuration"},
	}, h.delete)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-get",
		Method:      http.MethodGet,
		Path:        "/server/{id}",
		Summary:     "Get server configuration",
		Tags:        []string{"Server configuration"},
	}, h.getConfigs)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-get-rendered",
		Method:      http.MethodGet,
		Path:        "/server/{id}/rendered",
		Summary:     "Get server configuration rendered",
		Responses: map[string]*huma.Response{
			"default": {
				Description: "Plain text response",
				Content: map[string]*huma.MediaType{
					"text/plain": {
						Schema: &huma.Schema{Type: "string"},
					},
				},
			},
		},
		Tags: []string{"Server configuration"},
	}, h.getRendered)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-get-client-rendered",
		Method:      http.MethodPost,
		Path:        "/server/{id}/client/{name}/rendered",
		Summary:     "Get server client configuration rendered",
		Responses: map[string]*huma.Response{
			"default": {
				Description: "Plain text response",
				Content: map[string]*huma.MediaType{
					"text/plain": {
						Schema: &huma.Schema{Type: "string"},
					},
				},
			},
		},
		Tags: []string{"Server configuration"},
	}, h.getClientRendered)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-list",
		Method:      http.MethodGet,
		Path:        "/server",
		Summary:     "List server configurations",
		Tags:        []string{"Server configuration"},
	}, h.listServers)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-update",
		Method:      http.MethodPatch,
		Path:        "/server/{id}",
		Summary:     "Update server configurations",
		Tags:        []string{"Server configuration"},
	}, h.update)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-list-states",
		Method:      http.MethodGet,
		Path:        "/server/{id}/state",
		Summary:     "List server states",
		Tags:        []string{"Server configuration"},
	}, h.listStates)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-enable",
		Method:      http.MethodPost,
		Path:        "/server/{id}/state/{node_name}/enable",
		Summary:     "Enable server",
		Tags:        []string{"Server configuration"},
	}, h.enable)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-disable",
		Method:      http.MethodPost,
		Path:        "/server/{id}/state/{node_name}/disable",
		Summary:     "Disable server",
		Tags:        []string{"Server configuration"},
	}, h.disable)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-ovpn-reload",
		Method:      http.MethodPost,
		Path:        "/server/{id}/reload",
		Summary:     "Reload server",
		Tags:        []string{"Server configuration"},
	}, h.reload)
}
