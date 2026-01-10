package css

import (
	"os"
	"strings"
	"testing"
)

func TestCounterCSSExists(t *testing.T) {
	_, err := os.Stat("counter.css")
	if err != nil {
		t.Fatalf("counter.css file does not exist: %v", err)
	}
}

func TestCounterCSSContent(t *testing.T) {
	content, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".counter",
		".counter-btn",
		".counter-value",
		".counter-label",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("counter.css missing required class: %s", class)
		}
	}
}

func TestCounterCSSHasTransitions(t *testing.T) {
	content, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "transition") {
		t.Error("counter.css should include transition for smooth animations")
	}
}

func TestCounterCSSHasAccessibility(t *testing.T) {
	content, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus") {
		t.Error("counter.css should include :focus styles for accessibility")
	}

	if !strings.Contains(cssContent, "outline") {
		t.Error("counter.css should include outline for focus visibility")
	}
}

func TestCounterCSSHasDisabledState(t *testing.T) {
	content, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":disabled") {
		t.Error("counter.css should include :disabled state styling")
	}
}

func TestCounterCSSMinimumTouchTarget(t *testing.T) {
	content, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "44px") && !strings.Contains(cssContent, "48px") {
		t.Error("counter.css buttons should have minimum 44px touch target for accessibility")
	}
}

func TestCounterCSSUsesVariables(t *testing.T) {
	content, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "var(--") {
		t.Error("counter.css should use CSS variables for consistency")
	}
}

func TestCounterCSSHasHoverState(t *testing.T) {
	content, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":hover") {
		t.Error("counter.css should include :hover states for buttons")
	}
}
