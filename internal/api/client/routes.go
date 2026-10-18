package admin

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/zerodotfive/opentunl/internal/api/client/server"
	"github.com/zerodotfive/opentunl/internal/handler"
)

func RegisterRoutes(hapi huma.API, container handler.Container) {
	server.Register(hapi, container)
}
