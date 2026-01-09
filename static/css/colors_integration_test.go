package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestColorsIntegrationServing(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})

	req := httptest.NewRequest("GET", "/static/css/colors.css", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %s", contentType)
	}

	if w.Body.Len() == 0 {
		t.Error("Response body is empty")
	}
}

func TestColorsIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	colorsContent, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	variables := string(variablesContent)
	colors := string(colorsContent)

	requiredVars := []string{
		"--color-primary-600",
		"--color-success",
		"--color-warning",
		"--color-error",
		"--color-info",
		"--color-gray-50",
		"--color-gray-100",
		"--color-gray-200",
		"--color-surface",
		"--color-background",
	}

	for _, varName := range requiredVars {
		t.Run("variable_"+varName, func(t *testing.T) {
			if !strings.Contains(variables, varName) {
				t.Errorf("variables.css missing required variable: %s", varName)
			}

			if !strings.Contains(colors, "var("+varName) {
				t.Errorf("colors.css should reference variable: %s", varName)
			}
		})
	}
}

func TestColorsIntegrationAccessibility(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	semanticColors := []string{"success", "warning", "error", "info"}

	for _, semantic := range semanticColors {
		t.Run("semantic_"+semantic+"_complete", func(t *testing.T) {
			bgClass := ".bg-" + semantic
			textClass := ".text-" + semantic
			borderClass := ".border-" + semantic

			hasBackground := strings.Contains(cssContent, bgClass)
			hasText := strings.Contains(cssContent, textClass)
			hasBorder := strings.Contains(cssContent, borderClass)

			if !hasBackground || !hasText || !hasBorder {
				t.Errorf("Incomplete semantic color set for %s: bg=%v, text=%v, border=%v",
					semantic, hasBackground, hasText, hasBorder)
			}
		})
	}
}

func TestColorsIntegrationConsistency(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	t.Run("all_backgrounds_use_background-color", func(t *testing.T) {
		lines := strings.Split(cssContent, "\n")
		for i, line := range lines {
			if strings.Contains(line, ".bg-") && strings.Contains(line, "{") {
				found := false
				for j := i; j < len(lines) && !strings.Contains(lines[j], "}"); j++ {
					if strings.Contains(lines[j], "background-color") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Background class at line %d missing background-color property", i+1)
				}
			}
		}
	})

	t.Run("all_text_use_color", func(t *testing.T) {
		lines := strings.Split(cssContent, "\n")
		for i, line := range lines {
			if strings.Contains(line, ".text-") && strings.Contains(line, "{") && !strings.Contains(line, ":hover") {
				found := false
				for j := i; j < len(lines) && !strings.Contains(lines[j], "}"); j++ {
					if strings.Contains(lines[j], "color:") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Text class at line %d missing color property", i+1)
				}
			}
		}
	})

	t.Run("all_borders_use_border-color", func(t *testing.T) {
		lines := strings.Split(cssContent, "\n")
		for i, line := range lines {
			if strings.Contains(line, ".border-") && strings.Contains(line, "{") {
				found := false
				for j := i; j < len(lines) && !strings.Contains(lines[j], "}"); j++ {
					if strings.Contains(lines[j], "border-color") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Border class at line %d missing border-color property", i+1)
				}
			}
		}
	})
}

func TestColorsIntegrationHoverStates(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	interactiveClasses := []string{".bg-primary", ".bg-success", ".bg-error"}

	for _, class := range interactiveClasses {
		t.Run(class+"_has_hover", func(t *testing.T) {
			hoverState := class + ":hover"
			if !strings.Contains(cssContent, hoverState) {
				t.Errorf("Interactive class %s missing hover state", class)
			}
		})
	}
}

func TestColorsIntegrationGrayScaleComplete(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	grayLevels := []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900"}

	for _, level := range grayLevels {
		t.Run("gray_"+level+"_background", func(t *testing.T) {
			bgClass := ".bg-gray-" + level
			if !strings.Contains(cssContent, bgClass) {
				t.Errorf("Missing gray background class: %s", bgClass)
			}
		})
	}
}

func TestColorsIntegrationPrimaryScaleComplete(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	primaryLevels := []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900"}

	for _, level := range primaryLevels {
		t.Run("primary_"+level+"_background", func(t *testing.T) {
			bgClass := ".bg-primary-" + level
			if !strings.Contains(cssContent, bgClass) {
				t.Errorf("Missing primary background class: %s", bgClass)
			}
		})
	}
}

func TestColorsIntegrationUtilityVariants(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		base     string
		variants []string
	}{
		{
			name:     "success_variants",
			base:     "success",
			variants: []string{".bg-success", ".bg-success-light", ".text-success", ".text-success-dark", ".border-success"},
		},
		{
			name:     "warning_variants",
			base:     "warning",
			variants: []string{".bg-warning", ".bg-warning-light", ".text-warning", ".text-warning-dark", ".border-warning"},
		},
		{
			name:     "error_variants",
			base:     "error",
			variants: []string{".bg-error", ".bg-error-light", ".text-error", ".text-error-dark", ".border-error"},
		},
		{
			name:     "info_variants",
			base:     "info",
			variants: []string{".bg-info", ".bg-info-light", ".text-info", ".border-info"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, variant := range tt.variants {
				if !strings.Contains(cssContent, variant) {
					t.Errorf("Missing variant %s for %s", variant, tt.base)
				}
			}
		})
	}
}

func TestColorsIntegrationFileSize(t *testing.T) {
	info, err := os.Stat("colors.css")
	if err != nil {
		t.Fatalf("Failed to stat colors.css: %v", err)
	}

	size := info.Size()
	maxSize := int64(10 * 1024)

	if size > maxSize {
		t.Errorf("colors.css is too large: %d bytes (max %d bytes)", size, maxSize)
	}

	t.Logf("colors.css size: %d bytes", size)
}
