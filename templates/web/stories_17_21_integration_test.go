package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStory17LoadingStatesIntegration(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `href="/static/css/loading_states.css"`) {
				t.Error("Template missing loading_states.css link")
			}

			if !strings.Contains(html, `src="/static/js/loading_states.js"`) {
				t.Error("Template missing loading_states.js script")
			}
		})
	}
}

func TestStory18ErrorDisplayIntegration(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `href="/static/css/error_display.css"`) {
				t.Error("Template missing error_display.css link")
			}
		})
	}
}

func TestStory19KeyboardNavigationIntegration(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `href="/static/css/keyboard_navigation.css"`) {
				t.Error("Template missing keyboard_navigation.css link")
			}

			if !strings.Contains(html, `src="/static/js/keyboard_navigation.js"`) {
				t.Error("Template missing keyboard_navigation.js script")
			}

			if !strings.Contains(html, `class="skip-link"`) {
				t.Error("Template missing skip link")
			}

			if !strings.Contains(html, `href="#main-content"`) {
				t.Error("Template missing skip link href")
			}
		})
	}
}

func TestStory20ScreenReaderIntegration(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `src="/static/js/screen_reader.js"`) {
				t.Error("Template missing screen_reader.js script")
			}

			if !strings.Contains(html, `role="main"`) {
				t.Error("Template missing main landmark role")
			}

			if !strings.Contains(html, `aria-label="Main content"`) {
				t.Error("Template missing main content aria-label")
			}

			if !strings.Contains(html, `id="main-content"`) {
				t.Error("Template missing main-content id for skip link target")
			}
		})
	}
}

func TestStory21FocusManagementIntegration(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `href="/static/css/focus_management.css"`) {
				t.Error("Template missing focus_management.css link")
			}

			if !strings.Contains(html, `src="/static/js/focus_management.js"`) {
				t.Error("Template missing focus_management.js script")
			}
		})
	}
}

func TestAllStoriesFullIntegration(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	requiredCSS := []string{
		"/static/css/loading_states.css",
		"/static/css/error_display.css",
		"/static/css/keyboard_navigation.css",
		"/static/css/focus_management.css",
	}

	requiredJS := []string{
		"/static/js/loading_states.js",
		"/static/js/keyboard_navigation.js",
		"/static/js/screen_reader.js",
		"/static/js/focus_management.js",
	}

	requiredElements := []struct {
		name    string
		pattern string
	}{
		{"skip link", `class="skip-link"`},
		{"skip link href", `href="#main-content"`},
		{"main landmark", `role="main"`},
		{"main aria-label", `aria-label="Main content"`},
		{"main id", `id="main-content"`},
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			for _, css := range requiredCSS {
				if !strings.Contains(html, css) {
					t.Errorf("Missing CSS: %s", css)
				}
			}

			for _, js := range requiredJS {
				if !strings.Contains(html, js) {
					t.Errorf("Missing JS: %s", js)
				}
			}

			for _, elem := range requiredElements {
				if !strings.Contains(html, elem.pattern) {
					t.Errorf("Missing %s: %s", elem.name, elem.pattern)
				}
			}
		})
	}
}

func TestStoryIntegrationLoadOrder(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			cssIndex := strings.Index(html, `href="/static/css/loading_states.css"`)
			jsIndex := strings.Index(html, `src="/static/js/loading_states.js"`)

			if cssIndex == -1 || jsIndex == -1 {
				t.Fatal("Missing loading states CSS or JS")
			}

			if cssIndex > jsIndex {
				t.Error("CSS should be loaded before JS")
			}

			headEnd := strings.Index(html, "</head>")
			bodyEnd := strings.Index(html, "</body>")

			if cssIndex > headEnd {
				t.Error("CSS should be in <head>")
			}

			if jsIndex < bodyEnd-500 {
				t.Error("JS should be near end of <body>")
			}
		})
	}
}

func TestStoryIntegrationAccessibilityComplete(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			skipLinkIndex := strings.Index(html, `class="skip-link"`)
			bodyIndex := strings.Index(html, "<body")
			bodyEndTag := strings.Index(html[bodyIndex:], ">")

			if skipLinkIndex == -1 {
				t.Fatal("Missing skip link")
			}

			if skipLinkIndex > bodyIndex+bodyEndTag+100 {
				t.Error("Skip link should be immediately after <body> tag")
			}

			mainIndex := strings.Index(html, `id="main-content"`)
			if mainIndex == -1 {
				t.Error("Missing main-content id for skip link target")
			}

			if !strings.Contains(html, `role="main"`) {
				t.Error("Missing main landmark")
			}

			if !strings.Contains(html, `aria-label="Main content"`) && 
			   !strings.Contains(html, `aria-label="Main navigation"`) {
				t.Error("Missing ARIA labels for landmarks")
			}
		})
	}
}

