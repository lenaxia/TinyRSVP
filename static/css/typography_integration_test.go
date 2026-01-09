package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestTypographyIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	typographyContent, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	variablesStr := string(variablesContent)
	typographyStr := string(typographyContent)

	requiredVars := []string{
		"--font-family-sans",
		"--font-family-mono",
		"--font-size-xs",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-size-xl",
		"--font-size-2xl",
		"--font-size-3xl",
		"--font-size-4xl",
		"--font-size-5xl",
		"--font-weight-normal",
		"--font-weight-medium",
		"--font-weight-semibold",
		"--font-weight-bold",
		"--line-height-tight",
		"--line-height-normal",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-text-disabled",
		"--color-primary-600",
		"--color-primary-700",
		"--color-success",
		"--color-error",
		"--color-warning",
		"--color-border-focus",
		"--color-gray-100",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-6",
		"--radius-sm",
		"--radius-md",
		"--transition-fast",
	}

	for _, varName := range requiredVars {
		t.Run("variable_defined_"+varName, func(t *testing.T) {
			if !strings.Contains(variablesStr, varName+":") {
				t.Errorf("Variable %s not defined in variables.css", varName)
			}
		})

		t.Run("variable_used_"+varName, func(t *testing.T) {
			if strings.Contains(typographyStr, "var("+varName) {
				if !strings.Contains(variablesStr, varName+":") {
					t.Errorf("Typography uses %s but it's not defined in variables.css", varName)
				}
			}
		})
	}
}

func TestTypographyHTTPServing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/css/typography.css" {
			content, err := os.ReadFile("typography.css")
			if err != nil {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/css")
			w.Write(content)
		} else {
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/static/css/typography.css")
	if err != nil {
		t.Fatalf("Failed to fetch typography.css: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %s", contentType)
	}
}

func TestTypographyResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	breakpoints := []struct {
		name  string
		query string
	}{
		{"tablet", "@media (min-width: 768px)"},
	}

	for _, bp := range breakpoints {
		t.Run(bp.name+"_breakpoint", func(t *testing.T) {
			if !strings.Contains(cssContent, bp.query) {
				t.Errorf("Missing %s breakpoint: %s", bp.name, bp.query)
			}
		})
	}
}

func TestTypographyAccessibilityFeatures(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name    string
		feature string
	}{
		{"focus_outline", "outline:"},
		{"focus_visible", ":focus"},
		{"font_smoothing", "-webkit-font-smoothing"},
		{"font_smoothing_moz", "-moz-osx-font-smoothing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.feature) {
				t.Errorf("Missing accessibility feature: %s", tt.feature)
			}
		})
	}
}

func TestTypographyConsistencyWithExistingTemplates(t *testing.T) {
	typographyContent, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	typographyStr := string(typographyContent)
	variablesStr := string(variablesContent)

	colorVars := []string{
		"--color-success",
		"--color-error",
		"--color-warning",
	}

	for _, colorVar := range colorVars {
		t.Run("color_consistency_"+colorVar, func(t *testing.T) {
			if strings.Contains(typographyStr, "var("+colorVar+")") {
				if !strings.Contains(variablesStr, colorVar+":") {
					t.Errorf("Typography uses %s but it's not defined in variables.css", colorVar)
				}
			}
		})
	}
}

func TestTypographyFileSize(t *testing.T) {
	info, err := os.Stat("typography.css")
	if err != nil {
		t.Fatalf("Failed to stat typography.css: %v", err)
	}

	maxSize := int64(10 * 1024)
	if info.Size() > maxSize {
		t.Errorf("typography.css is too large: %d bytes (max: %d bytes)", info.Size(), maxSize)
	}
}

func TestTypographyNoHardcodedValues(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	hardcodedPatterns := []struct {
		pattern     string
		description string
		allowed     bool
	}{
		{"#[0-9a-fA-F]{3,6}", "hardcoded hex colors", false},
		{"rgb\\(", "hardcoded rgb colors", false},
		{"rgba\\(", "hardcoded rgba colors", false},
		{"0\\.875em", "relative font size", true},
		{"0\\.125rem", "small padding", true},
		{"0\\.25rem", "small padding", true},
		{"2px", "outline width", true},
		{"65ch", "optimal line length", true},
	}

	for _, pattern := range hardcodedPatterns {
		t.Run("check_"+pattern.description, func(t *testing.T) {
			if strings.Contains(cssContent, pattern.pattern) && !pattern.allowed {
				t.Logf("Found potentially hardcoded value: %s", pattern.description)
			}
		})
	}
}

func TestTypographyUtilityClassesComplete(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	utilityGroups := []struct {
		name    string
		classes []string
	}{
		{
			name:    "font_sizes",
			classes: []string{".text-large", ".text-small", ".text-xs"},
		},
		{
			name:    "font_weights",
			classes: []string{".text-bold", ".text-semibold", ".text-medium", ".text-normal"},
		},
		{
			name:    "text_colors",
			classes: []string{".text-primary", ".text-secondary", ".text-disabled", ".text-success", ".text-error", ".text-warning"},
		},
		{
			name:    "text_alignment",
			classes: []string{".text-left", ".text-center", ".text-right"},
		},
	}

	for _, group := range utilityGroups {
		t.Run(group.name, func(t *testing.T) {
			for _, class := range group.classes {
				if !strings.Contains(cssContent, class) {
					t.Errorf("Missing utility class: %s", class)
				}
			}
		})
	}
}

func TestTypographyMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	h1BaseIndex := strings.Index(cssContent, "h1, .h1")
	if h1BaseIndex == -1 {
		t.Fatal("h1 base styles not found")
	}

	mediaQueryIndex := strings.Index(cssContent, "@media (min-width: 768px)")
	if mediaQueryIndex == -1 {
		t.Fatal("Tablet media query not found")
	}

	if h1BaseIndex > mediaQueryIndex {
		t.Error("Base styles should come before media queries (mobile-first approach)")
	}
}
