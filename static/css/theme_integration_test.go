package css

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestThemeIntegration(t *testing.T) {
	t.Run("all page templates include theme CSS", func(t *testing.T) {
		templatesDir := "../../templates/web"
		
		files, err := filepath.Glob(filepath.Join(templatesDir, "*.html"))
		if err != nil {
			t.Fatalf("Failed to list template files: %v", err)
		}
		
		if len(files) == 0 {
			t.Skip("No template files found, skipping test")
		}
		
		baseContent, err := os.ReadFile(filepath.Join(templatesDir, "partials", "base.html"))
		if err != nil {
			t.Fatalf("Failed to read base.html: %v", err)
		}
		baseHTML := string(baseContent)
		
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Errorf("Failed to read %s: %v", file, err)
				continue
			}
			
			htmlContent := string(content)
			filename := filepath.Base(file)
			
			usesBase := strings.Contains(htmlContent, `{{template "base"`)
			hasVariablesCSS := strings.Contains(htmlContent, "/static/css/variables.css")
			hasThemeToggleCSS := strings.Contains(htmlContent, "/static/css/theme_toggle.css")
			
			if usesBase {
				if !strings.Contains(baseHTML, "/static/css/variables.css") {
					t.Errorf("%s uses base template, but base.html doesn't include variables.css", filename)
				}
				if !strings.Contains(baseHTML, "/static/css/theme_toggle.css") {
					t.Errorf("%s uses base template, but base.html doesn't include theme_toggle.css", filename)
				}
			} else {
				if !hasVariablesCSS {
					t.Errorf("%s should include variables.css for theme support (either directly or via base template)", filename)
				}
				if !hasThemeToggleCSS {
					t.Errorf("%s should include theme_toggle.css (either directly or via base template)", filename)
				}
			}
		}
	})
	
	t.Run("all page templates include theme JavaScript", func(t *testing.T) {
		templatesDir := "../../templates/web"
		
		files, err := filepath.Glob(filepath.Join(templatesDir, "*.html"))
		if err != nil {
			t.Fatalf("Failed to list template files: %v", err)
		}
		
		if len(files) == 0 {
			t.Skip("No template files found, skipping test")
		}
		
		baseContent, err := os.ReadFile(filepath.Join(templatesDir, "partials", "base.html"))
		if err != nil {
			t.Fatalf("Failed to read base.html: %v", err)
		}
		baseHTML := string(baseContent)
		
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Errorf("Failed to read %s: %v", file, err)
				continue
			}
			
			htmlContent := string(content)
			filename := filepath.Base(file)
			
			usesBase := strings.Contains(htmlContent, `{{template "base"`)
			hasThemeControllerJS := strings.Contains(htmlContent, "/static/js/theme_controller.js")
			
			if usesBase {
				if !strings.Contains(baseHTML, "/static/js/theme_controller.js") {
					t.Errorf("%s uses base template, but base.html doesn't include theme_controller.js", filename)
				}
			} else {
				if !hasThemeControllerJS {
					t.Errorf("%s should include theme_controller.js (either directly or via base template)", filename)
				}
			}
		}
	})
	
	t.Run("navigation template includes theme toggle button", func(t *testing.T) {
		navPath := "../../templates/web/partials/navigation.html"
		content, err := os.ReadFile(navPath)
		if err != nil {
			t.Fatalf("Failed to read navigation.html: %v", err)
		}
		
		htmlContent := string(content)
		
		if !strings.Contains(htmlContent, `id="theme-toggle"`) {
			t.Error("navigation.html should include theme toggle button with id='theme-toggle'")
		}
		
		if !strings.Contains(htmlContent, `class="theme-toggle"`) {
			t.Error("navigation.html should include theme toggle button with class='theme-toggle'")
		}
		
		if !strings.Contains(htmlContent, `class="theme-icon"`) {
			t.Error("navigation.html should include theme icon span with class='theme-icon'")
		}
		
		if !strings.Contains(htmlContent, `class="sr-only"`) {
			t.Error("navigation.html should include screen reader text with class='sr-only'")
		}
		
		if !strings.Contains(htmlContent, `aria-label`) {
			t.Error("navigation.html theme toggle should have aria-label for accessibility")
		}
	})
	
	t.Run("base template includes theme assets", func(t *testing.T) {
		basePath := "../../templates/web/partials/base.html"
		content, err := os.ReadFile(basePath)
		if err != nil {
			t.Fatalf("Failed to read base.html: %v", err)
		}
		
		htmlContent := string(content)
		
		if !strings.Contains(htmlContent, "/static/css/variables.css") {
			t.Error("base.html should include variables.css")
		}
		
		if !strings.Contains(htmlContent, "/static/css/theme_toggle.css") {
			t.Error("base.html should include theme_toggle.css")
		}
		
		if !strings.Contains(htmlContent, "/static/js/theme_controller.js") {
			t.Error("base.html should include theme_controller.js")
		}
		
		if !strings.Contains(htmlContent, `{{template "navigation"`) {
			t.Error("base.html should include navigation partial")
		}
	})
}
