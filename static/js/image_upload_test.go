package js

import (
	"os"
	"strings"
	"testing"
)

func TestImageUploadJS_FileExists(t *testing.T) {
	if _, err := os.Stat("image_upload.js"); os.IsNotExist(err) {
		t.Error("image_upload.js does not exist")
	}
}

func TestImageUploadJS_ValidSyntax(t *testing.T) {
	content, err := os.ReadFile("image_upload.js")
	if err != nil {
		t.Fatalf("Failed to read image_upload.js: %v", err)
	}

	jsContent := string(content)

	requiredElements := []string{
		"class ImageUploader",
		"constructor()",
		"init()",
		"attachEventListeners()",
		"handleFile(",
		"validateFile(",
		"previewFile(",
		"uploadFile(",
		"removeImage()",
		"showProgress()",
		"hideProgress()",
		"showError(",
		"hideError()",
		"showSuccess(",
		"hideSuccess()",
		"updatePreview(",
	}

	for _, element := range requiredElements {
		if !strings.Contains(jsContent, element) {
			t.Errorf("image_upload.js missing required element: %s", element)
		}
	}
}

func TestImageUploadJS_HasValidationLogic(t *testing.T) {
	content, err := os.ReadFile("image_upload.js")
	if err != nil {
		t.Fatalf("Failed to read image_upload.js: %v", err)
	}

	jsContent := string(content)

	validationChecks := []string{
		"maxSize",
		"maxDimensions",
		"allowedTypes",
		"5 * 1024 * 1024",
		"4096",
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, check := range validationChecks {
		if !strings.Contains(jsContent, check) {
			t.Errorf("image_upload.js missing validation check: %s", check)
		}
	}
}

func TestImageUploadJS_HasDragDropSupport(t *testing.T) {
	content, err := os.ReadFile("image_upload.js")
	if err != nil {
		t.Fatalf("Failed to read image_upload.js: %v", err)
	}

	jsContent := string(content)

	dragDropElements := []string{
		"dragover",
		"dragleave",
		"drop",
		"drag-over",
		"dataTransfer",
	}

	for _, element := range dragDropElements {
		if !strings.Contains(jsContent, element) {
			t.Errorf("image_upload.js missing drag-drop element: %s", element)
		}
	}
}

func TestImageUploadJS_HasProgressIndicator(t *testing.T) {
	content, err := os.ReadFile("image_upload.js")
	if err != nil {
		t.Fatalf("Failed to read image_upload.js: %v", err)
	}

	jsContent := string(content)

	progressElements := []string{
		"upload-progress",
		"progress-fill",
		"progress-text",
		"showProgress",
		"hideProgress",
	}

	for _, element := range progressElements {
		if !strings.Contains(jsContent, element) {
			t.Errorf("image_upload.js missing progress element: %s", element)
		}
	}
}

func TestImageUploadJS_HasErrorHandling(t *testing.T) {
	content, err := os.ReadFile("image_upload.js")
	if err != nil {
		t.Fatalf("Failed to read image_upload.js: %v", err)
	}

	jsContent := string(content)

	errorElements := []string{
		"upload-error",
		"error-message",
		"showError",
		"hideError",
		"catch",
	}

	for _, element := range errorElements {
		if !strings.Contains(jsContent, element) {
			t.Errorf("image_upload.js missing error handling element: %s", element)
		}
	}
}

func TestImageUploadJS_HasCSRFTokenHandling(t *testing.T) {
	content, err := os.ReadFile("image_upload.js")
	if err != nil {
		t.Fatalf("Failed to read image_upload.js: %v", err)
	}

	jsContent := string(content)

	csrfElements := []string{
		"csrf_token",
		"X-CSRF-Token",
	}

	for _, element := range csrfElements {
		if !strings.Contains(jsContent, element) {
			t.Errorf("image_upload.js missing CSRF element: %s", element)
		}
	}
}

func TestImageUploadJS_HasAPIEndpoint(t *testing.T) {
	content, err := os.ReadFile("image_upload.js")
	if err != nil {
		t.Fatalf("Failed to read image_upload.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "/api/events/") {
		t.Error("image_upload.js missing API endpoint")
	}

	if !strings.Contains(jsContent, "/images") {
		t.Error("image_upload.js missing /images path")
	}
}
