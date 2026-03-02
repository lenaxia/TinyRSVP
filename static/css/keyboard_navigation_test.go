package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestKeyboardNavigationFileExists(t *testing.T) {
	if _, err := os.Stat("keyboard_navigation.css"); os.IsNotExist(err) {
		t.Fatal("keyboard_navigation.css file does not exist")
	}
}

func TestKeyboardNavigationValidCSS(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("keyboard_navigation.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestSkipLink(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".skip-link") {
		t.Error("Missing .skip-link class")
	}

	pattern := regexp.MustCompile(`\.skip-link\s*\{[^}]*position:\s*absolute`)
	if !pattern.MatchString(cssContent) {
		t.Error(".skip-link should have position: absolute")
	}

	pattern = regexp.MustCompile(`\.skip-link\s*\{[^}]*top:\s*-\d+`)
	if !pattern.MatchString(cssContent) {
		t.Error(".skip-link should be positioned off-screen with negative top value")
	}
}

func TestSkipLinkFocus(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".skip-link:focus") {
		t.Error("Missing .skip-link:focus state")
	}

	pattern := regexp.MustCompile(`\.skip-link:focus\s*\{[^}]*top:\s*0`)
	if !pattern.MatchString(cssContent) {
		t.Error(".skip-link:focus should have top: 0 to become visible")
	}
}

func TestFocusIndicators(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus") {
		t.Error("Missing focus indicators")
	}

	pattern := regexp.MustCompile(`\*:focus\s*\{[^}]*outline:\s*2px\s+solid`)
	if !pattern.MatchString(cssContent) {
		t.Error("Global focus should have 2px solid outline")
	}

	pattern = regexp.MustCompile(`\*:focus\s*\{[^}]*outline-offset:\s*2px`)
	if !pattern.MatchString(cssContent) {
		t.Error("Global focus should have outline-offset: 2px")
	}
}

func TestFocusVisible(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus-visible") {
		t.Error("Missing :focus-visible pseudo-class for better UX")
	}
}

func TestInteractiveElementsFocus(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	interactiveElements := []string{
		"a:focus",
		"button:focus",
		"input:focus",
		"textarea:focus",
		"select:focus",
	}

	for _, element := range interactiveElements {
		t.Run("focus_"+element, func(t *testing.T) {
			if !strings.Contains(cssContent, element) {
				t.Errorf("Missing focus state for %s", element)
			}
		})
	}
}

func TestNoOutlineNone(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	pattern := regexp.MustCompile(`(?m)^\s*\*?:focus\s*\{\s*[^}]*outline:\s*none`)
	if pattern.MatchString(cssContent) {
		hasFocusVisible := strings.Contains(cssContent, ":focus-visible")
		hasNotFocusVisible := strings.Contains(cssContent, ":focus:not(:focus-visible)")

		if !hasFocusVisible && !hasNotFocusVisible {
			t.Error("Should not use outline: none on :focus without providing :focus-visible alternative")
		}
	}
}

func TestKeyboardNavigationUsesVariables(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	requiredVars := []string{
		"var(--color-border-focus)",
		"var(--color-primary-",
		"var(--spacing-",
		"var(--z-index-",
	}

	for _, varPrefix := range requiredVars {
		t.Run("uses_"+varPrefix, func(t *testing.T) {
			if !strings.Contains(cssContent, varPrefix) {
				t.Errorf("Keyboard navigation should use CSS variables with prefix: %s", varPrefix)
			}
		})
	}
}

func TestKeyboardNavigationNoHardcodedColors(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	hexColorPattern := regexp.MustCompile(`#[0-9a-fA-F]{3,6}`)
	matches := hexColorPattern.FindAllString(cssContent, -1)

	for _, match := range matches {
		if match != "#fff" && match != "#ffffff" && match != "#000" && match != "#000000" {
			t.Errorf("Keyboard navigation should not use hardcoded hex colors except pure black/white, found: %s", match)
		}
	}
}

func TestFocusContrast(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "var(--color-border-focus)") {
		t.Error("Focus indicators should use --color-border-focus variable for consistent contrast")
	}
}

func TestTabIndexHandling(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if strings.Contains(cssContent, "[tabindex=\"-1\"]:focus") {
		t.Log("Handles tabindex=\"-1\" focus styling")
	}
}

func TestSkipLinkAccessibility(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".skip-link") {
		t.Error("Missing skip-link for keyboard navigation")
	}

	if !strings.Contains(cssContent, "z-index:") {
		t.Error("Skip link should have high z-index to appear above other content")
	}
}

func TestFocusWithinSupport(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if strings.Contains(cssContent, ":focus-within") {
		t.Log("Uses :focus-within for enhanced focus indication")
	}
}
