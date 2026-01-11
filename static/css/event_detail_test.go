package css

import (
	"os"
	"strings"
	"testing"
)

func TestEventDetailCSSExists(t *testing.T) {
	_, err := os.Stat("event_detail.css")
	if err != nil {
		t.Fatalf("event_detail.css should exist: %v", err)
	}
}

func TestEventDetailCSSContainsRequiredClasses(t *testing.T) {
	content, err := os.ReadFile("event_detail.css")
	if err != nil {
		t.Fatalf("Failed to read event_detail.css: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".event-detail-container",
		".event-detail-section",
		".event-detail-list",
		".event-detail-item",
		".event-detail-actions",
		".action-buttons",
		".event-detail-metadata",
		".event-status-badge",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("event_detail.css should contain class %s", class)
		}
	}
}

func TestEventDetailCSSHasPaddingStyles(t *testing.T) {
	content, err := os.ReadFile("event_detail.css")
	if err != nil {
		t.Fatalf("Failed to read event_detail.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "padding") {
		t.Error("event_detail.css should contain padding styles")
	}
}

func TestEventDetailCSSHasGridLayout(t *testing.T) {
	content, err := os.ReadFile("event_detail.css")
	if err != nil {
		t.Fatalf("Failed to read event_detail.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "grid") {
		t.Error("event_detail.css should contain grid layout styles")
	}
}

func TestEventDetailCSSHasResponsiveDesign(t *testing.T) {
	content, err := os.ReadFile("event_detail.css")
	if err != nil {
		t.Fatalf("Failed to read event_detail.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@media") {
		t.Error("event_detail.css should contain responsive media queries")
	}
}

func TestEventDetailCSSActionButtonsLayout(t *testing.T) {
	content, err := os.ReadFile("event_detail.css")
	if err != nil {
		t.Fatalf("Failed to read event_detail.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".action-buttons") {
		t.Error("event_detail.css should contain .action-buttons class")
	}

	if !strings.Contains(cssContent, "gap") {
		t.Error("event_detail.css should use gap for button spacing")
	}
}

func TestEventDetailCSSStatusBadgeStyles(t *testing.T) {
	content, err := os.ReadFile("event_detail.css")
	if err != nil {
		t.Fatalf("Failed to read event_detail.css: %v", err)
	}

	cssContent := string(content)

	statusClasses := []string{
		".event-status-badge.draft",
		".event-status-badge.published",
		".event-status-badge.cancelled",
		".event-status-badge.archived",
	}

	for _, class := range statusClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("event_detail.css should contain status badge style %s", class)
		}
	}
}

func TestEventDetailCSSMobileFirst(t *testing.T) {
	content, err := os.ReadFile("event_detail.css")
	if err != nil {
		t.Fatalf("Failed to read event_detail.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "min-width") {
		t.Error("event_detail.css should use min-width for mobile-first responsive design")
	}
}

func TestEventDetailCSSFormGroupSpacing(t *testing.T) {
	content, err := os.ReadFile("event_detail.css")
	if err != nil {
		t.Fatalf("Failed to read event_detail.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "margin") {
		t.Error("event_detail.css should contain margin styles for spacing")
	}
}
