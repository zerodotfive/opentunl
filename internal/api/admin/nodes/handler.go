package nodes

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
		OperationID: "node-list",
		Method:      http.MethodGet,
		Path:        "/node",
		Summary:     "List nodes",
		Tags:        []string{"Server node"},
	}, h.list)

	huma.Register(hapi, huma.Operation{
		OperationID: "node-delete",
		Method:      http.MethodDelete,
		Path:        "/node/{name}",
		Summary:     "Delete node",
		Tags:        []string{"Server node"},
	}, h.delete)
}
