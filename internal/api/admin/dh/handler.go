package dh

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
		OperationID: "dh-list",
		Method:      http.MethodGet,
		Path:        "/dh",
		Summary:     "List dh",
		Tags:        []string{"DH"},
	}, h.list)

	huma.Register(hapi, huma.Operation{
		OperationID: "dh-delete",
		Method:      http.MethodDelete,
		Path:        "/dh/{id}",
		Summary:     "Delete DH",
		Tags:        []string{"DH"},
	}, h.delete)

	huma.Register(hapi, huma.Operation{
		OperationID: "dh-create",
		Method:      http.MethodPost,
		Path:        "/dh",
		Summary:     "Create DH",
		Tags:        []string{"DH"},
	}, h.create)
}
