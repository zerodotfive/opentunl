package config

type Config struct {
	Listen                    string
	DocsEnabled               bool
	AuthEnabled               bool
	AuthClientID              string
	AuthClientSecret          string
	AuthRedirectURL           string
	OpenIDConfigurationIssuer string
	JWTKey                    string
	AdminGroup                string
	PostgresURI               string
	Migrate                   bool
	ConfigOutputDir           string
}
