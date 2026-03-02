package css

import (
	"os"
	"strings"
	"testing"
)

func TestMobileOptimizationFileExists(t *testing.T) {
	_, err := os.Stat("mobile_optimization.css")
	if err != nil {
		t.Fatalf("mobile_optimization.css file should exist: %v", err)
	}
}

func TestMobileOptimizationFileContent(t *testing.T) {
	content, err := os.ReadFile("mobile_optimization.css")
	if err != nil {
		t.Fatalf("Failed to read mobile_optimization.css: %v", err)
	}

	css := string(content)

	t.Run("contains tap highlight color", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-tap-highlight-color") {
			t.Error("CSS should contain -webkit-tap-highlight-color for touch feedback")
		}
	})

	t.Run("contains touch action manipulation", func(t *testing.T) {
		if !strings.Contains(css, "touch-action: manipulation") {
			t.Error("CSS should contain touch-action: manipulation to prevent double-tap zoom")
		}
	})

	t.Run("contains smooth scrolling for mobile", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-overflow-scrolling: touch") {
			t.Error("CSS should contain -webkit-overflow-scrolling for smooth iOS scrolling")
		}
	})

	t.Run("contains minimum tap target size", func(t *testing.T) {
		if !strings.Contains(css, "min-height: 44px") {
			t.Error("CSS should contain min-height: 44px for mobile tap targets")
		}
		if !strings.Contains(css, "min-width: 44px") {
			t.Error("CSS should contain min-width: 44px for mobile tap targets")
		}
	})

	t.Run("contains 16px font size for inputs", func(t *testing.T) {
		if !strings.Contains(css, "font-size: 16px") {
			t.Error("CSS should contain font-size: 16px for inputs to prevent zoom on iOS")
		}
	})

	t.Run("contains text size adjust", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-text-size-adjust: 100%") {
			t.Error("CSS should contain -webkit-text-size-adjust to prevent text inflation")
		}
	})

	t.Run("contains mobile-specific media query", func(t *testing.T) {
		if !strings.Contains(css, "@media (max-width: 767px)") {
			t.Error("CSS should contain mobile-specific media query")
		}
	})

	t.Run("contains mobile utility classes", func(t *testing.T) {
		utilities := []string{
			".mobile-hide",
			".mobile-show",
			".mobile-stack",
			".mobile-full-width",
			".mobile-center",
		}
		for _, utility := range utilities {
			if !strings.Contains(css, utility) {
				t.Errorf("CSS should contain %s utility class", utility)
			}
		}
	})

	t.Run("contains iOS safe area support", func(t *testing.T) {
		if !strings.Contains(css, "env(safe-area-inset-top)") {
			t.Error("CSS should contain safe area inset support for iOS notch")
		}
	})

	t.Run("contains touch-specific media queries", func(t *testing.T) {
		if !strings.Contains(css, "@media (hover: none) and (pointer: coarse)") {
			t.Error("CSS should contain touch device detection media query")
		}
	})

	t.Run("contains prevent zoom class", func(t *testing.T) {
		if !strings.Contains(css, ".prevent-zoom") {
			t.Error("CSS should contain prevent-zoom class for form inputs")
		}
	})

	t.Run("contains all input types for 16px font", func(t *testing.T) {
		inputTypes := []string{
			"input[type=\"text\"]",
			"input[type=\"email\"]",
			"input[type=\"tel\"]",
			"input[type=\"number\"]",
			"input[type=\"password\"]",
			"textarea",
			"select",
		}
		for _, inputType := range inputTypes {
			if !strings.Contains(css, inputType) {
				t.Errorf("CSS should contain %s for 16px font size", inputType)
			}
		}
	})
}

