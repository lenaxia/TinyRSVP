package auth

import (
	"reflect"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/config"
)

func TestNewOIDCConfigFromAppConfig(t *testing.T) {
	tests := []struct {
		name    string
		appCfg  *config.Config
		want    *OIDCConfig
		wantErr bool
	}{
		{
			name: "all fields populated",
			appCfg: &config.Config{
				OIDC: config.OIDCConfig{
					IssuerURL:     "https://issuer.example.com",
					ClientID:      "client-id-123",
					ClientSecret:  "super-secret",
					RedirectURL:   "https://app.example.com/callback",
					SkipTLSVerify: true,
				},
			},
			want: &OIDCConfig{
				IssuerURL:     "https://issuer.example.com",
				ClientID:      "client-id-123",
				ClientSecret:  "super-secret",
				RedirectURL:   "https://app.example.com/callback",
				Scopes:        []string{"openid", "email", "profile"},
				SkipTLSVerify: false,
			},
		},
		{
			name:   "empty config yields empty oidc fields but default scopes",
			appCfg: &config.Config{},
			want: &OIDCConfig{
				IssuerURL:    "",
				ClientID:     "",
				ClientSecret: "",
				RedirectURL:  "",
				Scopes:       []string{"openid", "email", "profile"},
			},
		},
		{
			name: "partial fields preserved",
			appCfg: &config.Config{
				OIDC: config.OIDCConfig{
					IssuerURL: "https://partial.example.com",
					ClientID:  "partial-client",
				},
			},
			want: &OIDCConfig{
				IssuerURL:    "https://partial.example.com",
				ClientID:     "partial-client",
				ClientSecret: "",
				RedirectURL:  "",
				Scopes:       []string{"openid", "email", "profile"},
			},
		},
	}

	defaultScopes := []string{"openid", "email", "profile"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOIDCConfigFromAppConfig(tt.appCfg)

			if got == nil {
				t.Fatal("expected non-nil OIDCConfig, got nil")
			}

			if got.IssuerURL != tt.want.IssuerURL {
				t.Errorf("IssuerURL = %q, want %q", got.IssuerURL, tt.want.IssuerURL)
			}
			if got.ClientID != tt.want.ClientID {
				t.Errorf("ClientID = %q, want %q", got.ClientID, tt.want.ClientID)
			}
			if got.ClientSecret != tt.want.ClientSecret {
				t.Errorf("ClientSecret = %q, want %q", got.ClientSecret, tt.want.ClientSecret)
			}
			if got.RedirectURL != tt.want.RedirectURL {
				t.Errorf("RedirectURL = %q, want %q", got.RedirectURL, tt.want.RedirectURL)
			}

			if len(got.Scopes) != len(defaultScopes) {
				t.Errorf("Scopes length = %d, want %d", len(got.Scopes), len(defaultScopes))
			} else if !reflect.DeepEqual(got.Scopes, defaultScopes) {
				t.Errorf("Scopes = %v, want %v", got.Scopes, defaultScopes)
			}

			if got.SkipTLSVerify != tt.want.SkipTLSVerify {
				t.Errorf("SkipTLSVerify = %v, want %v (should not be copied from app config)", got.SkipTLSVerify, tt.want.SkipTLSVerify)
			}
		})
	}
}

func TestNewOIDCConfigFromAppConfig_NilAppConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when appCfg is nil, but did not panic")
		}
	}()
	_ = NewOIDCConfigFromAppConfig(nil)
}

func TestNewForwardAuthConfigFromAppConfig(t *testing.T) {
	tests := []struct {
		name   string
		appCfg *config.Config
		want   *ForwardAuthConfig
	}{
		{
			name: "all fields populated with multiple trusted ips",
			appCfg: &config.Config{
				ForwardAuth: config.ForwardAuthConfig{
					UserHeader:  "X-Forwarded-User",
					EmailHeader: "X-Forwarded-Email",
					TrustedIPs:  []string{"10.0.0.1", "10.0.0.2", "192.168.1.0/24"},
				},
			},
			want: &ForwardAuthConfig{
				UserHeader:  "X-Forwarded-User",
				EmailHeader: "X-Forwarded-Email",
				TrustedIPs:  []string{"10.0.0.1", "10.0.0.2", "192.168.1.0/24"},
			},
		},
		{
			name:   "empty config yields empty fields and nil trusted ips",
			appCfg: &config.Config{},
			want: &ForwardAuthConfig{
				UserHeader:  "",
				EmailHeader: "",
				TrustedIPs:  nil,
			},
		},
		{
			name: "single trusted ip",
			appCfg: &config.Config{
				ForwardAuth: config.ForwardAuthConfig{
					UserHeader:  "Remote-User",
					EmailHeader: "Remote-Email",
					TrustedIPs:  []string{"127.0.0.1"},
				},
			},
			want: &ForwardAuthConfig{
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"127.0.0.1"},
			},
		},
		{
			name: "headers only without trusted ips",
			appCfg: &config.Config{
				ForwardAuth: config.ForwardAuthConfig{
					UserHeader:  "X-User",
					EmailHeader: "X-Email",
				},
			},
			want: &ForwardAuthConfig{
				UserHeader:  "X-User",
				EmailHeader: "X-Email",
				TrustedIPs:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewForwardAuthConfigFromAppConfig(tt.appCfg)

			if got == nil {
				t.Fatal("expected non-nil ForwardAuthConfig, got nil")
			}

			if got.UserHeader != tt.want.UserHeader {
				t.Errorf("UserHeader = %q, want %q", got.UserHeader, tt.want.UserHeader)
			}
			if got.EmailHeader != tt.want.EmailHeader {
				t.Errorf("EmailHeader = %q, want %q", got.EmailHeader, tt.want.EmailHeader)
			}

			if tt.want.TrustedIPs == nil {
				if got.TrustedIPs != nil && len(got.TrustedIPs) != 0 {
					t.Errorf("TrustedIPs = %v, want nil or empty", got.TrustedIPs)
				}
			} else {
				if !reflect.DeepEqual(got.TrustedIPs, tt.want.TrustedIPs) {
					t.Errorf("TrustedIPs = %v, want %v", got.TrustedIPs, tt.want.TrustedIPs)
				}
			}
		})
	}
}

func TestNewForwardAuthConfigFromAppConfig_NilAppConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when appCfg is nil, but did not panic")
		}
	}()
	_ = NewForwardAuthConfigFromAppConfig(nil)
}
