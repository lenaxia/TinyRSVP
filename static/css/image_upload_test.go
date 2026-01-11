package css

import (
	"os"
	"strings"
	"testing"
)

func TestImageUploadCSS_FileExists(t *testing.T) {
	if _, err := os.Stat("image_upload.css"); os.IsNotExist(err) {
		t.Error("image_upload.css does not exist")
	}
}

func TestImageUploadCSS_HasRequiredClasses(t *testing.T) {
	content, err := os.ReadFile("image_upload.css")
	if err != nil {
		t.Fatalf("Failed to read image_upload.css: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".image-upload-section",
		".image-upload-title",
		".image-upload-help",
		".image-upload-container",
		".image-preview",
		".preview-image",
		".image-placeholder",
		".placeholder-icon",
		".placeholder-text",
		".btn-remove-image",
		".image-upload-controls",
		".upload-requirements",
		".upload-progress",
		".progress-bar",
		".progress-fill",
		".progress-text",
		".upload-error",
		".error-message",
		".upload-success",
		".success-message",
		".sr-only",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("image_upload.css missing required class: %s", class)
		}
	}
}

func TestImageUploadCSS_HasResponsiveDesign(t *testing.T) {
	content, err := os.ReadFile("image_upload.css")
	if err != nil {
		t.Fatalf("Failed to read image_upload.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@media") {
		t.Error("image_upload.css missing media queries for responsive design")
	}

	if !strings.Contains(cssContent, "max-width: 767px") {
		t.Error("image_upload.css missing mobile breakpoint")
	}

	if !strings.Contains(cssContent, "min-width: 768px") {
		t.Error("image_upload.css missing tablet/desktop breakpoint")
	}
}

func TestImageUploadCSS_HasAccessibilityFeatures(t *testing.T) {
	content, err := os.ReadFile("image_upload.css")
	if err != nil {
		t.Fatalf("Failed to read image_upload.css: %v", err)
	}

	cssContent := string(content)

	accessibilityFeatures := []string{
		".sr-only",
		"min-width: 44px",
		"min-height: 44px",
		"focus",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(cssContent, feature) {
			t.Errorf("image_upload.css missing accessibility feature: %s", feature)
		}
	}
}

func TestImageUploadCSS_UsesCSSVariables(t *testing.T) {
	content, err := os.ReadFile("image_upload.css")
	if err != nil {
		t.Fatalf("Failed to read image_upload.css: %v", err)
	}

	cssContent := string(content)

	cssVariables := []string{
		"var(--spacing-",
		"var(--color-",
		"var(--radius-",
		"var(--font-size-",
	}

	for _, variable := range cssVariables {
		if !strings.Contains(cssContent, variable) {
			t.Errorf("image_upload.css should use CSS variable pattern: %s", variable)
		}
	}
}

func TestImageUploadCSS_HasDragOverState(t *testing.T) {
	content, err := os.ReadFile("image_upload.css")
	if err != nil {
		t.Fatalf("Failed to read image_upload.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".drag-over") {
		t.Error("image_upload.css missing .drag-over state for drag-and-drop")
	}
}

func TestImageUploadCSS_HasTransitions(t *testing.T) {
	content, err := os.ReadFile("image_upload.css")
	if err != nil {
		t.Fatalf("Failed to read image_upload.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "transition:") {
		t.Error("image_upload.css should have transitions for smooth UX")
	}
}
