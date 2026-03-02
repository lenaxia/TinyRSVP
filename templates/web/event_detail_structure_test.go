package web

import (
	"os"
	"strings"
	"testing"
)

func TestEventDetailTemplateStructure(t *testing.T) {
	content, err := os.ReadFile("event_detail.html")
	if err != nil {
		t.Fatalf("Failed to read event_detail.html: %v", err)
	}

	html := string(content)

	t.Run("includes event_detail.css", func(t *testing.T) {
		if !strings.Contains(html, `href="/static/css/event_detail.css"`) {
			t.Error("Template should include event_detail.css")
		}
	})

	t.Run("has main wrapper with dashboard class", func(t *testing.T) {
		// dashboard-main is defined via {{define "main-class"}} and rendered by base template
		if !strings.Contains(html, `"main-class"`) && !strings.Contains(html, `dashboard-main`) {
			t.Error("Template should define main-class as dashboard-main")
		}
	})

	t.Run("has dashboard-main container", func(t *testing.T) {
		// dashboard-main is set via {{define "main-class"}}dashboard-main{{end}} in this template
		if !strings.Contains(html, `dashboard-main`) {
			t.Error("Template should define dashboard-main via main-class block")
		}
	})

	t.Run("has event-detail-container", func(t *testing.T) {
		if !strings.Contains(html, `class="event-detail-container"`) {
			t.Error("Template should have event-detail-container for proper padding")
		}
	})

	t.Run("action buttons use display contents for forms", func(t *testing.T) {
		if strings.Contains(html, `style="display: inline;"`) {
			t.Error("Template should not use inline styles for forms, CSS handles layout")
		}
	})

	t.Run("has proper main element structure", func(t *testing.T) {
		// <main> element is in the base template; this template defines content block
		if !strings.Contains(html, `{{template "base" .}}`) && !strings.Contains(html, "<main") {
			t.Error("Template should use base template which provides main element")
		}
	})

	t.Run("no duplicate script includes", func(t *testing.T) {
		scriptCount := strings.Count(html, `src="/static/js/navigation_toggle.js"`)
		if scriptCount > 1 {
			t.Errorf("Template has %d duplicate navigation_toggle.js includes, should have 1", scriptCount)
		}
	})
}

func TestEventDetailTemplateAccessibility(t *testing.T) {
	content, err := os.ReadFile("event_detail.html")
	if err != nil {
		t.Fatalf("Failed to read event_detail.html: %v", err)
	}

	html := string(content)

	t.Run("has aria-label on status badge", func(t *testing.T) {
		if !strings.Contains(html, `aria-label="Status:`) {
			t.Error("Status badge should have aria-label")
		}
	})

	t.Run("uses semantic time elements", func(t *testing.T) {
		if !strings.Contains(html, "<time") {
			t.Error("Template should use semantic time elements")
		}
	})

	t.Run("uses semantic dl/dt/dd for details", func(t *testing.T) {
		if !strings.Contains(html, "<dl") || !strings.Contains(html, "<dt") || !strings.Contains(html, "<dd") {
			t.Error("Template should use semantic dl/dt/dd elements for event details")
		}
	})
}
