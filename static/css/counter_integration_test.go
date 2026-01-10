package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCounterCSSServing(t *testing.T) {
	handler := http.FileServer(http.Dir("."))
	
	req := httptest.NewRequest(http.MethodGet, "/counter.css", nil)
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
	if !strings.Contains(body, ".counter") {
		t.Error("Response should contain .counter class")
	}
}

func TestCounterCSSIntegrationWithVariables(t *testing.T) {
	counterContent, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}
	
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}
	
	counterCSS := string(counterContent)
	variablesCSS := string(variablesContent)
	
	usedVars := []string{
		"--spacing-2",
		"--spacing-3",
		"--color-surface",
		"--color-border",
		"--radius-md",
		"--color-background",
		"--color-text-primary",
		"--font-size-base",
		"--font-weight-medium",
		"--color-primary-600",
		"--color-border-focus",
	}
	
	for _, varName := range usedVars {
		if strings.Contains(counterCSS, "var("+varName+")") {
			if !strings.Contains(variablesCSS, varName+":") {
				t.Errorf("counter.css uses %s but it's not defined in variables.css", varName)
			}
		}
	}
}

func TestCounterCSSNoConflicts(t *testing.T) {
	_, err := os.ReadFile("counter.css")
	if err != nil {
		t.Fatalf("Failed to read counter.css: %v", err)
	}
	
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Skip("forms.css not found, skipping conflict test")
	}
	
	formsCSS := string(formsContent)
	
	if strings.Contains(formsCSS, ".counter") {
		t.Error("counter.css classes conflict with forms.css")
	}
	
	if strings.Contains(formsCSS, ".counter-btn") {
		t.Error("counter.css classes conflict with forms.css")
	}
}
