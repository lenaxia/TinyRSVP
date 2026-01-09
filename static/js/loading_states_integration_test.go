package js

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLoadingStatesJSIntegrationServing(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})

	req := httptest.NewRequest("GET", "/static/js/loading_states.js", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/javascript" {
		t.Errorf("Expected Content-Type application/javascript, got %s", contentType)
	}
}

func TestLoadingStatesJSIntegrationWithCSS(t *testing.T) {
	jsContent, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	cssContent, err := os.ReadFile("../css/loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	jsStr := string(jsContent)
	cssStr := string(cssContent)

	jsClasses := []string{
		"loading",
		"spinner",
		"loading-overlay",
		"progress-bar",
	}

	for _, class := range jsClasses {
		t.Run("class_"+class, func(t *testing.T) {
			if !strings.Contains(jsStr, class) {
				t.Errorf("JavaScript should reference class: %s", class)
			}
			if !strings.Contains(cssStr, "."+class) {
				t.Errorf("CSS should define class: .%s", class)
			}
		})
	}
}

func TestLoadingStatesJSIntegrationModulePattern(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "const LoadingStates = (() => {") {
		t.Error("JavaScript should use IIFE module pattern")
	}

	if !strings.Contains(jsContent, "return {") {
		t.Error("JavaScript should return public API")
	}
}

func TestLoadingStatesJSIntegrationStateManagement(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "new Map()") {
		t.Error("JavaScript should use Map for state management")
	}

	if !strings.Contains(jsContent, ".set(") && !strings.Contains(jsContent, ".get(") && !strings.Contains(jsContent, ".delete(") {
		t.Error("JavaScript should use Map methods for state operations")
	}
}

func TestLoadingStatesJSIntegrationARIAImplementation(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	ariaAttributes := map[string]string{
		"aria-busy":      "setAttribute",
		"aria-live":      "setAttribute",
		"aria-label":     "setAttribute",
		"aria-valuenow":  "setAttribute",
		"role":           "setAttribute",
	}

	for attr, method := range ariaAttributes {
		t.Run("aria_"+attr, func(t *testing.T) {
			if !strings.Contains(jsContent, attr) {
				t.Errorf("JavaScript should use ARIA attribute: %s", attr)
			}
			if !strings.Contains(jsContent, method) {
				t.Errorf("JavaScript should use %s method", method)
			}
		})
	}
}

func TestLoadingStatesJSIntegrationButtonManagement(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	buttonOperations := []string{
		"classList.add",
		"classList.remove",
		"disabled = true",
		"disabled = false",
		"textContent",
	}

	for _, operation := range buttonOperations {
		if !strings.Contains(jsContent, operation) {
			t.Errorf("JavaScript should use operation: %s", operation)
		}
	}
}

func TestLoadingStatesJSIntegrationSpinnerCreation(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "createElement") {
		t.Error("JavaScript should dynamically create spinner elements")
	}

	if !strings.Contains(jsContent, "appendChild") {
		t.Error("JavaScript should append spinner to container")
	}

	if !strings.Contains(jsContent, ".remove()") {
		t.Error("JavaScript should remove spinner elements")
	}
}

func TestLoadingStatesJSIntegrationOverlayManagement(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "document.body.appendChild") {
		t.Error("JavaScript should append overlay to body")
	}

	if !strings.Contains(jsContent, "document.body.style.overflow") {
		t.Error("JavaScript should manage body overflow during overlay")
	}

	if !strings.Contains(jsContent, "querySelector('.loading-overlay')") {
		t.Error("JavaScript should check for existing overlay")
	}
}

func TestLoadingStatesJSIntegrationProgressBarUpdate(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "style.width") {
		t.Error("JavaScript should update progress bar width")
	}

	if !strings.Contains(jsContent, "Math.max") && !strings.Contains(jsContent, "Math.min") {
		t.Error("JavaScript should clamp progress percentage")
	}

	if !strings.Contains(jsContent, "%") {
		t.Error("JavaScript should use percentage for progress width")
	}
}

func TestLoadingStatesJSIntegrationTimeoutSupport(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "setTimeout") {
		t.Error("JavaScript should support timeout for auto-hiding loading states")
	}

	if !strings.Contains(jsContent, "options.timeout") {
		t.Error("JavaScript should check for timeout option")
	}
}

func TestLoadingStatesJSIntegrationElementSelector(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "typeof") || !strings.Contains(jsContent, "=== 'string'") {
		t.Error("JavaScript should handle both string selectors and element references")
	}

	if !strings.Contains(jsContent, "querySelector") {
		t.Error("JavaScript should use querySelector for string selectors")
	}
}

func TestLoadingStatesJSIntegrationErrorHandling(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "if (!") {
		t.Error("JavaScript should include null/undefined checks")
	}

	if !strings.Contains(jsContent, "return") {
		t.Error("JavaScript should use early returns for error cases")
	}
}

func TestLoadingStatesJSIntegrationStateTracking(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "loadingStates") {
		t.Error("JavaScript should track loading states")
	}

	if !strings.Contains(jsContent, "Date.now()") {
		t.Error("JavaScript should track timing information")
	}
}

func TestLoadingStatesJSIntegrationCleanup(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "clearAll") {
		t.Error("JavaScript should provide cleanup function")
	}

	if !strings.Contains(jsContent, ".clear()") {
		t.Error("JavaScript should clear state maps")
	}
}

func TestLoadingStatesJSIntegrationPublicAPI(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	publicMethods := []string{
		"showButtonLoading",
		"hideButtonLoading",
		"showSpinner",
		"hideSpinner",
		"showOverlay",
		"hideOverlay",
		"updateProgress",
		"setLoadingState",
		"clearLoadingState",
	}

	returnBlock := ""
	lines := strings.Split(jsContent, "\n")
	inReturn := false
	for _, line := range lines {
		if strings.Contains(line, "return {") {
			inReturn = true
		}
		if inReturn {
			returnBlock += line + "\n"
			if strings.Contains(line, "};") && !strings.Contains(line, "return") {
				break
			}
		}
	}

	for _, method := range publicMethods {
		t.Run("public_method_"+method, func(t *testing.T) {
			if !strings.Contains(returnBlock, method) {
				t.Errorf("Public API should expose method: %s", method)
			}
		})
	}
}

func TestLoadingStatesJSIntegrationNoGlobalPollution(t *testing.T) {
	content, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "window.") && !strings.Contains(jsContent, "// window.") {
		lines := strings.Split(jsContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, "window.") && !strings.Contains(line, "//") {
				t.Error("JavaScript should avoid polluting global namespace with window assignments")
				break
			}
		}
	}
}
