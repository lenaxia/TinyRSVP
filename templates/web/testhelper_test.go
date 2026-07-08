package web

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// setTestAuthHeader sets the X-Test-User-ID header for test authentication bypass.
func setTestAuthHeader() chromedp.Action {
	return network.SetExtraHTTPHeaders(network.Headers{
		"X-Test-User-ID": "1",
	})
}

// testFuncMap returns the full production FuncMap for use in tests.
// This matches the funcMap defined in cmd/server/main.go.
func testFuncMap() template.FuncMap {
	return template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"until": func(count int) []int {
			result := make([]int, count)
			for i := 0; i < count; i++ {
				result[i] = i
			}
			return result
		},
		"iterate": func(count int) []int {
			result := make([]int, count)
			for i := 0; i < count; i++ {
				result[i] = i
			}
			return result
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("Monday, January 2, 2006 at 3:04 PM MST")
		},
		"formatTime": func(t time.Time) string {
			return t.Format("3:04 PM MST")
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
	}
}

// stubNavigation is a minimal navigation template stub for tests.
// The real navigation.html uses .ActivePage which may not be present in typed
// test data structs. This stub renders a minimal nav without that field access.
const stubNavigation = `{{define "navigation"}}
<a href="#main-content" class="skip-link">Skip to main content</a>
<nav class="app-nav" role="navigation" aria-label="Main navigation">
    <a href="/" class="logo">TinyRSVP</a>
    <ul class="app-nav-menu">
        <li><a href="/dashboard" class="nav-link active">Dashboard</a></li>
        <li><a href="/events" class="nav-link">Events</a></li>
        <li><a href="/invites" class="nav-link">Invites</a></li>
        <li><a href="/settings" class="nav-link">Settings</a></li>
        <li><a href="/admin/users" class="nav-link">Admin</a></li>
    </ul>
</nav>
{{end}}`

// parseWithBase parses a template file together with the base template and
// common partials, using the full production FuncMap and a stub navigation
// that does not require .ActivePage on the data object.
// templateName is the leaf template file (e.g. "dashboard.html").
func parseWithBase(templateName string) (*template.Template, error) {
	// Start with stub navigation so it's the baseline definition of "navigation".
	t, err := template.New(templateName).Funcs(testFuncMap()).Parse(stubNavigation)
	if err != nil {
		return nil, err
	}
	// ParseFiles adds base, page_header, components, and the leaf template.
	// base.html calls {{template "navigation" .}} which will use our stub
	// above since ParseFiles does not re-define "navigation" (it defines
	// "base", "css-common", etc.).
	_, err = t.ParseFiles(
		"partials/base.html",
		"partials/page_header.html",
		"partials/components.html",
		templateName,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// parsePartialsOnly parses just the partials/components.html file (plus the
// navigation stub for completeness) so tests can render individual partial
// blocks directly, without a full page. Used by partial-level unit tests.
func parsePartialsOnly() (*template.Template, error) {
	t, err := template.New("partials").Funcs(testFuncMap()).Parse(stubNavigation)
	if err != nil {
		return nil, err
	}
	_, err = t.ParseFiles(
		"partials/page_header.html",
		"partials/components.html",
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// executeTemplate executes a template with the given data and returns the HTML string.
// It calls t.Fatal if execution fails.
func executeTemplate(t *testing.T, tmpl *template.Template, data interface{}) string {
	t.Helper()
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}
	return buf.String()
}
