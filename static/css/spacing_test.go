package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSpacingFileExists(t *testing.T) {
	if _, err := os.Stat("spacing.css"); os.IsNotExist(err) {
		t.Fatal("spacing.css file does not exist")
	}
}

func TestSpacingValidCSS(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("spacing.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestSpacingMarginUtilities(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	spacingValues := []string{"0", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20", "24"}

	for _, value := range spacingValues {
		t.Run("m-"+value, func(t *testing.T) {
			className := ".m-" + value
			if !strings.Contains(cssContent, className) {
				t.Errorf("Missing margin utility class: %s", className)
			}
			pattern := regexp.MustCompile(`\.m-` + value + `\s*\{\s*margin:\s*var\(--spacing-` + value + `\)`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Margin utility %s should use var(--spacing-%s)", className, value)
			}
		})
	}
}

func TestSpacingMarginDirectionalUtilities(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	directions := []struct {
		prefix   string
		property string
	}{
		{"mt", "margin-top"},
		{"mr", "margin-right"},
		{"mb", "margin-bottom"},
		{"ml", "margin-left"},
		{"mx", "margin-left"},
		{"my", "margin-top"},
	}

	spacingValues := []string{"0", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20", "24"}

	for _, dir := range directions {
		for _, value := range spacingValues {
			t.Run(dir.prefix+"-"+value, func(t *testing.T) {
				className := "." + dir.prefix + "-" + value
				if !strings.Contains(cssContent, className) {
					t.Errorf("Missing directional margin utility: %s", className)
				}
				pattern := regexp.MustCompile(`\.` + dir.prefix + `-` + value + `\s*\{[^}]*` + regexp.QuoteMeta(dir.property) + `[^}]*\}`)
				if !pattern.MatchString(cssContent) {
					t.Errorf("Directional margin utility %s should have %s property", className, dir.property)
				}
			})
		}
	}
}

func TestSpacingPaddingUtilities(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	spacingValues := []string{"0", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20", "24"}

	for _, value := range spacingValues {
		t.Run("p-"+value, func(t *testing.T) {
			className := ".p-" + value
			if !strings.Contains(cssContent, className) {
				t.Errorf("Missing padding utility class: %s", className)
			}
			pattern := regexp.MustCompile(`\.p-` + value + `\s*\{\s*padding:\s*var\(--spacing-` + value + `\)`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Padding utility %s should use var(--spacing-%s)", className, value)
			}
		})
	}
}

func TestSpacingPaddingDirectionalUtilities(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	directions := []struct {
		prefix   string
		property string
	}{
		{"pt", "padding-top"},
		{"pr", "padding-right"},
		{"pb", "padding-bottom"},
		{"pl", "padding-left"},
		{"px", "padding-left"},
		{"py", "padding-top"},
	}

	spacingValues := []string{"0", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20", "24"}

	for _, dir := range directions {
		for _, value := range spacingValues {
			t.Run(dir.prefix+"-"+value, func(t *testing.T) {
				className := "." + dir.prefix + "-" + value
				if !strings.Contains(cssContent, className) {
					t.Errorf("Missing directional padding utility: %s", className)
				}
				pattern := regexp.MustCompile(`\.` + dir.prefix + `-` + value + `\s*\{[^}]*` + regexp.QuoteMeta(dir.property) + `[^}]*\}`)
				if !pattern.MatchString(cssContent) {
					t.Errorf("Directional padding utility %s should have %s property", className, dir.property)
				}
			})
		}
	}
}

func TestSpacingGapUtilities(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	spacingValues := []string{"0", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20", "24"}

	for _, value := range spacingValues {
		t.Run("gap-"+value, func(t *testing.T) {
			className := ".gap-" + value
			if !strings.Contains(cssContent, className) {
				t.Errorf("Missing gap utility class: %s", className)
			}
			pattern := regexp.MustCompile(`\.gap-` + value + `\s*\{\s*gap:\s*var\(--spacing-` + value + `\)`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Gap utility %s should use var(--spacing-%s)", className, value)
			}
		})
	}
}

func TestSpacingGapDirectionalUtilities(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	directions := []struct {
		prefix   string
		property string
	}{
		{"gap-x", "column-gap"},
		{"gap-y", "row-gap"},
	}

	spacingValues := []string{"0", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20", "24"}

	for _, dir := range directions {
		for _, value := range spacingValues {
			t.Run(dir.prefix+"-"+value, func(t *testing.T) {
				className := "." + dir.prefix + "-" + value
				if !strings.Contains(cssContent, className) {
					t.Errorf("Missing directional gap utility: %s", className)
				}
				pattern := regexp.MustCompile(`\.` + strings.ReplaceAll(dir.prefix, "-", `\-`) + `-` + value + `\s*\{[^}]*` + regexp.QuoteMeta(dir.property) + `[^}]*\}`)
				if !pattern.MatchString(cssContent) {
					t.Errorf("Directional gap utility %s should have %s property", className, dir.property)
				}
			})
		}
	}
}

func TestSpacingResponsiveUtilities(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	breakpoints := []string{"md", "lg"}

	for _, bp := range breakpoints {
		t.Run("breakpoint_"+bp, func(t *testing.T) {
			pattern := regexp.MustCompile(`@media\s*\([^)]*min-width[^)]*\)`)
			if !pattern.MatchString(cssContent) {
				t.Error("Missing responsive media queries")
			}
		})
	}

	responsiveClasses := []string{
		"md:m-4",
		"md:p-4",
		"md:gap-4",
		"lg:m-6",
		"lg:p-6",
		"lg:gap-6",
	}

	for _, class := range responsiveClasses {
		t.Run(class, func(t *testing.T) {
			escapedClass := strings.ReplaceAll(class, ":", `\:`)
			if !strings.Contains(cssContent, "."+escapedClass) {
				t.Errorf("Missing responsive utility class: %s", class)
			}
		})
	}
}

func TestSpacingUsesVariables(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "var(--spacing-") {
		t.Error("Spacing utilities should use CSS variables from variables.css")
	}

	hardcodedPattern := regexp.MustCompile(`(?:margin|padding|gap):\s*\d+(?:px|rem|em)`)
	if hardcodedPattern.MatchString(cssContent) {
		t.Error("Spacing utilities should not use hardcoded values, use CSS variables instead")
	}
}

func TestSpacingNegativeMargins(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	negativeValues := []string{"1", "2", "3", "4", "5", "6", "8", "10", "12"}

	for _, value := range negativeValues {
		t.Run("-m-"+value, func(t *testing.T) {
			className := ".-m-" + value
			if !strings.Contains(cssContent, className) {
				t.Errorf("Missing negative margin utility class: %s", className)
			}
		})
	}
}

func TestSpacingAutoMargins(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	autoClasses := []string{
		".m-auto",
		".mx-auto",
		".my-auto",
		".ml-auto",
		".mr-auto",
		".mt-auto",
		".mb-auto",
	}

	for _, class := range autoClasses {
		t.Run(class, func(t *testing.T) {
			if !strings.Contains(cssContent, class) {
				t.Errorf("Missing auto margin utility: %s", class)
			}
			pattern := regexp.MustCompile(regexp.QuoteMeta(class) + `\s*\{[^}]*:\s*auto`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Auto margin utility %s should set margin to auto", class)
			}
		})
	}
}

func TestSpacingConsistency(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	spacingValues := []string{"0", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20", "24"}

	for _, value := range spacingValues {
		marginCount := strings.Count(cssContent, ".m-"+value)
		paddingCount := strings.Count(cssContent, ".p-"+value)
		gapCount := strings.Count(cssContent, ".gap-"+value)

		if marginCount == 0 {
			t.Errorf("Missing margin utilities for spacing-%s", value)
		}
		if paddingCount == 0 {
			t.Errorf("Missing padding utilities for spacing-%s", value)
		}
		if gapCount == 0 {
			t.Errorf("Missing gap utilities for spacing-%s", value)
		}
	}
}

func TestSpacingEightPointScale(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	eightPointValues := []string{"0", "8", "16", "24"}

	for _, value := range eightPointValues {
		t.Run("8pt-scale-"+value, func(t *testing.T) {
			if !strings.Contains(cssContent, "var(--spacing-"+value+")") {
				t.Errorf("Spacing system should include 8-point scale value: %s", value)
			}
		})
	}
}
