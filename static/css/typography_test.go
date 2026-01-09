package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestTypographyFileExists(t *testing.T) {
	if _, err := os.Stat("typography.css"); os.IsNotExist(err) {
		t.Fatal("typography.css file does not exist")
	}
}

func TestTypographyValidCSS(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("typography.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestTypographyBaseStyles(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		selector string
		property string
	}{
		{"body font-family", "body", "font-family"},
		{"body font-size", "body", "font-size"},
		{"body line-height", "body", "line-height"},
		{"body color", "body", "color"},
		{"body font-smoothing", "body", "-webkit-font-smoothing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(tt.selector) + `\s*\{[^}]*` + regexp.QuoteMeta(tt.property) + `[^}]*\}`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Missing %s property in %s selector", tt.property, tt.selector)
			}
		})
	}
}

func TestTypographyHeadingStyles(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	headings := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
	properties := []string{"font-size", "font-weight", "line-height"}

	for _, heading := range headings {
		for _, prop := range properties {
			t.Run(heading+"_"+prop, func(t *testing.T) {
				pattern := regexp.MustCompile(`(?s)` + heading + `[^{]*\{[^}]*` + prop + `[^}]*\}`)
				if !pattern.MatchString(cssContent) {
					t.Errorf("Missing %s property for %s", prop, heading)
				}
			})
		}
	}
}

func TestTypographyHeadingClasses(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	classes := []string{".h1", ".h2", ".h3", ".h4", ".h5", ".h6"}

	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			if !strings.Contains(cssContent, class) {
				t.Errorf("Missing heading class: %s", class)
			}
		})
	}
}

func TestTypographyResponsiveHeadings(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@media (min-width: 768px)") {
		t.Error("Missing tablet breakpoint media query")
	}

	pattern := regexp.MustCompile(`@media\s*\(min-width:\s*768px\)[^}]*\{[^}]*h1[^}]*font-size`)
	if !pattern.MatchString(cssContent) {
		t.Error("Missing responsive h1 font-size at tablet breakpoint")
	}
}

func TestTypographyParagraphStyles(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		selector string
		property string
	}{
		{"paragraph margin", "p", "margin-bottom"},
		{"paragraph max-width", "p", "max-width"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(tt.selector) + `\s*\{[^}]*` + regexp.QuoteMeta(tt.property) + `[^}]*\}`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Missing %s property in %s selector", tt.property, tt.selector)
			}
		})
	}
}

func TestTypographyTextUtilities(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	utilities := []string{
		".text-large",
		".text-small",
		".text-xs",
		".text-bold",
		".text-semibold",
		".text-medium",
		".text-normal",
		".text-primary",
		".text-secondary",
		".text-disabled",
		".text-success",
		".text-error",
		".text-warning",
		".text-left",
		".text-center",
		".text-right",
	}

	for _, utility := range utilities {
		t.Run(utility, func(t *testing.T) {
			if !strings.Contains(cssContent, utility) {
				t.Errorf("Missing utility class: %s", utility)
			}
		})
	}
}

func TestTypographyLinkStyles(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		selector string
		property string
	}{
		{"link color", "a", "color"},
		{"link text-decoration", "a", "text-decoration"},
		{"link hover", "a:hover", "color"},
		{"link focus outline", "a:focus", "outline"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(tt.selector) + `\s*\{[^}]*` + regexp.QuoteMeta(tt.property) + `[^}]*\}`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Missing %s property in %s selector", tt.property, tt.selector)
			}
		})
	}
}

func TestTypographyListStyles(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		selector string
		property string
	}{
		{"ul margin", "ul", "margin-bottom"},
		{"ol margin", "ol", "margin-bottom"},
		{"ul padding", "ul", "padding-left"},
		{"ol padding", "ol", "padding-left"},
		{"li margin", "li", "margin-bottom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(tt.selector) + `[^{]*\{[^}]*` + regexp.QuoteMeta(tt.property) + `[^}]*\}`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Missing %s property in %s selector", tt.property, tt.selector)
			}
		})
	}
}

func TestTypographyCodeStyles(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		selector string
		property string
	}{
		{"code font-family", "code", "font-family"},
		{"code font-size", "code", "font-size"},
		{"code background", "code", "background-color"},
		{"code padding", "code", "padding"},
		{"pre font-family", "pre", "font-family"},
		{"pre background", "pre", "background-color"},
		{"pre padding", "pre", "padding"},
		{"pre overflow", "pre", "overflow-x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(tt.selector) + `\s*\{[^}]*` + regexp.QuoteMeta(tt.property) + `[^}]*\}`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Missing %s property in %s selector", tt.property, tt.selector)
			}
		})
	}
}

func TestTypographyUsesVariables(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	requiredVars := []string{
		"--font-family-sans",
		"--font-family-mono",
		"--font-size-base",
		"--font-weight-bold",
		"--font-weight-semibold",
		"--font-weight-medium",
		"--line-height-tight",
		"--line-height-normal",
		"--color-text-primary",
		"--color-primary-600",
		"--spacing-",
	}

	for _, varName := range requiredVars {
		t.Run("uses_"+varName, func(t *testing.T) {
			if !strings.Contains(cssContent, "var("+varName) {
				t.Errorf("Typography should use CSS variable: %s", varName)
			}
		})
	}
}

func TestTypographyReadability(t *testing.T) {
	content, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "max-width") {
		t.Error("Typography should include max-width for optimal readability (45-75 characters)")
	}

	pattern := regexp.MustCompile(`max-width:\s*\d+ch`)
	if !pattern.MatchString(cssContent) {
		t.Error("Typography should use 'ch' unit for max-width to ensure optimal line length")
	}
}
