package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestHandleThemePreview_XSSProtection(t *testing.T) {
	tests := []struct {
		name    string
		param   string
		payload string
	}{
		{
			name:    "title with script tag is escaped",
			param:   "title",
			payload: `<script>alert("xss")</script>`,
		},
		{
			name:    "location with script tag is escaped",
			param:   "location",
			payload: `<script>alert(1)</script>`,
		},
		{
			name:    "description with script tag is escaped",
			param:   "description",
			payload: `</p><script>document.cookie</script><p>`,
		},
		{
			name:    "title with img onerror",
			param:   "title",
			payload: `<img src=x onerror=alert(1)>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTemplateService{
				GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					return &models.Template{
						ID:       1,
						Name:     "Test Theme",
						Category: "plain",
					}, nil
				},
			}
			h := &TemplateHandlers{service: mockService}

			target := "/themes/preview?theme_id=1&" + tt.param + "=" + url.QueryEscape(tt.payload)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()

			h.HandleThemePreview(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", w.Code)
			}

			body := w.Body.String()
			// Real XSS = an unescaped tag that the browser would execute/render.
			// Escaped payloads appear as &lt;script&gt; / &lt;img — harmless text.
			if strings.Contains(body, "<script>") {
				t.Errorf("response body contains unescaped <script> tag (XSS):\n%s", body)
			}
			if strings.Contains(body, "<img ") && strings.Contains(body, "onerror=") {
				t.Errorf("response body contains an unescaped <img> element with onerror (XSS):\n%s", body)
			}
			// Confirm the payload was escaped, not dropped or passed through raw.
			if !strings.Contains(body, "&lt;script") && !strings.Contains(body, "&lt;img") {
				t.Errorf("expected payload to be HTML-escaped in output, but no escaped form found")
			}
		})
	}
}
