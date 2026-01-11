package js

import (
	"os"
	"strings"
	"testing"
)

func TestThemePreviewModalJSExists(t *testing.T) {
	if _, err := os.Stat("theme_preview_modal.js"); os.IsNotExist(err) {
		t.Fatal("theme_preview_modal.js file does not exist")
	}
}

func TestThemePreviewModalJSStructure(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name     string
		pattern  string
		category string
	}{
		{"has ThemePreviewModal class", "class ThemePreviewModal", "structure"},
		{"has constructor", "constructor()", "structure"},
		{"has init method", "init()", "structure"},
		{"has open method", "open(", "functionality"},
		{"has close method", "close(", "functionality"},
		{"has loadPreview method", "loadPreview(", "functionality"},
		{"has togglePreviewTheme method", "togglePreviewTheme(", "functionality"},
		{"has selectCurrentTheme method", "selectCurrentTheme(", "functionality"},
		{"has setupFocusTrap method", "setupFocusTrap(", "accessibility"},
		{"has announce method", "announce(", "accessibility"},
		{"listens for theme-preview-requested", "theme-preview-requested", "integration"},
		{"dispatches theme-selected event", "theme-selected", "integration"},
		{"handles Escape key", "Escape", "keyboard"},
		{"handles backdrop click", "modal-backdrop", "interaction"},
		{"handles close button", "modal-close", "interaction"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				t.Errorf("theme_preview_modal.js missing %s (pattern: %s)", tt.name, tt.pattern)
			}
		})
	}
}

func TestThemePreviewModalJSFocusManagement(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	jsContent := string(content)

	focusTests := []struct {
		name    string
		pattern string
	}{
		{"stores last focused element", "lastFocusedElement"},
		{"focuses first focusable on open", "firstFocusable"},
		{"returns focus on close", "lastFocusedElement.focus()"},
		{"implements Tab key trap", "key === 'Tab'"},
		{"handles Shift+Tab", "shiftKey"},
	}

	for _, tt := range focusTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				t.Errorf("theme_preview_modal.js missing focus management: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalJSBodyScrollPrevention(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "document.body.style.overflow") {
		t.Error("theme_preview_modal.js should prevent body scroll when modal is open")
	}
}

func TestThemePreviewModalJSEventFormDataExtraction(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	jsContent := string(content)

	dataTests := []struct {
		name    string
		pattern string
	}{
		{"extracts title", "title"},
		{"extracts location", "location"},
		{"extracts start_time", "start_time"},
		{"extracts description", "description"},
		{"provides defaults", "Sample"},
	}

	for _, tt := range dataTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				t.Errorf("theme_preview_modal.js missing event data extraction: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalJSIframeUsage(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "iframe") {
		t.Error("theme_preview_modal.js should use iframe for preview isolation")
	}

	if !strings.Contains(jsContent, ".src") {
		t.Error("theme_preview_modal.js should set iframe src for preview loading")
	}
}

func TestThemePreviewModalJSThemeToggle(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	jsContent := string(content)

	toggleTests := []struct {
		name    string
		pattern string
	}{
		{"has preview theme state", "previewTheme"},
		{"toggles between light and dark", "light"},
		{"toggles between light and dark", "dark"},
		{"updates toggle button icon", "theme-icon"},
		{"reloads preview on toggle", "loadPreview"},
	}

	for _, tt := range toggleTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				t.Errorf("theme_preview_modal.js missing theme toggle: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalJSInitialization(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "DOMContentLoaded") {
		t.Error("theme_preview_modal.js should initialize on DOMContentLoaded")
	}

	if !strings.Contains(jsContent, "new ThemePreviewModal()") {
		t.Error("theme_preview_modal.js should instantiate ThemePreviewModal class")
	}
}

func TestThemePreviewModalJSScreenReaderSupport(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	jsContent := string(content)

	srTests := []struct {
		name    string
		pattern string
	}{
		{"announces modal open", "opened"},
		{"announces modal close", "closed"},
		{"uses aria-live", "aria-live"},
		{"uses role status", "role"},
	}

	for _, tt := range srTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				t.Errorf("theme_preview_modal.js missing screen reader support: %s", tt.name)
			}
		})
	}
}
