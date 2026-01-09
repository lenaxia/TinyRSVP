package handlers

import (
	"net/http"
)

type MiddlewareAdapter struct {
	requireAuthFunc  func(http.Handler) http.Handler
	requireAdminFunc func(http.Handler) http.Handler
}

func NewMiddlewareAdapter(
	requireAuthFunc func(http.Handler) http.Handler,
	requireAdminFunc func(http.Handler) http.Handler,
) *MiddlewareAdapter {
	return &MiddlewareAdapter{
		requireAuthFunc:  requireAuthFunc,
		requireAdminFunc: requireAdminFunc,
	}
}

func (m *MiddlewareAdapter) RequireAuth(next http.Handler) http.Handler {
	if m.requireAuthFunc == nil {
		return next
	}
	return m.requireAuthFunc(next)
}

func (m *MiddlewareAdapter) RequireAdmin(next http.Handler) http.Handler {
	if m.requireAdminFunc == nil {
		return next
	}
	return m.requireAdminFunc(next)
}
