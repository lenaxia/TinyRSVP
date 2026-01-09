package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSpacingIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	spacingContent, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	variablesStr := string(variablesContent)
	spacingStr := string(spacingContent)

	requiredVars := []string{
		"--spacing-0",
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-5",
		"--spacing-6",
		"--spacing-8",
		"--spacing-10",
		"--spacing-12",
		"--spacing-16",
		"--spacing-20",
		"--spacing-24",
	}

	for _, varName := range requiredVars {
		t.Run("variable_defined_"+varName, func(t *testing.T) {
			if !strings.Contains(variablesStr, varName+":") {
				t.Errorf("Variable %s not defined in variables.css", varName)
			}
		})

		t.Run("variable_used_"+varName, func(t *testing.T) {
			if strings.Contains(spacingStr, "var("+varName+")") {
				if !strings.Contains(variablesStr, varName+":") {
					t.Errorf("Spacing uses %s but it's not defined in variables.css", varName)
				}
			}
		})
	}
}

func TestSpacingHTTPServing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/css/spacing.css" {
			content, err := os.ReadFile("spacing.css")
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

	resp, err := http.Get(server.URL + "/static/css/spacing.css")
	if err != nil {
		t.Fatalf("Failed to fetch spacing.css: %v", err)
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

func TestSpacingResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	breakpoints := []struct {
		name  string
		query string
	}{
		{"tablet", "@media (min-width: 768px)"},
		{"desktop", "@media (min-width: 1024px)"},
	}

	for _, bp := range breakpoints {
		t.Run(bp.name+"_breakpoint", func(t *testing.T) {
			if !strings.Contains(cssContent, bp.query) {
				t.Errorf("Missing %s breakpoint: %s", bp.name, bp.query)
			}
		})
	}
}

func TestSpacingConsistencyWithExistingCSS(t *testing.T) {
	spacingContent, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	spacingStr := string(spacingContent)
	variablesStr := string(variablesContent)

	spacingVars := []string{
		"--spacing-0",
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-5",
		"--spacing-6",
		"--spacing-8",
		"--spacing-10",
		"--spacing-12",
		"--spacing-16",
		"--spacing-20",
		"--spacing-24",
	}

	for _, spacingVar := range spacingVars {
		t.Run("spacing_consistency_"+spacingVar, func(t *testing.T) {
			if strings.Contains(spacingStr, "var("+spacingVar+")") {
				if !strings.Contains(variablesStr, spacingVar+":") {
					t.Errorf("Spacing uses %s but it's not defined in variables.css", spacingVar)
				}
			}
		})
	}
}

func TestSpacingFileSize(t *testing.T) {
	info, err := os.Stat("spacing.css")
	if err != nil {
		t.Fatalf("Failed to stat spacing.css: %v", err)
	}

	maxSize := int64(30 * 1024)
	if info.Size() > maxSize {
		t.Errorf("spacing.css is too large: %d bytes (max: %d bytes)", info.Size(), maxSize)
	}
}

func TestSpacingNoHardcodedValues(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	if strings.Contains(cssContent, ": 0px") || strings.Contains(cssContent, ": 4px") || strings.Contains(cssContent, ": 8px") {
		t.Error("Spacing utilities should not use hardcoded pixel values, use CSS variables instead")
	}

	if strings.Contains(cssContent, ": 0rem") || strings.Contains(cssContent, ": 1rem") || strings.Contains(cssContent, ": 2rem") {
		t.Error("Spacing utilities should not use hardcoded rem values, use CSS variables instead")
	}
}