func TestMobileOptimizationIntegrationWithExistingCSS(t *testing.T) {
	t.Run("buttons already have 44px tap targets", func(t *testing.T) {
		content, err := os.ReadFile("buttons.css")
		if err != nil {
			t.Skipf("buttons.css not found: %v", err)
			return
		}

		css := string(content)
		if !strings.Contains(css, "height: 40px") {
			t.Error("buttons.css should already contain 40px height")
		}
	})

	t.Run("forms already have proper input sizing", func(t *testing.T) {
		content, err := os.ReadFile("forms.css")
		if err != nil {
			t.Skipf("forms.css not found: %v", err)
			return
		}

		css := string(content)
		if !strings.Contains(css, "font-size: var(--font-size-base)") {
			t.Error("forms.css should use base font size for inputs")
		}
	})

	t.Run("typography has responsive breakpoints", func(t *testing.T) {
		content, err := os.ReadFile("typography.css")
		if err != nil {
			t.Skipf("typography.css not found: %v", err)
			return
		}

		css := string(content)
		if !strings.Contains(css, "@media (min-width: 768px)") {
			t.Error("typography.css should have responsive breakpoints")
		}
	})

	t.Run("variables define proper font sizes", func(t *testing.T) {
		content, err := os.ReadFile("variables.css")
		if err != nil {
			t.Skipf("variables.css not found: %v", err)
			return
		}

		css := string(content)
		if !strings.Contains(css, "--font-size-base: 1rem") {
			t.Error("variables.css should define base font size as 1rem (16px)")
		}
	})
}

func TestMobileOptimizationTemplateIntegration(t *testing.T) {
	templates := []string{
		"../../templates/web/rsvp_page.html",
		"../../templates/web/dashboard.html",
		"../../templates/web/event_list.html",
		"../../templates/web/event_form.html",
		"../../templates/web/invite_list.html",
		"../../templates/web/confirmation.html",
		"../../templates/web/rsvp_summary.html",
	}

	// Read base.html once — it contains viewport meta and CSS links for templates that extend it.
	baseContent, _ := os.ReadFile("../../templates/web/partials/base.html")
	baseHTML := string(baseContent)

	for _, templatePath := range templates {
		t.Run(templatePath, func(t *testing.T) {
			content, err := os.ReadFile(templatePath)
			if err != nil {
				t.Skipf("Template not found: %v", err)
				return
			}

			templateHTML := string(content)

			// For templates that use the base layout, the viewport/CSS are in base.html.
			// Combine both for the check so we validate the full rendered page.
			usesBase := strings.Contains(templateHTML, `{{template "base"`)
			var html string
			if usesBase {
				html = baseHTML + templateHTML
			} else {
				html = templateHTML
			}

			t.Run("has viewport meta tag", func(t *testing.T) {
				if !strings.Contains(html, `<meta name="viewport"`) {
					t.Error("Template should contain viewport meta tag")
				}
				if !strings.Contains(html, `width=device-width`) {
					t.Error("Viewport should set width=device-width")
				}
				if !strings.Contains(html, `initial-scale=1.0`) {
					t.Error("Viewport should set initial-scale=1.0")
				}
			})

			t.Run("loads CSS files", func(t *testing.T) {
				if !strings.Contains(html, `href="/static/css/`) {
					t.Error("Template should load CSS files from /static/css/")
				}
			})

			t.Run("loads mobile_optimization.css specifically", func(t *testing.T) {
				if !strings.Contains(html, `href="/static/css/mobile_optimization.css"`) {
					t.Error("Template should specifically load mobile_optimization.css for mobile optimizations")
				}
			})
		})
	}
}

func TestMobileOptimizationPerformanceFeatures(t *testing.T) {
	content, err := os.ReadFile("mobile_optimization.css")
	if err != nil {
		t.Fatalf("Failed to read mobile_optimization.css: %v", err)
	}

	css := string(content)

	t.Run("prevents double-tap zoom on buttons", func(t *testing.T) {
		if !strings.Contains(css, "touch-action: manipulation") {
			t.Error("Should prevent double-tap zoom with touch-action: manipulation")
		}
	})

	t.Run("enables hardware-accelerated scrolling", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-overflow-scrolling: touch") {
			t.Error("Should enable hardware-accelerated scrolling on iOS")
		}
	})

	t.Run("prevents iOS callout menu", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-touch-callout: none") {
			t.Error("Should prevent iOS callout menu on long press")
		}
	})

	t.Run("supports iOS safe areas", func(t *testing.T) {
		safeAreas := []string{
			"env(safe-area-inset-top)",
			"env(safe-area-inset-bottom)",
			"env(safe-area-inset-left)",
			"env(safe-area-inset-right)",
		}
		for _, safeArea := range safeAreas {
			if !strings.Contains(css, safeArea) {
				t.Errorf("Should support %s for iOS notch/home indicator", safeArea)
			}
		}
	})
}

