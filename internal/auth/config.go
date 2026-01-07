package auth

import (
	"github.com/yourusername/tinyrsvp/internal/config"
)

func NewOIDCConfigFromAppConfig(appCfg *config.Config) *OIDCConfig {
	return &OIDCConfig{
		IssuerURL:    appCfg.OIDC.IssuerURL,
		ClientID:     appCfg.OIDC.ClientID,
		ClientSecret: appCfg.OIDC.ClientSecret,
		RedirectURL:  appCfg.OIDC.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
	}
}
