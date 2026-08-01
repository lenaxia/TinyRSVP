package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
)

// renderHTML executes name against tmpl into a buffer and only writes headers
// after a successful render, so a mid-render failure produces a clean 500
// instead of a truncated response with headers already sent. This is the
// single render path for all page handlers.
func renderHTML(w http.ResponseWriter, tmpl *template.Template, name string, status int, data interface{}) {
	if tmpl == nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("template execution failed", "template", name, "error", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}
