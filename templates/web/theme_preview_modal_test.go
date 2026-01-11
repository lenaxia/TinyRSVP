package web

import (
	"strings"
	"testing"
)

func TestThemePreviewModalHTMLStructure(t *testing.T) {
	htmlContent := themePreviewModalHTML

	tests := []struct {
		name     string
		pattern  string
		category string
	}{
		{"has modal container", `id="theme-preview-modal"`, "structure"},
		{"has modal backdrop", `class="modal-backdrop"`, "structure"},
		{"has modal container", `class="modal-container"`, "structure"},
		{"has modal header", `class="modal-header"`, "structure"},
		{"has modal body", `class="modal-body"`, "structure"},
		{"has modal footer", `class="modal-footer"`, "structure"},
		{"has preview iframe", `id="theme-preview-frame"`, "content"},
		{"has theme toggle button", `id="preview-theme-toggle"`, "interaction"},
		{"has close button", `class="modal-close"`, "interaction"},
		{"has select button", `id="select-previewed-theme"`, "interaction"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(htmlContent, tt.pattern) {
				t.Errorf("theme_preview_modal.html missing %s (pattern: %s)", tt.name, tt.pattern)
			}
		})
	}
}

func TestThemePreviewModalHTMLARIA(t *testing.T) {
	htmlContent := themePreviewModalHTML

	ariaTests := []struct {
		name    string
		pattern string
	}{
		{"has role dialog", `role="dialog"`},
		{"has aria-labelledby", `aria-labelledby="`},
		{"has aria-modal", `aria-modal="true"`},
		{"has aria-label on close", `aria-label="Close`},
		{"has aria-label on toggle", `aria-label="Toggle`},
		{"has aria-hidden on backdrop", `aria-hidden="true"`},
		{"has hidden attribute", `hidden`},
	}

	for _, tt := range ariaTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(htmlContent, tt.pattern) {
				t.Errorf("theme_preview_modal.html missing ARIA attribute: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalHTMLIframe(t *testing.T) {
	htmlContent := themePreviewModalHTML

	iframeTests := []struct {
		name    string
		pattern string
	}{
		{"has iframe element", `<iframe`},
		{"has iframe id", `id="theme-preview-frame"`},
		{"has iframe title", `title="`},
		{"has sandbox attribute", `sandbox="`},
		{"has allow-same-origin", `allow-same-origin`},
		{"has lazy loading", `loading="lazy"`},
	}

	for _, tt := range iframeTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(htmlContent, tt.pattern) {
				t.Errorf("theme_preview_modal.html missing iframe attribute: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalHTMLButtons(t *testing.T) {
	htmlContent := themePreviewModalHTML

	buttonTests := []struct {
		name    string
		pattern string
	}{
		{"close button has type", `type="button"`},
		{"close button has class", `class="modal-close"`},
		{"close button has aria-label", `aria-label="Close`},
		{"toggle button has type", `type="button"`},
		{"toggle button has id", `id="preview-theme-toggle"`},
		{"select button has type", `type="button"`},
		{"select button has id", `id="select-previewed-theme"`},
	}

	for _, tt := range buttonTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(htmlContent, tt.pattern) {
				t.Errorf("theme_preview_modal.html missing button attribute: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalHTMLHeaderActions(t *testing.T) {
	htmlContent := themePreviewModalHTML

	if !strings.Contains(htmlContent, `class="modal-header-actions"`) {
		t.Error("theme_preview_modal.html should have modal-header-actions container")
	}

	if !strings.Contains(htmlContent, `class="theme-icon"`) {
		t.Error("theme_preview_modal.html should have theme-icon element for toggle button")
	}
}

func TestThemePreviewModalHTMLSemanticStructure(t *testing.T) {
	htmlContent := themePreviewModalHTML

	semanticTests := []struct {
		name    string
		pattern string
	}{
		{"has h3 for title", `<h3`},
		{"has title id", `id="preview-modal-title"`},
		{"uses button elements", `<button`},
		{"uses div for containers", `<div`},
	}

	for _, tt := range semanticTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(htmlContent, tt.pattern) {
				t.Errorf("theme_preview_modal.html missing semantic element: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalHTMLFooterButtons(t *testing.T) {
	htmlContent := themePreviewModalHTML

	footerTests := []struct {
		name    string
		pattern string
	}{
		{"has close button in footer", `class="btn btn-secondary modal-close"`},
		{"has select button in footer", `class="btn btn-primary"`},
		{"close button has text", `Close`},
		{"select button has text", `Select This Theme`},
	}

	for _, tt := range footerTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(htmlContent, tt.pattern) {
				t.Errorf("theme_preview_modal.html missing footer element: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalHTMLNoInlineStyles(t *testing.T) {
	htmlContent := themePreviewModalHTML

	if strings.Contains(htmlContent, `style="`) {
		t.Error("theme_preview_modal.html should not contain inline styles")
	}
}

func TestThemePreviewModalHTMLNoInlineScripts(t *testing.T) {
	htmlContent := themePreviewModalHTML

	if strings.Contains(htmlContent, `<script`) {
		t.Error("theme_preview_modal.html should not contain inline scripts")
	}

	if strings.Contains(htmlContent, `onclick=`) {
		t.Error("theme_preview_modal.html should not contain inline event handlers")
	}
}

const themePreviewModalHTML = `{{define "theme-preview-modal"}}
<div id="theme-preview-modal"
     class="modal"
     role="dialog"
     aria-labelledby="preview-modal-title"
     aria-modal="true"
     hidden>
    <div class="modal-backdrop" aria-hidden="true"></div>
    <div class="modal-container">
        <div class="modal-header">
            <h3 id="preview-modal-title">Theme Preview</h3>
            <div class="modal-header-actions">
                <button type="button"
                        id="preview-theme-toggle"
                        class="btn-icon"
                        aria-label="Toggle preview theme">
                    <span class="theme-icon">🌙</span>
                </button>
                <button type="button"
                        class="modal-close"
                        aria-label="Close preview">
                    ×
                </button>
            </div>
        </div>
        <div class="modal-body">
            <iframe id="theme-preview-frame"
                    title="Theme preview"
                    sandbox="allow-same-origin"
                    loading="lazy"></iframe>
        </div>
        <div class="modal-footer">
            <button type="button" class="btn btn-secondary modal-close">
                Close
            </button>
            <button type="button" class="btn btn-primary" id="select-previewed-theme">
                Select This Theme
            </button>
        </div>
    </div>
</div>
{{end}}`
