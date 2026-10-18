package admin

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/zerodotfive/opentunl/internal/api/admin/clientcrt"
	"github.com/zerodotfive/opentunl/internal/api/admin/dh"
	"github.com/zerodotfive/opentunl/internal/api/admin/nodes"
	"github.com/zerodotfive/opentunl/internal/api/admin/rootcrt"
	"github.com/zerodotfive/opentunl/internal/api/admin/servercrt"
	"github.com/zerodotfive/opentunl/internal/api/admin/serverovpn"
	"github.com/zerodotfive/opentunl/internal/handler"
)

func RegisterRoutes(hapi huma.API, container handler.Container) {
	rootcrt.Register(hapi, container)
	servercrt.Register(hapi, container)
	clientcrt.Register(hapi, container)
	serverovpn.Register(hapi, container)
	nodes.Register(hapi, container)
	dh.Register(hapi, container)
}
