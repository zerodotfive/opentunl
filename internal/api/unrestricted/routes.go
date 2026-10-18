package unrestricted

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/zerodotfive/opentunl/internal/api/unrestricted/health"
	"github.com/zerodotfive/opentunl/internal/api/unrestricted/login"
	"github.com/zerodotfive/opentunl/internal/handler"
)

func RegisterRoutes(hapi huma.API, container handler.Container) {
	health.Register(hapi, container)
	login.Register(hapi, container)
}
