package web

import (
	"bytes"
	"strings"
	"testing"
)

// These tests exercise the reusable partials defined in
// templates/web/partials/components.html individually so we can iterate on
// them without touching every consumer page.

func renderPartial(t *testing.T, name string, data map[string]interface{}) string {
	t.Helper()
	tmpl, err := parsePartialsOnly()
	if err != nil {
		t.Fatalf("parse partials: %v", err)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		t.Fatalf("execute %q: %v", name, err)
	}
	return buf.String()
}

func TestStatsCard_MinimalRender(t *testing.T) {
	out := renderPartial(t, "stats-card", map[string]interface{}{
		"Title": "Total Users",
		"Value": 42,
	})

	for _, want := range []string{
		`class="stats-card`,
		`Total Users`,
		`42`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stats-card output missing %q\ngot: %s", want, out)
		}
	}
}

func TestStatsCard_WithSubtitleAndAccent(t *testing.T) {
	out := renderPartial(t, "stats-card", map[string]interface{}{
		"Title":    "Events",
		"Value":    7,
		"Subtitle": "2 draft, 5 published",
		"Accent":   "success",
	})

	if !strings.Contains(out, "stats-card-accent-success") {
		t.Errorf("expected accent class 'stats-card-accent-success', got:\n%s", out)
	}
	if !strings.Contains(out, "2 draft, 5 published") {
		t.Errorf("expected subtitle text, got:\n%s", out)
	}
}

func TestStatsCard_WithHref_RendersAsLink(t *testing.T) {
	out := renderPartial(t, "stats-card", map[string]interface{}{
		"Title": "Invites",
		"Value": 12,
		"Href":  "/admin/invites",
	})

	if !strings.Contains(out, `<a `) {
		t.Errorf("expected <a> wrapper when Href set, got:\n%s", out)
	}
	if !strings.Contains(out, `href="/admin/invites"`) {
		t.Errorf("expected href attribute, got:\n%s", out)
	}
}

func TestStatsCard_WithoutHref_RendersAsDiv(t *testing.T) {
	out := renderPartial(t, "stats-card", map[string]interface{}{
		"Title": "Invites",
		"Value": 12,
	})

	// Should not be a link — should be a plain div.
	if strings.Contains(out, `<a `) && strings.Contains(out, `stats-card`) {
		t.Errorf("expected <div>, not <a>, when Href absent, got:\n%s", out)
	}
}

func TestStatsCard_EscapesUntrustedInput(t *testing.T) {
	out := renderPartial(t, "stats-card", map[string]interface{}{
		"Title": "<script>alert(1)</script>",
		"Value": "<img onerror=x>",
	})

	if strings.Contains(out, "<script>") {
		t.Errorf("stats-card must escape HTML in Title, got:\n%s", out)
	}
	if strings.Contains(out, "<img onerror") {
		t.Errorf("stats-card must escape HTML in Value, got:\n%s", out)
	}
}

func TestSectionPartial_RendersTitleAndBody(t *testing.T) {
	out := renderPartial(t, "section", map[string]interface{}{
		"Title": "Database Connection Pool",
		"Body":  "<p>Body content</p>",
	})

	for _, want := range []string{
		`class="ui-section`,
		`class="ui-section-title"`,
		`Database Connection Pool`,
		`class="ui-section-body"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("section partial missing %q\ngot: %s", want, out)
		}
	}
}

func TestSectionPartial_OptionalDescription(t *testing.T) {
	out := renderPartial(t, "section", map[string]interface{}{
		"Title":       "Server",
		"Description": "Runtime settings loaded from environment at startup.",
		"Body":        "<dl></dl>",
	})

	if !strings.Contains(out, "Runtime settings loaded") {
		t.Errorf("expected description to render, got:\n%s", out)
	}
	if !strings.Contains(out, "ui-section-description") {
		t.Errorf("expected description class, got:\n%s", out)
	}
}

func TestSectionPartial_TitleEscaped(t *testing.T) {
	out := renderPartial(t, "section", map[string]interface{}{
		"Title": "<b>Not HTML</b>",
		"Body":  "safe",
	})
	if strings.Contains(out, "<b>Not HTML</b>") {
		t.Errorf("section title must be escaped, got:\n%s", out)
	}
}

func TestActionCard_RendersRequiredFields(t *testing.T) {
	out := renderPartial(t, "action-card", map[string]interface{}{
		"Href":        "/admin/users",
		"Title":       "Manage Users",
		"Description": "View and manage system users",
	})

	for _, want := range []string{
		`class="action-card"`,
		`href="/admin/users"`,
		`Manage Users`,
		`View and manage system users`,
		`class="action-card-title"`,
		`class="action-card-description"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("action-card missing %q\ngot: %s", want, out)
		}
	}
}

