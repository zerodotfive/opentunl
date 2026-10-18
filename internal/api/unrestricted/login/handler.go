package login

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/zerodotfive/opentunl/internal/config"
	"github.com/zerodotfive/opentunl/internal/handler"
)

type Handler struct {
	config *config.Config
}

func Register(hapi huma.API, container handler.Container) {
	h := &Handler{
		config: container.Config,
	}

	huma.Get(hapi, "/auth/login", h.login)
	huma.Get(hapi, "/auth/logout", h.logout)
	huma.Get(hapi, "/auth/callback", h.callback)
	huma.Get(hapi, "/auth/info", h.info)
}
