package js

import (
	"os"
	"strings"
	"testing"
)

func TestSlidePanelJSExists(t *testing.T) {
	if _, err := os.Stat("slide_panel.js"); os.IsNotExist(err) {
		t.Fatal("slide_panel.js does not exist")
	}
}

func TestSlidePanelJSStructure(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name     string
		pattern  string
		required bool
	}{
		{"SlidePanel class", "class SlidePanel", true},
		{"constructor", "constructor(panelSelector, options", true},
		{"init method", "init()", true},
		{"attachEventListeners method", "attachEventListeners()", true},
		{"open method", "open()", true},
		{"close method", "close()", true},
		{"cancel method", "cancel()", true},
		{"save method", "save()", true},
		{"isOpen method", "isOpen()", true},
		{"window export", "window.SlidePanel", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				if tt.required {
					t.Errorf("Required pattern '%s' not found in slide_panel.js", tt.pattern)
				}
			}
		})
	}
}

func TestSlidePanelButtonSelectors(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "closeBtn") {
		t.Error("SlidePanel should have closeBtn property")
	}

	if !strings.Contains(jsContent, "cancelBtn") {
		t.Error("SlidePanel should have cancelBtn property")
	}

	if !strings.Contains(jsContent, "saveBtn") {
		t.Error("SlidePanel should have saveBtn property")
	}

	if !strings.Contains(jsContent, "replace('-panel', '')") && !strings.Contains(jsContent, "replace(/-panel$/, '')") {
		t.Error("Button selector construction should remove '-panel' suffix from baseClass. For '.rsvp-settings-panel', buttons should be '.rsvp-settings-close' not '.rsvp-settings-panel-close'")
	}
}

func TestSlidePanelButtonSelectorLogic(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "querySelector(`.${baseClass}-close`)") ||
		strings.Contains(jsContent, "querySelector(`.${baseClass}-cancel`)") ||
		strings.Contains(jsContent, "querySelector(`.${baseClass}-save`)") {

		if !strings.Contains(jsContent, "replace('-panel', '')") &&
			!strings.Contains(jsContent, "replace(/-panel$/, '')") {
			t.Error("Button selectors use baseClass but don't remove '-panel' suffix. For '.rsvp-settings-panel', buttons should be '.rsvp-settings-close' not '.rsvp-settings-panel-close'")
		}
	}
}

func TestSlidePanelEventListeners(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name    string
		pattern string
	}{
		{"close button listener", "this.closeBtn"},
		{"cancel button listener", "this.cancelBtn"},
		{"save button listener", "this.saveBtn"},
		{"overlay listener", "this.overlay"},
		{"escape key listener", "Escape"},
		{"addEventListener", "addEventListener"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				t.Errorf("Expected pattern '%s' not found for %s", tt.pattern, tt.name)
			}
		})
	}
}

func TestSlidePanelCloseButtonClick(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "this.closeBtn.addEventListener") {
		t.Error("Close button should have click event listener")
	}

	if !strings.Contains(jsContent, "() => this.close()") && !strings.Contains(jsContent, "this.close.bind(this)") {
		t.Error("Close button should call this.close() when clicked")
	}
}

func TestSlidePanelCancelButtonClick(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "this.cancelBtn.addEventListener") {
		t.Error("Cancel button should have click event listener")
	}

	if !strings.Contains(jsContent, "() => this.cancel()") && !strings.Contains(jsContent, "this.cancel.bind(this)") {
		t.Error("Cancel button should call this.cancel() when clicked")
	}
}

func TestSlidePanelSaveButtonClick(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "this.saveBtn.addEventListener") {
		t.Error("Save button should have click event listener")
	}

	if !strings.Contains(jsContent, "() => this.save()") && !strings.Contains(jsContent, "this.save.bind(this)") {
		t.Error("Save button should call this.save() when clicked")
	}
}

func TestSlidePanelCallbacks(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	callbacks := []string{"onOpen", "onClose", "onSave", "onCancel"}

	for _, callback := range callbacks {
		t.Run(callback, func(t *testing.T) {
			if !strings.Contains(jsContent, callback) {
				t.Errorf("SlidePanel should support %s callback", callback)
			}
		})
	}
}

func TestSlidePanelOpenMethod(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "classList.add('open')") {
		t.Error("open() should add 'open' class to panel and overlay")
	}

	if !strings.Contains(jsContent, "document.body.style.overflow = 'hidden'") {
		t.Error("open() should prevent body scroll")
	}

	if !strings.Contains(jsContent, "this.options.onOpen") {
		t.Error("open() should call onOpen callback if provided")
	}
}

func TestSlidePanelCloseMethod(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "classList.remove('open')") {
		t.Error("close() should remove 'open' class from panel and overlay")
	}

	if !strings.Contains(jsContent, "document.body.style.overflow = ''") {
		t.Error("close() should restore body scroll")
	}

	if !strings.Contains(jsContent, "this.options.onClose") {
		t.Error("close() should call onClose callback if provided")
	}
}

func TestSlidePanelCancelMethod(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "this.options.onCancel") {
		t.Error("cancel() should call onCancel callback if provided")
	}

	if !strings.Contains(jsContent, "this.close()") {
		t.Error("cancel() should call this.close()")
	}
}

func TestSlidePanelSaveMethod(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "this.options.onSave") {
		t.Error("save() should call onSave callback if provided")
	}

	if !strings.Contains(jsContent, "this.close()") {
		t.Error("save() should call this.close()")
	}
}

func TestSlidePanelEscapeKey(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "e.key === 'Escape'") && !strings.Contains(jsContent, "e.key === \"Escape\"") {
		t.Error("SlidePanel should listen for Escape key")
	}

	if !strings.Contains(jsContent, "this.isOpen()") {
		t.Error("Escape key handler should check if panel is open before closing")
	}
}

func TestSlidePanelOverlayClick(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "this.overlay.addEventListener") {
		t.Error("Overlay should have click event listener")
	}
}

func TestSlidePanelNoConsoleLog(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "console.log") {
		t.Error("Production JavaScript should not contain console.log statements")
	}
}

func TestSlidePanelValidSyntax(t *testing.T) {
	content, err := os.ReadFile("slide_panel.js")
	if err != nil {
		t.Fatalf("Failed to read slide_panel.js: %v", err)
	}

	jsContent := string(content)

	openBraces := strings.Count(jsContent, "{")
	closeBraces := strings.Count(jsContent, "}")

	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}

	openParens := strings.Count(jsContent, "(")
	closeParens := strings.Count(jsContent, ")")

	if openParens != closeParens {
		t.Errorf("Mismatched parentheses: %d open, %d close", openParens, closeParens)
	}
}
