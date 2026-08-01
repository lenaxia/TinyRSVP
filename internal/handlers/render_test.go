package handlers

import (
	"html/template"
	"testing"
)

// testTemplate returns a minimal template registered under name that renders
// the provided data. Used by tests that exercise the real render path instead
// of relying on the (now removed) nil-template fallback.
func testTemplate(t *testing.T, name string) *template.Template {
	t.Helper()
	return template.Must(template.New(name).Parse(`<html><body>{{.}}</body></html>`))
}