func TestActionCard_WithIcon(t *testing.T) {
	out := renderPartial(t, "action-card", map[string]interface{}{
		"Href":  "/admin/settings",
		"Icon":  "⚙",
		"Title": "System Settings",
	})

	if !strings.Contains(out, `class="action-card-icon"`) {
		t.Errorf("expected action-card-icon class when Icon set, got:\n%s", out)
	}
	if !strings.Contains(out, "⚙") {
		t.Errorf("expected icon char to render, got:\n%s", out)
	}
}

func TestActionCard_EscapesTitle(t *testing.T) {
	out := renderPartial(t, "action-card", map[string]interface{}{
		"Href":  "/x",
		"Title": "<script>alert(1)</script>",
	})
	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Errorf("action-card must escape Title, got:\n%s", out)
	}
}

func TestStatusBadge_RendersVariants(t *testing.T) {
	cases := []struct {
		variant string
		wantCls string
	}{
		{"success", "status-badge-success"},
		{"warning", "status-badge-warning"},
		{"error", "status-badge-error"},
		{"info", "status-badge-info"},
		{"neutral", "status-badge-neutral"},
	}

	for _, tc := range cases {
		t.Run(tc.variant, func(t *testing.T) {
			out := renderPartial(t, "status-badge", map[string]interface{}{
				"Variant": tc.variant,
				"Label":   "Healthy",
			})
			if !strings.Contains(out, tc.wantCls) {
				t.Errorf("status-badge %s missing class %s\ngot: %s", tc.variant, tc.wantCls, out)
			}
			if !strings.Contains(out, "Healthy") {
				t.Errorf("status-badge missing label text, got: %s", out)
			}
		})
	}
}

func TestStatusBadge_DefaultsToNeutral(t *testing.T) {
	out := renderPartial(t, "status-badge", map[string]interface{}{
		"Label": "Unknown",
	})
	if !strings.Contains(out, "status-badge-neutral") {
		t.Errorf("expected default variant to be neutral, got: %s", out)
	}
}

func TestStatusBadge_EscapesLabel(t *testing.T) {
	out := renderPartial(t, "status-badge", map[string]interface{}{
		"Variant": "success",
		"Label":   "<b>bad</b>",
	})
	if strings.Contains(out, "<b>bad</b>") {
		t.Errorf("status-badge must escape Label, got: %s", out)
	}
}

func TestMetricTile_RendersValueAndLabel(t *testing.T) {
	out := renderPartial(t, "metric-tile", map[string]interface{}{
		"Label": "Total Users",
		"Value": 42,
	})

	for _, want := range []string{
		`class="metric-tile"`,
		`class="metric-tile-value"`,
		`class="metric-tile-label"`,
		`Total Users`,
		`42`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("metric-tile missing %q\ngot: %s", want, out)
		}
	}
}

func TestMetricTile_WithHref_RendersAsLink(t *testing.T) {
	out := renderPartial(t, "metric-tile", map[string]interface{}{
		"Label": "Events",
		"Value": 5,
		"Href":  "/events",
	})
	if !strings.Contains(out, `href="/events"`) {
		t.Errorf("expected href, got: %s", out)
	}
	if !strings.Contains(out, `<a `) {
		t.Errorf("expected <a> element, got: %s", out)
	}
}

func TestMetricTile_WithSub(t *testing.T) {
	out := renderPartial(t, "metric-tile", map[string]interface{}{
		"Label": "Response Rate",
		"Value": "78%",
		"Sub":   "45 of 58 responded",
	})
	if !strings.Contains(out, "45 of 58 responded") {
		t.Errorf("expected sub-text to render, got: %s", out)
	}
	if !strings.Contains(out, "metric-tile-sub") {
		t.Errorf("expected metric-tile-sub class, got: %s", out)
	}
}

func TestDefinitionList_RendersPairs(t *testing.T) {
	out := renderPartial(t, "definition-list", map[string]interface{}{
		"Rows": []map[string]interface{}{
			{"Label": "Host", "Value": "localhost"},
			{"Label": "Port", "Value": 8080},
		},
	})

	for _, want := range []string{
		`class="definition-list"`,
		`<dt>Host</dt>`,
		`<dd>localhost</dd>`,
		`<dt>Port</dt>`,
		`<dd>8080</dd>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("definition-list missing %q\ngot: %s", want, out)
		}
	}
}

func TestDefinitionList_EmptyRows(t *testing.T) {
	out := renderPartial(t, "definition-list", map[string]interface{}{
		"Rows": []map[string]interface{}{},
	})
	if !strings.Contains(out, `class="definition-list"`) {
		t.Errorf("empty definition-list should still render outer element, got: %s", out)
	}
}

func TestDefinitionList_EscapesLabels(t *testing.T) {
	out := renderPartial(t, "definition-list", map[string]interface{}{
		"Rows": []map[string]interface{}{
			{"Label": "<b>Host</b>", "Value": "<i>x</i>"},
		},
	})
	if strings.Contains(out, "<b>Host</b>") {
		t.Errorf("definition-list must escape Label, got: %s", out)
	}
	if strings.Contains(out, "<i>x</i>") {
		t.Errorf("definition-list must escape Value, got: %s", out)
	}
}