func TestStoryIntegrationNoConflicts(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			duplicateSkipLinks := strings.Count(html, `class="skip-link"`)
			if duplicateSkipLinks > 1 {
				t.Errorf("Found %d skip links, expected 1", duplicateSkipLinks)
			}

			duplicateMainContent := strings.Count(html, `id="main-content"`)
			if duplicateMainContent > 1 {
				t.Errorf("Found %d main-content ids, expected 1", duplicateMainContent)
			}

			duplicateMainRole := strings.Count(html, `role="main"`)
			if duplicateMainRole > 1 {
				t.Errorf("Found %d main roles, expected 1", duplicateMainRole)
			}
		})
	}
}

func TestStoryIntegrationCSSFilesExist(t *testing.T) {
	cssFiles := []string{
		"../../static/css/loading_states.css",
		"../../static/css/error_display.css",
		"../../static/css/keyboard_navigation.css",
		"../../static/css/focus_management.css",
	}

	for _, file := range cssFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("CSS file does not exist: %s", file)
			}
		})
	}
}

func TestStoryIntegrationJSFilesExist(t *testing.T) {
	jsFiles := []string{
		"../../static/js/loading_states.js",
		"../../static/js/keyboard_navigation.js",
		"../../static/js/screen_reader.js",
		"../../static/js/focus_management.js",
	}

	for _, file := range jsFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("JS file does not exist: %s", file)
			}
		})
	}
}

func TestStoryIntegrationNavigationConsistency(t *testing.T) {
	templatesWithNav := []string{
		"dashboard.html",
		"event_form.html",
	}

	for _, tmpl := range templatesWithNav {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `role="navigation"`) {
				t.Error("Navigation missing role attribute")
			}

			if !strings.Contains(html, `aria-label="Main navigation"`) {
				t.Error("Navigation missing aria-label")
			}
		})
	}
}

func TestStoryIntegrationFormAccessibility(t *testing.T) {
	formTemplates := []string{
		"event_form.html",
		"rsvp_page.html",
	}

	for _, tmpl := range formTemplates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `aria-describedby`) {
				t.Error("Form missing aria-describedby for help text")
			}

			if !strings.Contains(html, `aria-label`) {
				t.Error("Form missing aria-label attributes")
			}

			if !strings.Contains(html, `required`) {
				t.Error("Form missing required attributes")
			}
		})
	}
}

func TestStoryIntegrationLoadingStateElements(t *testing.T) {
	templatesWithLoading := []string{
		"dashboard.html",
		"event_list.html",
		"invite_list.html",
		"rsvp_summary.html",
	}

	for _, tmpl := range templatesWithLoading {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `.Loading`) {
				t.Error("Template missing loading state check")
			}

			if !strings.Contains(html, `loading-spinner`) || !strings.Contains(html, `loading-state`) {
				t.Error("Template missing loading state elements")
			}

			if !strings.Contains(html, `aria-live="polite"`) && !strings.Contains(html, `aria-busy="true"`) {
				t.Error("Loading state missing ARIA attributes")
			}
		})
	}
}

func TestStoryIntegrationErrorStateElements(t *testing.T) {
	templatesWithErrors := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
	}

	for _, tmpl := range templatesWithErrors {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			if !strings.Contains(html, `.Error`) {
				t.Error("Template missing error state check")
			}

			if !strings.Contains(html, `role="alert"`) {
				t.Error("Error state missing role=alert")
			}
		})
	}
}

func TestStoryIntegrationScriptLoadOrder(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			loadingIdx := strings.Index(html, `src="/static/js/loading_states.js"`)
			keyboardIdx := strings.Index(html, `src="/static/js/keyboard_navigation.js"`)
			screenReaderIdx := strings.Index(html, `src="/static/js/screen_reader.js"`)
			focusIdx := strings.Index(html, `src="/static/js/focus_management.js"`)

			if loadingIdx == -1 || keyboardIdx == -1 || screenReaderIdx == -1 || focusIdx == -1 {
				t.Fatal("Missing required JS files")
			}

			if loadingIdx > keyboardIdx || keyboardIdx > screenReaderIdx || screenReaderIdx > focusIdx {
				t.Error("JS files not in correct order: loading_states -> keyboard_navigation -> screen_reader -> focus_management")
			}
		})
	}
}

func TestStoryIntegrationCSSLoadOrder(t *testing.T) {
	templates := []string{
		"dashboard.html",
		"event_list.html",
		"event_form.html",
		"invite_list.html",
		"rsvp_summary.html",
		"rsvp_page.html",
		"confirmation.html",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			content, err := os.ReadFile(tmpl)
			if err != nil {
				t.Fatalf("Failed to read template: %v", err)
			}
			html := string(content)

			variablesIdx := strings.Index(html, `href="/static/css/variables.css"`)
			loadingIdx := strings.Index(html, `href="/static/css/loading_states.css"`)

			if variablesIdx == -1 || loadingIdx == -1 {
				t.Fatal("Missing required CSS files")
			}

			if variablesIdx > loadingIdx {
				t.Error("variables.css should be loaded before loading_states.css")
			}
		})
	}
}
