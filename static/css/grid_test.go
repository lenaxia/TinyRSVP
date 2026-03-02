package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestGridFileExists(t *testing.T) {
	if _, err := os.Stat("grid.css"); err != nil {
		t.Fatalf("grid.css file does not exist: %v", err)
	}
}

func TestGridFileReadable(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("grid.css is empty")
	}
}

func TestGridContainerClass(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".grid") {
		t.Error("Missing .grid class")
	}

	gridPattern := regexp.MustCompile(`\.grid\s*\{[^}]*display:\s*grid`)
	if !gridPattern.MatchString(cssContent) {
		t.Error(".grid class should have display: grid")
	}
}

func TestGridColumnClasses(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name    string
		class   string
		columns string
	}{
		{"1 column", ".grid-cols-1", "1"},
		{"2 columns", ".grid-cols-2", "2"},
		{"3 columns", ".grid-cols-3", "3"},
		{"4 columns", ".grid-cols-4", "4"},
		{"5 columns", ".grid-cols-5", "5"},
		{"6 columns", ".grid-cols-6", "6"},
		{"12 columns", ".grid-cols-12", "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.class) {
				t.Errorf("Missing %s class", tt.class)
			}

			pattern := regexp.MustCompile(regexp.QuoteMeta(tt.class) + `\s*\{[^}]*grid-template-columns:\s*repeat\(` + tt.columns + `,\s*1fr\)`)
			if !pattern.MatchString(cssContent) {
				t.Errorf("%s should have grid-template-columns: repeat(%s, 1fr)", tt.class, tt.columns)
			}
		})
	}
}

func TestResponsiveGridClasses(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name       string
		breakpoint string
		class      string
	}{
		{"md 2 columns", "768px", `.md\\:grid-cols-2`},
		{"md 3 columns", "768px", `.md\\:grid-cols-3`},
		{"md 4 columns", "768px", `.md\\:grid-cols-4`},
		{"lg 2 columns", "1024px", `.lg\\:grid-cols-2`},
		{"lg 3 columns", "1024px", `.lg\\:grid-cols-3`},
		{"lg 4 columns", "1024px", `.lg\\:grid-cols-4`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mediaQuery := `@media\s*\(min-width:\s*` + regexp.QuoteMeta(tt.breakpoint) + `\)`
			if !regexp.MustCompile(mediaQuery).MatchString(cssContent) {
				t.Errorf("Missing media query for %s", tt.breakpoint)
			}

			if !regexp.MustCompile(tt.class).MatchString(cssContent) {
				t.Errorf("Missing responsive class %s", tt.class)
			}
		})
	}
}

func TestFlexboxUtilities(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		class    string
		property string
		value    string
	}{
		{"flex display", ".flex", "display", "flex"},
		{"flex column", ".flex-col", "flex-direction", "column"},
		{"flex row", ".flex-row", "flex-direction", "row"},
		{"items center", ".items-center", "align-items", "center"},
		{"items start", ".items-start", "align-items", "flex-start"},
		{"items end", ".items-end", "align-items", "flex-end"},
		{"justify center", ".justify-center", "justify-content", "center"},
		{"justify between", ".justify-between", "justify-content", "space-between"},
		{"justify around", ".justify-around", "justify-content", "space-around"},
		{"justify start", ".justify-start", "justify-content", "flex-start"},
		{"justify end", ".justify-end", "justify-content", "flex-end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.class) {
				t.Errorf("Missing %s class", tt.class)
			}

			pattern := regexp.MustCompile(regexp.QuoteMeta(tt.class) + `\s*\{[^}]*` + regexp.QuoteMeta(tt.property) + `:\s*` + regexp.QuoteMeta(tt.value))
			if !pattern.MatchString(cssContent) {
				t.Errorf("%s should have %s: %s", tt.class, tt.property, tt.value)
			}
		})
	}
}

func TestFlexWrapUtilities(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name  string
		class string
		value string
	}{
		{"flex wrap", ".flex-wrap", "wrap"},
		{"flex nowrap", ".flex-nowrap", "nowrap"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(regexp.QuoteMeta(tt.class) + `\s*\{[^}]*flex-wrap:\s*` + regexp.QuoteMeta(tt.value))
			if !pattern.MatchString(cssContent) {
				t.Errorf("%s should have flex-wrap: %s", tt.class, tt.value)
			}
		})
	}
}

func TestContainerClass(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".container") {
		t.Error("Missing .container class")
	}

	tests := []struct {
		name     string
		property string
	}{
		{"width 100%", "width:\\s*100%"},
		{"margin auto", "margin-(left|right):\\s*auto"},
		{"padding left", "padding-left:\\s*var\\(--spacing-4\\)"},
		{"padding right", "padding-right:\\s*var\\(--spacing-4\\)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(`\.container\s*\{[^}]*` + tt.property)
			if !pattern.MatchString(cssContent) {
				t.Errorf(".container should have %s", tt.name)
			}
		})
	}
}

func TestContainerResponsiveMaxWidth(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name       string
		breakpoint string
		maxWidth   string
	}{
		{"md breakpoint", "768px", "var\\(--container-md\\)"},
		{"lg breakpoint", "1024px", "var\\(--container-lg\\)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(`@media\s*\(min-width:\s*` + regexp.QuoteMeta(tt.breakpoint) + `\)[^}]*\.container\s*\{[^}]*max-width:\s*` + tt.maxWidth)
			if !pattern.MatchString(cssContent) {
				t.Errorf("Missing container max-width for %s", tt.name)
			}
		})
	}
}

func TestGridAutoFitPattern(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".grid-auto-fit") {
		t.Error("Missing .grid-auto-fit class")
	}

	pattern := regexp.MustCompile(`\.grid-auto-fit\s*\{[^}]*grid-template-columns:\s*repeat\(auto-fit,\s*minmax\([^)]+\)\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".grid-auto-fit should use auto-fit with minmax")
	}
}

func TestGridAutoFillPattern(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".grid-auto-fill") {
		t.Error("Missing .grid-auto-fill class")
	}

	pattern := regexp.MustCompile(`\.grid-auto-fill\s*\{[^}]*grid-template-columns:\s*repeat\(auto-fill,\s*minmax\([^)]+\)\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".grid-auto-fill should use auto-fill with minmax")
	}
}

func TestGridSpanClasses(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name  string
		class string
	}{
		{"span 1", ".col-span-1"},
		{"span 2", ".col-span-2"},
		{"span 3", ".col-span-3"},
		{"span 4", ".col-span-4"},
		{"span 6", ".col-span-6"},
		{"span 12", ".col-span-12"},
		{"span full", ".col-span-full"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.class) {
				t.Errorf("Missing %s class", tt.class)
			}
		})
	}
}

func TestFlexGrowShrinkClasses(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		class    string
		property string
		value    string
	}{
		{"flex 1", ".flex-1", "flex", "1 1 0%"},
		{"flex auto", ".flex-auto", "flex", "1 1 auto"},
		{"flex none", ".flex-none", "flex", "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.class) {
				t.Errorf("Missing %s class", tt.class)
			}
		})
	}
}
