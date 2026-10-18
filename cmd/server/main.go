package main

import (
	"github.com/spf13/viper"
	"github.com/zerodotfive/opentunl/internal/app"
	"github.com/zerodotfive/opentunl/internal/config"
)

func main() {
	viper.SetDefault("LISTEN", "127.0.0.1:8080")
	viper.SetDefault("DOCS_ENABLED", false)
	viper.SetDefault("AUTH_ENABLED", true)
	viper.SetDefault("POSTGRES_URI", "postgres://postgres:password@localhost:5432/opentunl?sslmode=prefer")
	viper.SetDefault("MIGRATE", true)
	viper.SetDefault("CONFIG_OUTPUT_DIR", "/tmp")
	viper.AutomaticEnv()

	app.Run(&config.Config{
		Listen:                    viper.GetString("LISTEN"),
		DocsEnabled:               viper.GetBool("DOCS_ENABLED"),
		AuthEnabled:               viper.GetBool("AUTH_ENABLED"),
		AuthClientID:              viper.GetString("AUTH_CLIENT_ID"),
		AuthClientSecret:          viper.GetString("AUTH_CLIENT_SECRET"),
		AuthRedirectURL:           viper.GetString("AUTH_REDIRECT_URL"),
		OpenIDConfigurationIssuer: viper.GetString("OPENID_CONFIGURATION_ISSUER"),
		JWTKey:                    viper.GetString("JWT_KEY"),
		AdminGroup:                viper.GetString("ADMIN_GROUP"),
		PostgresURI:               viper.GetString("POSTGRES_URI"),
		Migrate:                   viper.GetBool("MIGRATE"),
		ConfigOutputDir:           viper.GetString("CONFIG_OUTPUT_DIR"),
	})
}
