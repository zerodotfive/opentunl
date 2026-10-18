package servercrt

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
		OperationID: "server-certificate-list",
		Method:      http.MethodGet,
		Path:        "/ca/{root_id}/server",
		Summary:     "List server certificates",
		Tags:        []string{"Server certificate"},
	}, h.list)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-certificate-revoke",
		Method:      http.MethodDelete,
		Path:        "/ca/{root_id}/server/{id}",
		Summary:     "Revoke server certificate",
		Tags:        []string{"Server certificate"},
	}, h.revoke)

	huma.Register(hapi, huma.Operation{
		OperationID: "server-certificate-create",
		Method:      http.MethodPost,
		Path:        "/ca/{root_id}/server",
		Summary:     "Create server certificate",
		Tags:        []string{"Server certificate"},
	}, h.create)
}
