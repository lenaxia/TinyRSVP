package auth

import (
	"github.com/lenaxia/tinyrsvp/internal/config"
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

func NewForwardAuthConfigFromAppConfig(appCfg *config.Config) *ForwardAuthConfig {
	return &ForwardAuthConfig{
		UserHeader:  appCfg.ForwardAuth.UserHeader,
		EmailHeader: appCfg.ForwardAuth.EmailHeader,
		TrustedIPs:  appCfg.ForwardAuth.TrustedIPs,
	}
}
