package server

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
		OperationID: "server-client-rendered",
		Method:      http.MethodPost,
		Path:        "/server/{id}/rendered",
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
		OperationID: "server-client-ovpn-list",
		Method:      http.MethodGet,
		Path:        "/server",
		Summary:     "List server configurations",
		Tags:        []string{"Server configuration"},
	}, h.listServers)
}
