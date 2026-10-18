package client

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/zerodotfive/opentunl/internal/handler"
)

func RegisterRoutes(hapi huma.API, container handler.Container) {
	Register(hapi, container)
}
