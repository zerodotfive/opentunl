package app

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/zerodotfive/opentunl/internal/api/unrestricted"
	"github.com/zerodotfive/opentunl/internal/auth"
	"github.com/zerodotfive/opentunl/internal/config"
	"github.com/zerodotfive/opentunl/internal/handler"
	"github.com/zerodotfive/opentunl/internal/info"
	"github.com/zerodotfive/opentunl/internal/logger"
	"github.com/zerodotfive/opentunl/internal/node"
	"github.com/zerodotfive/opentunl/internal/ovpn"
	"github.com/zerodotfive/opentunl/migrations"

	adminApi "github.com/zerodotfive/opentunl/internal/api/admin"
	adminUI "github.com/zerodotfive/opentunl/internal/ui/admin"

	clientApi "github.com/zerodotfive/opentunl/internal/api/client"
	clientUI "github.com/zerodotfive/opentunl/internal/ui/client"
)

func Run(config *config.Config) {
	log.Printf("opentunl version %s", info.ApplicationVersion)
	if config.AuthEnabled && len(config.JWTKey) < 32 {
		log.Fatal(errors.New("JWT key length must be at least 32 bytes"))
	}

	log.SetPrefix("[opentunl self] ")
	postgres, err := pgxpool.New(context.Background(), config.PostgresURI)
	if err != nil {
		log.Fatal(err)
	}
	defer postgres.Close()

	supervisor := ovpn.NewSupervisor(config.ConfigOutputDir, "opentunl", postgres)
	supervisor.Start()

	nodeProvider := node.NewNodeProvider(nil, postgres)
	nodeProvider.Start()
	container := handler.Container{
		Config:     config,
		Logger:     log.Default(),
		Postgres:   postgres,
		Supervisor: supervisor,
	}

	if config.Migrate {
		db := stdlib.OpenDBFromPool(postgres)
		defer func() { _ = db.Close() }()

		goose.SetBaseFS(migrations.EmbedMigrations)

		if err := goose.SetDialect("postgres"); err != nil {
			log.Fatal(err)
		}

		if err := goose.Up(db, "."); err != nil {
			log.Fatal(err)
		}
	}

	humaConfig := huma.DefaultConfig(info.Application, info.ApplicationVersion)
	humaConfig.DocsRenderer = huma.DocsRendererStoplightElements

	if !config.DocsEnabled {
		humaConfig.DocsRenderer = ""
		humaConfig.DocsPath = ""
		humaConfig.OpenAPIPath = ""
		humaConfig.SchemasPath = ""
	}

	router := http.NewServeMux()

	hapi := humago.New(router, humaConfig)
	hapi.UseMiddleware(logger.NewLoggerMiddleware(container))
	humaRoot := huma.NewGroup(hapi, "")

	humaUnrestricted := huma.NewGroup(humaRoot, "")
	unrestricted.RegisterRoutes(humaUnrestricted, container)

	humaAdminRestricted := huma.NewGroup(humaRoot, "")
	if config.AuthEnabled {
		humaAdminRestricted.UseMiddleware(auth.NewAdminAuthMiddleware(container))
	}

	humaAdminRestrictedAPI := huma.NewGroup(humaAdminRestricted, "/api/"+info.ApiVersion+"/admin")
	adminApi.RegisterRoutes(humaAdminRestrictedAPI, container)

	humaAdminRestrictedUI := huma.NewGroup(humaAdminRestricted, "/admin")
	adminUI.RegisterRoutes(humaAdminRestrictedUI, container)

	humaClientRestricted := huma.NewGroup(humaRoot, "")
	if config.AuthEnabled {
		humaClientRestricted.UseMiddleware(auth.NewClientAuthMiddleware(container))
	}

	humaClientRestrictedAPI := huma.NewGroup(humaClientRestricted, "/api/"+info.ApiVersion+"/client")
	clientApi.RegisterRoutes(humaClientRestrictedAPI, container)

	humaClientRestrictedUI := huma.NewGroup(humaClientRestricted, "")
	clientUI.RegisterRoutes(humaClientRestrictedUI, container)

	if err := http.ListenAndServe(config.Listen, router); err != nil {
		log.Fatal(err)
	}
}
