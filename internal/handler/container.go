package handler

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zerodotfive/opentunl/internal/config"
	"github.com/zerodotfive/opentunl/internal/ovpn"
)

type Container struct {
	Config     *config.Config
	Logger     *log.Logger
	Postgres   *pgxpool.Pool
	Supervisor *ovpn.Supervisor
}
