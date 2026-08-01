package config

import (
	"fmt"
)

func (c *Config) loadForwardAuthFromEnv() error {
	var err error

	c.ForwardAuth.Enabled, err = getEnvBool("FORWARD_AUTH_ENABLED", false)
	if err != nil {
		return fmt.Errorf("FORWARD_AUTH_ENABLED: %w", err)
	}

	c.ForwardAuth.UserHeader = getEnvString("FORWARD_AUTH_USER_HEADER", "")
	c.ForwardAuth.EmailHeader = getEnvString("FORWARD_AUTH_EMAIL_HEADER", "")
	c.ForwardAuth.TrustedIPs = parseIPListFromEnv("FORWARD_AUTH_TRUSTED_IPS")

	return nil
}

func (c *Config) validateForwardAuth() error {
	if !c.ForwardAuth.Enabled {
		return nil
	}

	if c.ForwardAuth.UserHeader == "" {
		return fmt.Errorf("user header is required when forward auth is enabled")
	}

	if c.ForwardAuth.EmailHeader == "" {
		return fmt.Errorf("email header is required when forward auth is enabled")
	}

	if len(c.ForwardAuth.TrustedIPs) == 0 {
		return fmt.Errorf("at least one trusted IP is required when forward auth is enabled")
	}

	return validateIPList("forward auth", c.ForwardAuth.TrustedIPs)
}

// validateMetrics validates the METRICS_TRUSTED_IPS entries, if any. An empty
// list is valid (defaults to loopback-only at the middleware layer).
func (c *Config) validateMetrics() error {
	return validateIPList("metrics", c.Metrics.TrustedIPs)
}

func (c *Config) validateAuthModes() error {
	if c.OIDC.Enabled && c.ForwardAuth.Enabled {
		return fmt.Errorf("cannot enable both OIDC and forward auth")
	}
	return nil
}

