package web

import (
	"bytes"
	"html/template"
	"os"
	"strings"
	"testing"
)

type navTestData struct {
	ActivePage string
	IsAdmin    bool
}

func parseNavTemplate(t *testing.T) *template.Template {
	t.Helper()
	raw, err := os.ReadFile("navigation.html")
	if err != nil {
		t.Fatalf("Failed to read navigation.html: %v", err)
	}
	tmpl, err := template.New("navigation").Parse(string(raw))
	if err != nil {
		t.Fatalf("Failed to parse navigation.html: %v", err)
	}
	return tmpl
}

func TestNavigationAdminLinkGatedByRole(t *testing.T) {
	tmpl := parseNavTemplate(t)

	t.Run("admin link shown for admin user", func(t *testing.T) {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, navTestData{ActivePage: "dashboard", IsAdmin: true}); err != nil {
			t.Fatalf("template execute failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `href="/admin"`) {
			t.Error("expected Admin nav link to be present when IsAdmin=true")
		}
	})

	t.Run("admin link hidden for non-admin user", func(t *testing.T) {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, navTestData{ActivePage: "dashboard", IsAdmin: false}); err != nil {
			t.Fatalf("template execute failed: %v", err)
		}
		out := buf.String()
		if strings.Contains(out, `href="/admin"`) {
			t.Error("expected Admin nav link to be absent when IsAdmin=false")
		}
	})
}

func TestNavigationAlwaysShowsDashboardAndEvents(t *testing.T) {
	tmpl := parseNavTemplate(t)

	for _, isAdmin := range []bool{true, false} {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, navTestData{ActivePage: "dashboard", IsAdmin: isAdmin}); err != nil {
			t.Fatalf("template execute failed (isAdmin=%v): %v", isAdmin, err)
		}
		out := buf.String()
		if !strings.Contains(out, `href="/"`) {
			t.Errorf("Dashboard link missing (isAdmin=%v)", isAdmin)
		}
		if !strings.Contains(out, `href="/events"`) {
			t.Errorf("Events link missing (isAdmin=%v)", isAdmin)
		}
	}
}