func TestSpacingUtilityClassesComplete(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	utilityGroups := []struct {
		name    string
		classes []string
	}{
		{
			name:    "margin_all",
			classes: []string{".m-0", ".m-4", ".m-8", ".m-16"},
		},
		{
			name:    "margin_top",
			classes: []string{".mt-0", ".mt-4", ".mt-8", ".mt-16"},
		},
		{
			name:    "margin_bottom",
			classes: []string{".mb-0", ".mb-4", ".mb-8", ".mb-16"},
		},
		{
			name:    "margin_left",
			classes: []string{".ml-0", ".ml-4", ".ml-8", ".ml-16"},
		},
		{
			name:    "margin_right",
			classes: []string{".mr-0", ".mr-4", ".mr-8", ".mr-16"},
		},
		{
			name:    "margin_horizontal",
			classes: []string{".mx-0", ".mx-4", ".mx-8", ".mx-auto"},
		},
		{
			name:    "margin_vertical",
			classes: []string{".my-0", ".my-4", ".my-8"},
		},
		{
			name:    "padding_all",
			classes: []string{".p-0", ".p-4", ".p-8", ".p-16"},
		},
		{
			name:    "padding_top",
			classes: []string{".pt-0", ".pt-4", ".pt-8", ".pt-16"},
		},
		{
			name:    "padding_bottom",
			classes: []string{".pb-0", ".pb-4", ".pb-8", ".pb-16"},
		},
		{
			name:    "padding_left",
			classes: []string{".pl-0", ".pl-4", ".pl-8", ".pl-16"},
		},
		{
			name:    "padding_right",
			classes: []string{".pr-0", ".pr-4", ".pr-8", ".pr-16"},
		},
		{
			name:    "padding_horizontal",
			classes: []string{".px-0", ".px-4", ".px-8"},
		},
		{
			name:    "padding_vertical",
			classes: []string{".py-0", ".py-4", ".py-8"},
		},
		{
			name:    "gap",
			classes: []string{".gap-0", ".gap-4", ".gap-8", ".gap-16"},
		},
		{
			name:    "gap_horizontal",
			classes: []string{".gap-x-0", ".gap-x-4", ".gap-x-8"},
		},
		{
			name:    "gap_vertical",
			classes: []string{".gap-y-0", ".gap-y-4", ".gap-y-8"},
		},
		{
			name:    "negative_margins",
			classes: []string{".-m-1", ".-m-2", ".-m-4", ".-m-8"},
		},
		{
			name:    "auto_margins",
			classes: []string{".m-auto", ".mx-auto", ".my-auto", ".ml-auto", ".mr-auto"},
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

func TestSpacingMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	m4BaseIndex := strings.Index(cssContent, ".m-4")
	if m4BaseIndex == -1 {
		t.Fatal(".m-4 base styles not found")
	}

	mediaQueryIndex := strings.Index(cssContent, "@media (min-width: 768px)")
	if mediaQueryIndex == -1 {
		t.Fatal("Tablet media query not found")
	}

	if m4BaseIndex > mediaQueryIndex {
		t.Error("Base styles should come before media queries (mobile-first approach)")
	}
}

func TestSpacingResponsiveUtilitiesPresent(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	responsiveClasses := []struct {
		breakpoint string
		classes    []string
	}{
		{
			breakpoint: "md",
			classes:    []string{".md\\:m-4", ".md\\:p-4", ".md\\:gap-4"},
		},
		{
			breakpoint: "lg",
			classes:    []string{".lg\\:m-6", ".lg\\:p-6", ".lg\\:gap-6"},
		},
	}

	for _, rc := range responsiveClasses {
		t.Run("responsive_"+rc.breakpoint, func(t *testing.T) {
			for _, class := range rc.classes {
				if !strings.Contains(cssContent, class) {
					t.Errorf("Missing responsive utility class: %s", class)
				}
			}
		})
	}
}

func TestSpacingEightPointScaleAdherence(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	eightPointValues := []string{"0", "8", "16", "24"}

	for _, value := range eightPointValues {
		t.Run("8pt_scale_"+value, func(t *testing.T) {
			if !strings.Contains(cssContent, "var(--spacing-"+value+")") {
				t.Errorf("Spacing system should include 8-point scale value: %s", value)
			}
		})
	}
}

func TestSpacingFlexboxGridCompatibility(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	gapProperties := []string{
		"gap:",
		"column-gap:",
		"row-gap:",
	}

	for _, prop := range gapProperties {
		t.Run("gap_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf("Missing gap property for flexbox/grid: %s", prop)
			}
		})
	}
}

func TestSpacingNegativeMarginsForOverlap(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	negativeMargins := []string{".-m-1", ".-m-2", ".-m-4", ".-m-8"}

	for _, class := range negativeMargins {
		t.Run("negative_margin_"+class, func(t *testing.T) {
			if !strings.Contains(cssContent, class) {
				t.Errorf("Missing negative margin utility: %s", class)
			}
		})
	}

	if !strings.Contains(cssContent, "calc(") {
		t.Error("Negative margins should use calc() for proper calculation")
	}
}

func TestSpacingAutoMarginsForCentering(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	autoMargins := []string{".m-auto", ".mx-auto", ".my-auto", ".ml-auto", ".mr-auto", ".mt-auto", ".mb-auto"}

	for _, class := range autoMargins {
		t.Run("auto_margin_"+class, func(t *testing.T) {
			if !strings.Contains(cssContent, class) {
				t.Errorf("Missing auto margin utility: %s", class)
			}
		})
	}

	if !strings.Contains(cssContent, ": auto") {
		t.Error("Auto margin utilities should set margin to auto")
	}
}

func TestSpacingConsistencyAcrossUtilities(t *testing.T) {
	content, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	cssContent := string(content)

	spacingValues := []string{"0", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20", "24"}

	for _, value := range spacingValues {
		t.Run("consistency_"+value, func(t *testing.T) {
			hasMargin := strings.Contains(cssContent, ".m-"+value)
			hasPadding := strings.Contains(cssContent, ".p-"+value)
			hasGap := strings.Contains(cssContent, ".gap-"+value)

			if !hasMargin || !hasPadding || !hasGap {
				t.Errorf("Inconsistent spacing utilities for value %s: margin=%v, padding=%v, gap=%v",
					value, hasMargin, hasPadding, hasGap)
			}
		})
	}
}

func TestSpacingIntegrationWithTypography(t *testing.T) {
	spacingContent, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	typographyContent, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	spacingStr := string(spacingContent)
	typographyStr := string(typographyContent)

	commonSpacingVars := []string{"--spacing-2", "--spacing-3", "--spacing-4", "--spacing-6"}

	for _, varName := range commonSpacingVars {
		t.Run("shared_variable_"+varName, func(t *testing.T) {
			usedInSpacing := strings.Contains(spacingStr, "var("+varName+")")
			usedInTypography := strings.Contains(typographyStr, "var("+varName+")")

			if usedInSpacing && usedInTypography {
				t.Logf("Variable %s is shared between spacing and typography systems", varName)
			}
		})
	}
}