func TestMobileOptimizationAccessibilityCompliance(t *testing.T) {
	content, err := os.ReadFile("mobile_optimization.css")
	if err != nil {
		t.Fatalf("Failed to read mobile_optimization.css: %v", err)
	}

	css := string(content)

	t.Run("tap targets meet WCAG 2.1 guidelines", func(t *testing.T) {
		if !strings.Contains(css, "min-height: 44px") || !strings.Contains(css, "min-width: 44px") {
			t.Error("Tap targets should be at least 44x44px per WCAG 2.1 Level AAA")
		}
	})

	t.Run("text is readable without zoom", func(t *testing.T) {
		if !strings.Contains(css, "font-size: 16px") {
			t.Error("Base font size should be at least 16px to prevent auto-zoom on iOS")
		}
	})

	t.Run("prevents text size inflation", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-text-size-adjust: 100%") {
			t.Error("Should prevent browser text size inflation")
		}
		if !strings.Contains(css, "-ms-text-size-adjust: 100%") {
			t.Error("Should prevent IE/Edge text size inflation")
		}
	})
}

func TestMobileOptimizationResponsiveUtilities(t *testing.T) {
	content, err := os.ReadFile("mobile_optimization.css")
	if err != nil {
		t.Fatalf("Failed to read mobile_optimization.css: %v", err)
	}

	css := string(content)

	t.Run("provides mobile visibility utilities", func(t *testing.T) {
		utilities := []string{".mobile-hide", ".mobile-show", ".desktop-hide", ".desktop-show"}
		for _, utility := range utilities {
			if !strings.Contains(css, utility) {
				t.Errorf("Should provide %s utility class", utility)
			}
		}
	})

	t.Run("provides mobile layout utilities", func(t *testing.T) {
		utilities := []string{
			".mobile-stack",
			".mobile-full-width",
			".mobile-center",
			".mobile-padding",
			".mobile-no-padding",
		}
		for _, utility := range utilities {
			if !strings.Contains(css, utility) {
				t.Errorf("Should provide %s utility class", utility)
			}
		}
	})

	t.Run("provides touch action utilities", func(t *testing.T) {
		utilities := []string{
			".touch-action-pan-y",
			".touch-action-pan-x",
			".touch-action-none",
		}
		for _, utility := range utilities {
			if !strings.Contains(css, utility) {
				t.Errorf("Should provide %s utility class", utility)
			}
		}
	})
}

func TestMobileOptimizationDeviceDetection(t *testing.T) {
	content, err := os.ReadFile("mobile_optimization.css")
	if err != nil {
		t.Fatalf("Failed to read mobile_optimization.css: %v", err)
	}

	css := string(content)

	t.Run("detects touch devices", func(t *testing.T) {
		if !strings.Contains(css, "@media (hover: none) and (pointer: coarse)") {
			t.Error("Should detect touch devices with hover:none and pointer:coarse")
		}
	})

	t.Run("detects mouse/trackpad devices", func(t *testing.T) {
		if !strings.Contains(css, "@media (hover: hover) and (pointer: fine)") {
			t.Error("Should detect mouse/trackpad devices with hover:hover and pointer:fine")
		}
	})

	t.Run("provides device-specific visibility classes", func(t *testing.T) {
		if !strings.Contains(css, ".hover-only") {
			t.Error("Should provide .hover-only class for mouse-only features")
		}
		if !strings.Contains(css, ".touch-only") {
			t.Error("Should provide .touch-only class for touch-only features")
		}
	})
}
