package templates

import (
	"context"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestPreviewTemplate_Integration_WithDefaultTemplates(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(nil, validator)
	ctx := context.Background()

	defaultTemplates := []struct {
		name         string
		templateType models.TemplateType
		htmlContent  string
		expectInHTML string
	}{
		{
			name:         "invite_email default",
			templateType: models.TemplateTypeInviteEmail,
			htmlContent:  "<h1>{{.Event.Title}}</h1><p>Dear {{.Invite.Name}},</p><p>RSVP: {{.RSVPURL}}</p>",
			expectInHTML: "Sample Event",
		},
		{
			name:         "rsvp_page default",
			templateType: models.TemplateTypeRSVPPage,
			htmlContent:  "<h1>{{.Event.Title}}</h1><p>Token: {{.Token}}</p>",
			expectInHTML: "sample-token-preview",
		},
		{
			name:         "confirmation_page default",
			templateType: models.TemplateTypeConfirmationPage,
			htmlContent:  "<h1>Thank you!</h1><p>Event: {{.Event.Title}}</p><p>Response: {{.RSVP.Response}}</p>",
			expectInHTML: "yes",
		},
	}

	for _, tt := range defaultTemplates {
		t.Run(tt.name, func(t *testing.T) {
			req := &PreviewRequest{
				Type:        tt.templateType,
				HTMLContent: tt.htmlContent,
			}

			resp, err := service.PreviewTemplate(ctx, req)
			if err != nil {
				t.Fatalf("PreviewTemplate() error = %v", err)
			}

			if resp == nil {
				t.Fatal("Expected non-nil response")
			}

			if resp.HTMLPreview == "" {
				t.Error("Expected non-empty HTML preview")
			}

			if !strings.Contains(resp.HTMLPreview, tt.expectInHTML) {
				t.Errorf("Expected HTML preview to contain '%s', got: %s", tt.expectInHTML, resp.HTMLPreview)
			}
		})
	}
}

func TestPreviewTemplate_Integration_AllTemplateTypes(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(nil, validator)
	ctx := context.Background()

	types := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, templateType := range types {
		t.Run(string(templateType), func(t *testing.T) {
			req := &PreviewRequest{
				Type:        templateType,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
			}

			resp, err := service.PreviewTemplate(ctx, req)
			if err != nil {
				t.Fatalf("PreviewTemplate() error = %v", err)
			}

			if resp == nil {
				t.Fatal("Expected non-nil response")
			}

			if !strings.Contains(resp.HTMLPreview, "Sample Event") {
				t.Error("Expected HTML preview to contain event title")
			}
		})
	}
}

func TestPreviewTemplate_Integration_AllVariableTypes(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(nil, validator)
	ctx := context.Background()

	tests := []struct {
		name         string
		templateType models.TemplateType
		htmlContent  string
		expectInHTML string
	}{
		{
			name:         "string variable",
			templateType: models.TemplateTypeInviteEmail,
			htmlContent:  "{{.Event.Title}}",
			expectInHTML: "Sample Event",
		},
		{
			name:         "time variable",
			templateType: models.TemplateTypeInviteEmail,
			htmlContent:  "{{.Event.StartTime}}",
			expectInHTML: "2026",
		},
		{
			name:         "pointer time variable",
			templateType: models.TemplateTypeInviteEmail,
			htmlContent:  "{{.Event.EndTime}}",
			expectInHTML: "2026",
		},
		{
			name:         "int variable",
			templateType: models.TemplateTypeInviteEmail,
			htmlContent:  "Max plus ones: {{.MaxPlusOnes}}",
			expectInHTML: "2",
		},
		{
			name:         "nested struct",
			templateType: models.TemplateTypeInviteEmail,
			htmlContent:  "{{.Invite.Name}} - {{.Invite.Email}}",
			expectInHTML: "John Doe",
		},
		{
			name:         "array iteration",
			templateType: models.TemplateTypeRSVPPage,
			htmlContent:  "{{range .Questions}}{{.QuestionText}}{{end}}",
			expectInHTML: "Dietary restrictions?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &PreviewRequest{
				Type:        tt.templateType,
				HTMLContent: tt.htmlContent,
			}

			resp, err := service.PreviewTemplate(ctx, req)
			if err != nil {
				t.Fatalf("PreviewTemplate() error = %v", err)
			}

			if !strings.Contains(resp.HTMLPreview, tt.expectInHTML) {
				t.Errorf("Expected HTML preview to contain '%s', got: %s", tt.expectInHTML, resp.HTMLPreview)
			}
		})
	}
}

