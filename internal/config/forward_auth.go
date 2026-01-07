package config

import (
	"fmt"
	"net"
	"strings"
)

func (c *Config) loadForwardAuthFromEnv() error {
	var err error

	c.ForwardAuth.Enabled, err = getEnvBool("FORWARD_AUTH_ENABLED", false)
	if err != nil {
		return fmt.Errorf("FORWARD_AUTH_ENABLED: %w", err)
	}

	c.ForwardAuth.UserHeader = getEnvString("FORWARD_AUTH_USER_HEADER", "")
	c.ForwardAuth.EmailHeader = getEnvString("FORWARD_AUTH_EMAIL_HEADER", "")

	trustedIPsStr := getEnvString("FORWARD_AUTH_TRUSTED_IPS", "")
	if trustedIPsStr != "" {
		c.ForwardAuth.TrustedIPs = strings.Split(trustedIPsStr, ",")
		for i := range c.ForwardAuth.TrustedIPs {
			c.ForwardAuth.TrustedIPs[i] = strings.TrimSpace(c.ForwardAuth.TrustedIPs[i])
		}
	}

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

	for _, ip := range c.ForwardAuth.TrustedIPs {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("invalid IP address: %s", ip)
		}
	}

	return nil
}

func (c *Config) validateAuthModes() error {
	if c.OIDC.Enabled && c.ForwardAuth.Enabled {
		return fmt.Errorf("cannot enable both OIDC and forward auth")
	}
	return nil
}
