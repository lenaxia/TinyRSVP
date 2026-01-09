package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.wrote = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func Logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			requestID := GetRequestID(r.Context())

			logger.Printf("%s %s %d %s %s",
				r.Method,
				r.URL.Path,
				ww.status,
				duration,
				requestID,
			)
		})
	}
}
