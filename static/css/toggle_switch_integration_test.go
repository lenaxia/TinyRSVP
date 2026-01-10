package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestToggleSwitchCSSServing(t *testing.T) {
	handler := http.FileServer(http.Dir("."))
	
	req := httptest.NewRequest(http.MethodGet, "/toggle_switch.css", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/css") {
		t.Errorf("Expected Content-Type to contain text/css, got %s", contentType)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, ".toggle-switch") {
		t.Error("Response should contain .toggle-switch class")
	}
}

func TestToggleSwitchCSSIntegrationWithVariables(t *testing.T) {
	toggleContent, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}
	
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}
	
	toggleCSS := string(toggleContent)
	variablesCSS := string(variablesContent)
	
	usedVars := []string{
		"--spacing-3",
		"--color-border",
		"--radius-full",
		"--color-background",
		"--color-primary-600",
		"--color-border-focus",
		"--color-text-primary",
		"--font-size-base",
	}
	
	for _, varName := range usedVars {
		if strings.Contains(toggleCSS, "var("+varName+")") {
			if !strings.Contains(variablesCSS, varName+":") {
				t.Errorf("toggle_switch.css uses %s but it's not defined in variables.css", varName)
			}
		}
	}
}

func TestToggleSwitchCSSNoConflicts(t *testing.T) {
	toggleContent, err := os.ReadFile("toggle_switch.css")
	if err != nil {
		t.Fatalf("Failed to read toggle_switch.css: %v", err)
	}
	
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Skip("forms.css not found, skipping conflict test")
	}
	
	toggleCSS := string(toggleContent)
	formsCSS := string(formsContent)
	
	if strings.Contains(formsCSS, ".toggle-switch") {
		t.Error("toggle_switch.css classes conflict with forms.css")
	}
	
	if strings.Contains(formsCSS, ".toggle-input") {
		t.Error("toggle_switch.css classes conflict with forms.css")
	}
	
	if strings.Contains(toggleCSS, "input[type=\"checkbox\"]") && strings.Contains(formsCSS, "input[type=\"checkbox\"]") {
		t.Log("Warning: Both files style checkbox inputs, ensure specificity is correct")
	}
}
