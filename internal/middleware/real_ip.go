package middleware

import (
	"context"
	"net/http"
	"strings"
)

const (
	RealIPKey contextKey = "realIP"
)

func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
			if ip != "" {
				ip = strings.TrimSpace(strings.Split(ip, ",")[0])
			}
		}
		if ip == "" {
			ip = r.RemoteAddr
		}

		ctx := context.WithValue(r.Context(), RealIPKey, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRealIP(ctx context.Context) string {
	if ip, ok := ctx.Value(RealIPKey).(string); ok {
		return ip
	}
	return ""
}
