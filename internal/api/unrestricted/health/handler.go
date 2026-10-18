package health

import (
	"log"

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

	huma.Get(hapi, "/health", h.get)
}
