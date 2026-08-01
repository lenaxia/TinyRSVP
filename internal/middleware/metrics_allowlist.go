package middleware

import (
	"net"
	"net/http"
	"strings"
)

// loopbackCIDRs are the networks always permitted to scrape /metrics, even when
// METRICS_TRUSTED_IPS is unset. This keeps same-host Prometheus scraping
// working out of the box while closing public exposure.
var loopbackCIDRs = []string{
	"127.0.0.0/8",
	"::1/128",
}

// MetricsIPAllowlist returns a middleware that restricts access to the wrapped
// handler to loopback addresses and any explicitly trusted IPs/CIDRs.
//
// When trustedIPs is empty, only loopback (127.0.0.0/8, ::1) is permitted.
// Prometheus cannot authenticate via browser session cookies, so an IP
// allowlist is the conventional way to protect a scrape endpoint. Operators
// scraping through a reverse proxy or from another host set
// METRICS_TRUSTED_IPS (comma-separated IPs or CIDRs) to the scraper address.
//
// The client IP is read via GetRealIP (set by the RealIP middleware from
// X-Real-IP / X-Forwarded-For) so the allowlist works behind a proxy.
func MetricsIPAllowlist(trustedIPs []string) func(http.Handler) http.Handler {
	parsed := mustParseCIDRs(append(loopbackCIDRs, trustedIPs...))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIPForAllowlist(r)
			if ip == nil {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			for _, network := range parsed {
				if network.Contains(ip) {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

func clientIPForAllowlist(r *http.Request) net.IP {
	raw := GetRealIP(r.Context())
	if raw == "" {
		raw = r.RemoteAddr
	}
	// Strip the port for both "host:port" and "[host]:port" forms.
	if host, _, err := net.SplitHostPort(raw); err == nil {
		raw = host
	}
	return net.ParseIP(strings.TrimSpace(raw))
}

func mustParseCIDRs(entries []string) []*net.IPNet {
	networks := make([]*net.IPNet, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if !strings.Contains(entry, "/") {
			if ip := net.ParseIP(entry); ip != nil {
				if ip.To4() != nil {
					entry += "/32"
				} else {
					entry += "/128"
				}
			}
		}
		_, network, err := net.ParseCIDR(entry)
		if err != nil {
			continue
		}
		networks = append(networks, network)
	}
	return networks
}
