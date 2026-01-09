package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestColorsFileExists(t *testing.T) {
	if _, err := os.Stat("colors.css"); os.IsNotExist(err) {
		t.Fatal("colors.css file does not exist")
	}
}

func TestColorsValidCSS(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("colors.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestColorsBackgroundUtilities(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	utilities := []string{
		".bg-primary",
		".bg-success",
		".bg-warning",
		".bg-error",
		".bg-info",
		".bg-gray-50",
		".bg-gray-100",
		".bg-gray-200",
		".bg-white",
		".bg-surface",
	}

	for _, utility := range utilities {
		t.Run(utility, func(t *testing.T) {
			if !strings.Contains(cssContent, utility) {
				t.Errorf("Missing background utility class: %s", utility)
			}
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(utility) + `\s*\{[^}]*background-color[^}]*\}`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Background utility %s missing background-color property", utility)
			}
		})
	}
}

func TestColorsTextUtilities(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	utilities := []string{
		".text-primary-600",
		".text-primary-700",
		".text-success",
		".text-warning",
		".text-error",
		".text-info",
		".text-gray-600",
		".text-gray-700",
		".text-gray-800",
		".text-gray-900",
	}

	for _, utility := range utilities {
		t.Run(utility, func(t *testing.T) {
			if !strings.Contains(cssContent, utility) {
				t.Errorf("Missing text utility class: %s", utility)
			}
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(utility) + `\s*\{[^}]*color[^}]*\}`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Text utility %s missing color property", utility)
			}
		})
	}
}

func TestColorsBorderUtilities(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	utilities := []string{
		".border-gray-200",
		".border-gray-300",
		".border-primary",
		".border-success",
		".border-warning",
		".border-error",
	}

	for _, utility := range utilities {
		t.Run(utility, func(t *testing.T) {
			if !strings.Contains(cssContent, utility) {
				t.Errorf("Missing border utility class: %s", utility)
			}
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(utility) + `\s*\{[^}]*border-color[^}]*\}`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Border utility %s missing border-color property", utility)
			}
		})
	}
}

func TestColorsUsesVariables(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	requiredVars := []string{
		"--color-primary-",
		"--color-success",
		"--color-warning",
		"--color-error",
		"--color-info",
		"--color-gray-",
		"--color-background",
		"--color-surface",
	}

	for _, varName := range requiredVars {
		t.Run("uses_"+varName, func(t *testing.T) {
			if !strings.Contains(cssContent, "var("+varName) {
				t.Errorf("Colors should use CSS variable: %s", varName)
			}
		})
	}
}

func TestColorsSemanticColors(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	semanticColors := []string{
		"success",
		"warning",
		"error",
		"info",
	}

	for _, semantic := range semanticColors {
		t.Run("semantic_"+semantic, func(t *testing.T) {
			bgClass := ".bg-" + semantic
			textClass := ".text-" + semantic
			borderClass := ".border-" + semantic

			if !strings.Contains(cssContent, bgClass) {
				t.Errorf("Missing background class for semantic color: %s", bgClass)
			}
			if !strings.Contains(cssContent, textClass) {
				t.Errorf("Missing text class for semantic color: %s", textClass)
			}
			if !strings.Contains(cssContent, borderClass) {
				t.Errorf("Missing border class for semantic color: %s", borderClass)
			}
		})
	}
}

func TestColorsGrayScale(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	grayLevels := []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900"}

	for _, level := range grayLevels {
		t.Run("gray_"+level, func(t *testing.T) {
			bgClass := ".bg-gray-" + level
			if !strings.Contains(cssContent, bgClass) {
				t.Errorf("Missing background class for gray level: %s", bgClass)
			}
		})
	}

	textGrayLevels := []string{"600", "700", "800", "900"}
	for _, level := range textGrayLevels {
		t.Run("text_gray_"+level, func(t *testing.T) {
			textClass := ".text-gray-" + level
			if !strings.Contains(cssContent, textClass) {
				t.Errorf("Missing text class for gray level: %s", textClass)
			}
		})
	}
}

func TestColorsPrimaryScale(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	primaryLevels := []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900"}

	for _, level := range primaryLevels {
		t.Run("primary_"+level, func(t *testing.T) {
			bgClass := ".bg-primary-" + level
			if !strings.Contains(cssContent, bgClass) {
				t.Errorf("Missing background class for primary level: %s", bgClass)
			}
		})
	}
}

func TestColorsHoverStates(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	hoverClasses := []string{
		".bg-primary:hover",
		".bg-success:hover",
		".bg-error:hover",
	}

	for _, hoverClass := range hoverClasses {
		t.Run(hoverClass, func(t *testing.T) {
			if !strings.Contains(cssContent, hoverClass) {
				t.Errorf("Missing hover state: %s", hoverClass)
			}
		})
	}
}

func TestColorsLightVariants(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	lightVariants := []string{
		".bg-success-light",
		".bg-warning-light",
		".bg-error-light",
		".bg-info-light",
	}

	for _, variant := range lightVariants {
		t.Run(variant, func(t *testing.T) {
			if !strings.Contains(cssContent, variant) {
				t.Errorf("Missing light variant: %s", variant)
			}
		})
	}
}

func TestColorsDarkVariants(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	darkVariants := []string{
		".text-success-dark",
		".text-warning-dark",
		".text-error-dark",
	}

	for _, variant := range darkVariants {
		t.Run(variant, func(t *testing.T) {
			if !strings.Contains(cssContent, variant) {
				t.Errorf("Missing dark variant: %s", variant)
			}
		})
	}
}

func TestColorsTransparentUtilities(t *testing.T) {
	content, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".bg-transparent") {
		t.Error("Missing .bg-transparent utility")
	}

	pattern := regexp.MustCompile(`\.bg-transparent\s*\{[^}]*background-color:\s*transparent[^}]*\}`)
	if !pattern.MatchString(cssContent) {
		t.Error(".bg-transparent should set background-color to transparent")
	}
}
