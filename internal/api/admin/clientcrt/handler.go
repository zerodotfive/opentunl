package clientcrt

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
		OperationID: "clients-certificate-list",
		Method:      http.MethodGet,
		Path:        "/ca/{root_id}/client",
		Summary:     "List client certificates",
		Tags:        []string{"Client certificate"},
	}, h.list)

	huma.Register(hapi, huma.Operation{
		OperationID: "clients-certificate-revoke",
		Method:      http.MethodDelete,
		Path:        "/ca/{root_id}/client/{id}",
		Summary:     "Revoke client certificate",
		Tags:        []string{"Client certificate"},
	}, h.revoke)
}