func TestPreviewTemplate_Integration_TemplateFunctions(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(nil, validator)
	ctx := context.Background()

	tests := []struct {
		name         string
		htmlContent  string
		expectInHTML string
	}{
		{
			name:         "upper function",
			htmlContent:  "{{.Event.Title | upper}}",
			expectInHTML: "SAMPLE EVENT",
		},
		{
			name:         "lower function",
			htmlContent:  "{{.Event.Title | lower}}",
			expectInHTML: "sample event",
		},
		{
			name:         "formatDateTime function",
			htmlContent:  "{{formatDateTime .Event.StartTime}}",
			expectInHTML: "2026",
		},
		{
			name:         "formatTime function",
			htmlContent:  "{{formatTime .Event.StartTime}}",
			expectInHTML: "M",
		},
		{
			name:         "default function",
			htmlContent:  "{{default .Event.Description \"No description\"}}",
			expectInHTML: "This is a sample event",
		},
		{
			name:         "truncate function",
			htmlContent:  "{{truncate .Event.Description 10}}",
			expectInHTML: "This is a ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &PreviewRequest{
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: tt.htmlContent,
			}

			resp, err := service.PreviewTemplate(ctx, req)
			if err != nil {
				t.Fatalf("PreviewTemplate() error = %v", err)
			}

			if !strings.Contains(resp.HTMLPreview, tt.expectInHTML) {
				t.Errorf("Expected HTML preview to contain '%s', got: %s", tt.expectInHTML, resp.HTMLPreview)
			}
		})
	}
}

func TestPreviewTemplate_Integration_ErrorHandling(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(nil, validator)
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *PreviewRequest
		wantErr     bool
		expectField string
	}{
		{
			name: "syntax error in HTML",
			req: &PreviewRequest{
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "{{.Event.Title",
			},
			wantErr:     true,
			expectField: "html_content",
		},
		{
			name: "undefined variable in HTML",
			req: &PreviewRequest{
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "{{.UndefinedVar}}",
			},
			wantErr:     true,
			expectField: "html_content",
		},
		{
			name: "syntax error in text",
			req: &PreviewRequest{
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				TextContent: strPtr("{{.Event.Title"),
			},
			wantErr:     true,
			expectField: "text_content",
		},
		{
			name: "undefined variable in text",
			req: &PreviewRequest{
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				TextContent: strPtr("{{.UndefinedVar}}"),
			},
			wantErr:     true,
			expectField: "text_content",
		},
		{
			name: "invalid template type",
			req: &PreviewRequest{
				Type:        models.TemplateType("invalid"),
				HTMLContent: "<h1>Test</h1>",
			},
			wantErr:     true,
			expectField: "type",
		},
		{
			name: "empty HTML content",
			req: &PreviewRequest{
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "",
			},
			wantErr:     true,
			expectField: "html_content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.PreviewTemplate(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("PreviewTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				validationErr, ok := err.(*models.ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}

				if validationErr.Field != tt.expectField {
					t.Errorf("Expected field '%s', got '%s'", tt.expectField, validationErr.Field)
				}
			} else {
				if resp == nil {
					t.Error("Expected non-nil response")
				}
			}
		})
	}
}
