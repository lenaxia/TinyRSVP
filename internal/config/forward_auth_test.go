package config

import (
	"os"
	"strings"
	"testing"
)

func TestForwardAuthConfig_LoadFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    ForwardAuthConfig
		wantErr bool
	}{
		{
			name: "valid forward auth config",
			envVars: map[string]string{
				"FORWARD_AUTH_ENABLED":      "true",
				"FORWARD_AUTH_USER_HEADER":  "Remote-User",
				"FORWARD_AUTH_EMAIL_HEADER": "Remote-Email",
				"FORWARD_AUTH_TRUSTED_IPS":  "127.0.0.1,10.0.0.1",
			},
			want: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"127.0.0.1", "10.0.0.1"},
			},
			wantErr: false,
		},
		{
			name: "forward auth disabled",
			envVars: map[string]string{
				"FORWARD_AUTH_ENABLED": "false",
			},
			want: ForwardAuthConfig{
				Enabled:     false,
				UserHeader:  "",
				EmailHeader: "",
				TrustedIPs:  nil,
			},
			wantErr: false,
		},
		{
			name: "single trusted IP",
			envVars: map[string]string{
				"FORWARD_AUTH_ENABLED":      "true",
				"FORWARD_AUTH_USER_HEADER":  "X-Forwarded-User",
				"FORWARD_AUTH_EMAIL_HEADER": "X-Forwarded-Email",
				"FORWARD_AUTH_TRUSTED_IPS":  "192.168.1.1",
			},
			want: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "X-Forwarded-User",
				EmailHeader: "X-Forwarded-Email",
				TrustedIPs:  []string{"192.168.1.1"},
			},
			wantErr: false,
		},
		{
			name: "authentik headers",
			envVars: map[string]string{
				"FORWARD_AUTH_ENABLED":      "true",
				"FORWARD_AUTH_USER_HEADER":  "X-authentik-username",
				"FORWARD_AUTH_EMAIL_HEADER": "X-authentik-email",
				"FORWARD_AUTH_TRUSTED_IPS":  "172.16.0.1",
			},
			want: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "X-authentik-username",
				EmailHeader: "X-authentik-email",
				TrustedIPs:  []string{"172.16.0.1"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg := &Config{}
			err := cfg.loadForwardAuthFromEnv()

			if (err != nil) != tt.wantErr {
				t.Errorf("loadForwardAuthFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg.ForwardAuth.Enabled != tt.want.Enabled {
					t.Errorf("Enabled = %v, want %v", cfg.ForwardAuth.Enabled, tt.want.Enabled)
				}
				if cfg.ForwardAuth.UserHeader != tt.want.UserHeader {
					t.Errorf("UserHeader = %v, want %v", cfg.ForwardAuth.UserHeader, tt.want.UserHeader)
				}
				if cfg.ForwardAuth.EmailHeader != tt.want.EmailHeader {
					t.Errorf("EmailHeader = %v, want %v", cfg.ForwardAuth.EmailHeader, tt.want.EmailHeader)
				}
				if len(cfg.ForwardAuth.TrustedIPs) != len(tt.want.TrustedIPs) {
					t.Errorf("TrustedIPs length = %v, want %v", len(cfg.ForwardAuth.TrustedIPs), len(tt.want.TrustedIPs))
				} else {
					for i := range cfg.ForwardAuth.TrustedIPs {
						if cfg.ForwardAuth.TrustedIPs[i] != tt.want.TrustedIPs[i] {
							t.Errorf("TrustedIPs[%d] = %v, want %v", i, cfg.ForwardAuth.TrustedIPs[i], tt.want.TrustedIPs[i])
						}
					}
				}
			}
		})
	}
}

func TestForwardAuthConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ForwardAuthConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"127.0.0.1"},
			},
			wantErr: false,
		},
		{
			name: "disabled config is valid",
			config: ForwardAuthConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "missing user header",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"127.0.0.1"},
			},
			wantErr: true,
			errMsg:  "user header is required",
		},
		{
			name: "missing email header",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "",
				TrustedIPs:  []string{"127.0.0.1"},
			},
			wantErr: true,
			errMsg:  "email header is required",
		},
		{
			name: "missing trusted IPs",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{},
			},
			wantErr: true,
			errMsg:  "at least one trusted IP is required",
		},
		{
			name: "nil trusted IPs",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  nil,
			},
			wantErr: true,
			errMsg:  "at least one trusted IP is required",
		},
		{
			name: "invalid IP format",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"not-an-ip"},
			},
			wantErr: true,
			errMsg:  "invalid forward auth IP address",
		},
		{
			name: "valid IPv6",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"::1", "2001:db8::1"},
			},
			wantErr: false,
		},
		{
			name: "mixed IPv4 and IPv6",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"127.0.0.1", "::1"},
			},
			wantErr: false,
		},
		{
			name: "valid CIDR range",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"172.16.0.0/12"},
			},
			wantErr: false,
		},
		{
			name: "mixed IPs and CIDR",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"127.0.0.1", "::1", "172.16.0.0/12"},
			},
			wantErr: false,
		},
		{
			name: "invalid CIDR range",
			config: ForwardAuthConfig{
				Enabled:     true,
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"999.999.0.0/8"},
			},
			wantErr: true,
			errMsg:  "invalid forward auth CIDR range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{ForwardAuth: tt.config}
			err := cfg.validateForwardAuth()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateForwardAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateForwardAuth() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestForwardAuthConfig_ConflictWithOIDC(t *testing.T) {
	tests := []struct {
		name        string
		oidcEnabled bool
		fwdEnabled  bool
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "both disabled is valid",
			oidcEnabled: false,
			fwdEnabled:  false,
			wantErr:     false,
		},
		{
			name:        "only OIDC enabled",
			oidcEnabled: true,
			fwdEnabled:  false,
			wantErr:     false,
		},
		{
			name:        "only forward auth enabled",
			oidcEnabled: false,
			fwdEnabled:  true,
			wantErr:     false,
		},
		{
			name:        "both enabled should error",
			oidcEnabled: true,
			fwdEnabled:  true,
			wantErr:     true,
			errMsg:      "cannot enable both OIDC and forward auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				OIDC: OIDCConfig{
					Enabled: tt.oidcEnabled,
				},
				ForwardAuth: ForwardAuthConfig{
					Enabled: tt.fwdEnabled,
				},
			}

			if tt.fwdEnabled {
				cfg.ForwardAuth.UserHeader = "Remote-User"
				cfg.ForwardAuth.EmailHeader = "Remote-Email"
				cfg.ForwardAuth.TrustedIPs = []string{"127.0.0.1"}
			}

			err := cfg.validateAuthModes()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateAuthModes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("validateAuthModes() error = %v, want error %q", err, tt.errMsg)
				}
			}
		})
	}
}
