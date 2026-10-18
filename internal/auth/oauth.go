package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/zerodotfive/opentunl/internal/config"
	"golang.org/x/oauth2"
)

func GetOAuth(ctx context.Context, config *config.Config) (*oidc.Provider, *oauth2.Config, error) {
	provider, err := oidc.NewProvider(ctx, config.OpenIDConfigurationIssuer)
	if err != nil {
		return nil, nil, err
	}

	return provider, &oauth2.Config{
		ClientID:     config.AuthClientID,
		ClientSecret: config.AuthClientSecret,
		RedirectURL:  config.AuthRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{"openid", "profile", "email", "groups"},
	}, nil
}
