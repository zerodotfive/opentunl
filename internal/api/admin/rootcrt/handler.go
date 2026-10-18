package rootcrt

import (
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zerodotfive/opentunl/internal/handler"
)

type Handler struct {
	postgres *pgxpool.Pool
	logger   *log.Logger
}

func Register(hapi huma.API, container handler.Container) {
	h := &Handler{
		postgres: container.Postgres,
		logger:   container.Logger,
	}

	huma.Register(hapi, huma.Operation{
		OperationID: "root-list",
		Method:      http.MethodGet,
		Path:        "/ca",
		Summary:     "List root certificates",
		Tags:        []string{"Root certificate"},
	}, h.list)

	huma.Register(hapi, huma.Operation{
		OperationID: "root-crl-get",
		Method:      http.MethodGet,
		Path:        "/ca/{id}/crl",
		Summary:     "Get CRL",
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
		Tags: []string{"Root certificate"},
	}, h.getCRL)

	huma.Register(hapi, huma.Operation{
		OperationID: "root-delete",
		Method:      http.MethodDelete,
		Path:        "/ca/{id}",
		Summary:     "Delete root certificate",
		Tags:        []string{"Root certificate"},
	}, h.delete)

	huma.Register(hapi, huma.Operation{
		OperationID: "root-update",
		Method:      http.MethodPut,
		Path:        "/ca/{id}",
		Summary:     "Update root certificate",
		Tags:        []string{"Root certificate"},
	}, h.update)

	huma.Register(hapi, huma.Operation{
		OperationID: "root-create",
		Method:      http.MethodPost,
		Path:        "/ca",
		Summary:     "Create root certificate",
		Tags:        []string{"Root certificate"},
	}, h.create)
}
