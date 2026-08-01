package config

import (
	"fmt"
	"net"
	"strings"
)

// parseIPListFromEnv reads a comma-separated list of IPs or CIDR ranges from
// the named environment variable and trims whitespace from each entry.
func parseIPListFromEnv(name string) []string {
	raw := getEnvString(name, "")
	if raw == "" {
		return nil
	}
	entries := strings.Split(raw, ",")
	for i := range entries {
		entries[i] = strings.TrimSpace(entries[i])
	}
	return entries
}

// validateIPList validates that every entry is either a valid IP address or a
// valid CIDR range. name labels the list for error messages.
func validateIPList(name string, entries []string) error {
	for _, entry := range entries {
		if strings.Contains(entry, "/") {
			if _, _, err := net.ParseCIDR(entry); err != nil {
				return fmt.Errorf("invalid %s CIDR range: %s", name, entry)
			}
		} else {
			if net.ParseIP(entry) == nil {
				return fmt.Errorf("invalid %s IP address: %s", name, entry)
			}
		}
	}
	return nil
}
