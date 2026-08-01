package config

import (
	"os"
	"strings"
	"testing"
)

func TestMetricsConfig_LoadFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    MetricsConfig
	}{
		{
			name:    "unset defaults to empty (loopback-only at middleware)",
			envVars: nil,
			want:    MetricsConfig{TrustedIPs: nil},
		},
		{
			name: "single IP",
			envVars: map[string]string{
				"METRICS_TRUSTED_IPS": "10.0.0.5",
			},
			want: MetricsConfig{TrustedIPs: []string{"10.0.0.5"}},
		},
		{
			name: "multiple IPs and CIDRs with whitespace",
			envVars: map[string]string{
				"METRICS_TRUSTED_IPS": "10.0.0.5, 172.19.0.0/16 , 192.168.1.1",
			},
			want: MetricsConfig{TrustedIPs: []string{"10.0.0.5", "172.19.0.0/16", "192.168.1.1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("METRICS_TRUSTED_IPS", "")
			if tt.envVars == nil {
				os.Unsetenv("METRICS_TRUSTED_IPS")
			} else {
				for k, v := range tt.envVars {
					t.Setenv(k, v)
				}
			}
			cfg := &Config{}
			cfg.Metrics.TrustedIPs = parseIPListFromEnv("METRICS_TRUSTED_IPS")
			if !equalIPLists(cfg.Metrics.TrustedIPs, tt.want.TrustedIPs) {
				t.Errorf("TrustedIPs = %v, want %v", cfg.Metrics.TrustedIPs, tt.want.TrustedIPs)
			}
		})
	}
}

func TestMetricsConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ips     []string
		wantErr bool
		errMsg  string
	}{
		{"empty is valid (loopback default)", nil, false, ""},
		{"valid IP", []string{"10.0.0.5"}, false, ""},
		{"valid CIDR", []string{"172.19.0.0/16"}, false, ""},
		{"valid mixed", []string{"10.0.0.5", "172.19.0.0/16"}, false, ""},
		{"invalid IP", []string{"not-an-ip"}, true, "invalid metrics IP address"},
		{"invalid CIDR", []string{"999.999.0.0/8"}, true, "invalid metrics CIDR range"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Metrics: MetricsConfig{TrustedIPs: tt.ips}}
			err := cfg.validateMetrics()
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateMetrics() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func equalIPLists(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
