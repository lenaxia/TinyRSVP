package css

import (
	"os"
	"strings"
	"testing"
)

func TestToggleSwitchCSSExists(t *testing.T) {
	_, err := os.Stat("toggle_switch.css")
	if err != nil {
		t.Fatalf("toggle_switch.css file does not exist: %v", err)
	}
}

func TestToggleSwitchCSSContent(t *testing.T) {
	content, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".toggle-switch",
		".toggle-input",
		".toggle-slider",
		".toggle-label",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("toggle_switch.css missing required class: %s", class)
		}
	}
}

func TestToggleSwitchCSSHasTransitions(t *testing.T) {
	content, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "transition") {
		t.Error("toggle_switch.css should include transition for smooth animations")
	}
}

func TestToggleSwitchCSSHasAccessibility(t *testing.T) {
	content, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus") {
		t.Error("toggle_switch.css should include :focus styles for accessibility")
	}

	if !strings.Contains(cssContent, "outline") {
		t.Error("toggle_switch.css should include outline for focus visibility")
	}
}

func TestToggleSwitchCSSHasCheckedState(t *testing.T) {
	content, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":checked") {
		t.Error("toggle_switch.css should include :checked pseudo-class for toggle state")
	}
}

func TestToggleSwitchCSSHasDisabledState(t *testing.T) {
	content, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":disabled") {
		t.Error("toggle_switch.css should include :disabled state styling")
	}
}

func TestToggleSwitchCSSMinimumTouchTarget(t *testing.T) {
	content, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "44px") && !strings.Contains(cssContent, "48px") {
		t.Error("toggle_switch.css should have minimum 44px touch target for accessibility")
	}
}

func TestToggleSwitchCSSUsesVariables(t *testing.T) {
	content, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "var(--") {
		t.Error("toggle_switch.css should use CSS variables for consistency")
	}
}
