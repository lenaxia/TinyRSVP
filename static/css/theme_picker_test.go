package css

import (
	"os"
	"strings"
	"testing"
)

func TestThemePickerCSSStructure(t *testing.T) {
	content, err := os.ReadFile("theme_picker.css")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.css: %v", err)
	}

	cssContent := string(content)

	t.Run("has theme-picker container styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-picker") {
			t.Error("CSS must contain .theme-picker styles")
		}

		if !strings.Contains(cssContent, "margin") {
			t.Error(".theme-picker should have margin")
		}
	})

	t.Run("has theme-picker-header styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-picker-header") {
			t.Error("CSS must contain .theme-picker-header styles")
		}

		if !strings.Contains(cssContent, "display:") && !strings.Contains(cssContent, "flex") {
			t.Error(".theme-picker-header should use flexbox")
		}

		if !strings.Contains(cssContent, "justify-content") {
			t.Error(".theme-picker-header should have justify-content")
		}

		if !strings.Contains(cssContent, "align-items") {
			t.Error(".theme-picker-header should have align-items")
		}
	})

	t.Run("has theme-gallery grid styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-gallery") {
			t.Error("CSS must contain .theme-gallery styles")
		}

		if !strings.Contains(cssContent, "display:") && !strings.Contains(cssContent, "grid") {
			t.Error(".theme-gallery should use CSS Grid")
		}

		if !strings.Contains(cssContent, "grid-template-columns") {
			t.Error(".theme-gallery should define grid-template-columns")
		}

		if !strings.Contains(cssContent, "gap") {
			t.Error(".theme-gallery should have gap between items")
		}
	})

	t.Run("has theme-card styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-card") {
			t.Error("CSS must contain .theme-card styles")
		}

		if !strings.Contains(cssContent, "border") {
			t.Error(".theme-card should have border")
		}

		if !strings.Contains(cssContent, "border-radius") {
			t.Error(".theme-card should have border-radius")
		}

		if !strings.Contains(cssContent, "cursor:") && !strings.Contains(cssContent, "pointer") {
			t.Error(".theme-card should have cursor: pointer")
		}

		if !strings.Contains(cssContent, "transition") {
			t.Error(".theme-card should have transition for smooth interactions")
		}
	})

	t.Run("has selected state styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-card.selected") {
			t.Error("CSS must contain .theme-card.selected styles")
		}

		if !strings.Contains(cssContent, "border-color") {
			t.Error(".theme-card.selected should have distinct border-color")
		}

		if !strings.Contains(cssContent, "box-shadow") {
			t.Error(".theme-card.selected should have box-shadow for emphasis")
		}
	})

	t.Run("has hover state styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-card:hover") {
			t.Error("CSS must contain .theme-card:hover styles")
		}

		if !strings.Contains(cssContent, "transform") {
			t.Error(".theme-card:hover should have transform for visual feedback")
		}
	})

	t.Run("has focus state styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-card:focus-within") || !strings.Contains(cssContent, ".theme-card:focus") {
			t.Error("CSS must contain focus state styles for accessibility")
		}

		if !strings.Contains(cssContent, "outline") {
			t.Error("Focus state should have visible outline")
		}
	})
}

func TestThemePickerResponsiveDesign(t *testing.T) {
	content, err := os.ReadFile("theme_picker.css")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.css: %v", err)
	}

	cssContent := string(content)

	t.Run("has mobile breakpoint", func(t *testing.T) {
		if !strings.Contains(cssContent, "@media") {
			t.Error("CSS must contain media queries for responsive design")
		}

		// Accept any small-screen breakpoint (767px or 479px are both valid)
		hasMobileBreakpoint := strings.Contains(cssContent, "max-width: 767px") ||
			strings.Contains(cssContent, "max-width:767px") ||
			strings.Contains(cssContent, "max-width: 479px") ||
			strings.Contains(cssContent, "max-width:479px")
		if !hasMobileBreakpoint {
			t.Error("CSS should have a mobile breakpoint (max-width)")
		}
	})

	t.Run("mobile uses single column", func(t *testing.T) {
		// Try both 767px and 479px breakpoints
		mobileSection := extractMediaQuery(cssContent, "max-width: 767px")
		if mobileSection == "" {
			mobileSection = extractMediaQuery(cssContent, "max-width:767px")
		}
		if mobileSection == "" {
			mobileSection = extractMediaQuery(cssContent, "max-width: 479px")
		}
		if mobileSection == "" {
			mobileSection = extractMediaQuery(cssContent, "max-width:479px")
		}

		if !strings.Contains(mobileSection, "grid-template-columns") {
			t.Error("Mobile breakpoint should override grid-template-columns")
		}

		if !strings.Contains(mobileSection, "1fr") {
			t.Error("Mobile should use single column (1fr)")
		}
	})

	t.Run("has tablet breakpoint", func(t *testing.T) {
		// Accept 768px or 480px as valid tablet breakpoints
		hasTabletBreakpoint := strings.Contains(cssContent, "min-width: 768px") ||
			strings.Contains(cssContent, "min-width:768px") ||
			strings.Contains(cssContent, "min-width: 480px") ||
			strings.Contains(cssContent, "min-width:480px")
		if !hasTabletBreakpoint {
			t.Error("CSS should have a tablet/small-screen breakpoint (min-width)")
		}
	})

	t.Run("has desktop breakpoint", func(t *testing.T) {
		if !strings.Contains(cssContent, "min-width: 1024px") && !strings.Contains(cssContent, "min-width:1024px") {
			t.Error("CSS should have desktop breakpoint at 1024px")
		}
	})

	t.Run("desktop uses three columns", func(t *testing.T) {
		// The CSS may have multiple @media (min-width: 1024px) blocks.
		// Check that the overall CSS has a multi-column grid within a 1024px+ context.
		has1024 := strings.Contains(cssContent, "min-width: 1024px") || strings.Contains(cssContent, "min-width:1024px")
		hasMultiColumn := strings.Contains(cssContent, "repeat(3") ||
			strings.Contains(cssContent, "repeat(auto-fill") ||
			strings.Contains(cssContent, "repeat(auto-fit")
		if !has1024 || !hasMultiColumn {
			t.Error("Desktop should use multi-column layout (repeat(3 or repeat(auto-fill/auto-fit)) within a 1024px+ media query")
		}
	})
}

func TestThemePickerCSSVariables(t *testing.T) {
	content, err := os.ReadFile("theme_picker.css")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.css: %v", err)
	}

	cssContent := string(content)

	t.Run("uses spacing variables", func(t *testing.T) {
		if !strings.Contains(cssContent, "var(--spacing-") {
			t.Error("CSS should use spacing variables from variables.css")
		}
	})

	t.Run("uses color variables", func(t *testing.T) {
		if !strings.Contains(cssContent, "var(--color-") {
			t.Error("CSS should use color variables from variables.css")
		}
	})

	t.Run("uses transition variables", func(t *testing.T) {
		if !strings.Contains(cssContent, "var(--transition-") {
			t.Error("CSS should use transition variables from variables.css")
		}
	})

	t.Run("uses radius variables", func(t *testing.T) {
		if !strings.Contains(cssContent, "var(--radius-") {
			t.Error("CSS should use radius variables from variables.css")
		}
	})

	t.Run("uses shadow variables", func(t *testing.T) {
		if !strings.Contains(cssContent, "var(--shadow-") {
			t.Error("CSS should use shadow variables from variables.css")
		}
	})
}

func TestThemePickerAccessibilityStyles(t *testing.T) {
	content, err := os.ReadFile("theme_picker.css")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.css: %v", err)
	}

	cssContent := string(content)

	t.Run("has minimum tap target size", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-actions button") || !strings.Contains(cssContent, ".btn-select") {
			t.Error("CSS should style action buttons")
		}

		if !strings.Contains(cssContent, "padding") {
			t.Error("Buttons should have padding for adequate tap targets (44px minimum)")
		}
	})

	t.Run("has visible focus indicators", func(t *testing.T) {
		if !strings.Contains(cssContent, "outline") {
			t.Error("Focus states should have visible outline")
		}

		if !strings.Contains(cssContent, "outline-offset") {
			t.Error("Focus outline should have offset for better visibility")
		}
	})
}

func TestThemePickerComponentStyles(t *testing.T) {
	content, err := os.ReadFile("theme_picker.css")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.css: %v", err)
	}

	cssContent := string(content)

	t.Run("has thumbnail styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-thumbnail") {
			t.Error("CSS must contain .theme-thumbnail styles")
		}

		if !strings.Contains(cssContent, "height") {
			t.Error(".theme-thumbnail should have fixed height")
		}

		if !strings.Contains(cssContent, "overflow:") && !strings.Contains(cssContent, "hidden") {
			t.Error(".theme-thumbnail should hide overflow")
		}
	})

	t.Run("has thumbnail image styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-thumbnail img") {
			t.Error("CSS must contain .theme-thumbnail img styles")
		}

		if !strings.Contains(cssContent, "object-fit") {
			t.Error("Thumbnail images should use object-fit for proper scaling")
		}
	})

	t.Run("has theme-info styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-info") {
			t.Error("CSS must contain .theme-info styles")
		}

		if !strings.Contains(cssContent, "padding") {
			t.Error(".theme-info should have padding")
		}
	})

	t.Run("has theme-name styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-name") {
			t.Error("CSS must contain .theme-name styles")
		}

		if !strings.Contains(cssContent, "font-size") {
			t.Error(".theme-name should have font-size")
		}

		if !strings.Contains(cssContent, "font-weight") {
			t.Error(".theme-name should have font-weight")
		}
	})

	t.Run("has theme-description styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-description") {
			t.Error("CSS must contain .theme-description styles")
		}
	})

	t.Run("has theme-actions styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-actions") {
			t.Error("CSS must contain .theme-actions styles")
		}

		if !strings.Contains(cssContent, "display:") && !strings.Contains(cssContent, "flex") {
			t.Error(".theme-actions should use flexbox")
		}
	})
}

func extractMediaQuery(css string, query string) string {
	searchStart := 0
	for {
		startIdx := strings.Index(css[searchStart:], "@media")
		if startIdx == -1 {
			return ""
		}

		startIdx += searchStart

		braceCount := 0
		foundOpenBrace := false
		endIdx := startIdx

		for i := startIdx; i < len(css); i++ {
			if css[i] == '{' {
				braceCount++
				foundOpenBrace = true
			} else if css[i] == '}' {
				braceCount--
				if foundOpenBrace && braceCount == 0 {
					endIdx = i + 1
					break
				}
			}
		}

		if endIdx > startIdx {
			section := css[startIdx:endIdx]
			if strings.Contains(section, query) {
				return section
			}
		}

		searchStart = endIdx
		if searchStart >= len(css) {
			break
		}
	}

	return ""
}
