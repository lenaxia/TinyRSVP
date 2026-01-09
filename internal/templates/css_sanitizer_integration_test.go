package templates

import (
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestValidator_ValidateTemplate_WithCSS(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	tests := []struct {
		name       string
		template   *models.Template
		wantErr    bool
		errPattern string
	}{
		{
			name: "valid template with safe CSS",
			template: &models.Template{
				Name:        "Safe CSS Template",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("body { color: blue; font-size: 14px; }"),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "template with javascript URL in CSS",
			template: &models.Template{
				Name:        "Malicious CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("background: url(javascript:alert('xss'));"),
				CreatedBy:   1,
			},
			wantErr:    true,
			errPattern: "javascript:",
		},
		{
			name: "template with expression in CSS",
			template: &models.Template{
				Name:        "Expression CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("width: expression(alert('xss'));"),
				CreatedBy:   1,
			},
			wantErr:    true,
			errPattern: "expression(",
		},
		{
			name: "template with behavior in CSS",
			template: &models.Template{
				Name:        "Behavior CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("behavior: url(xss.htc);"),
				CreatedBy:   1,
			},
			wantErr:    true,
			errPattern: "behavior:",
		},
		{
			name: "template with @import in CSS",
			template: &models.Template{
				Name:        "Import CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("@import url('https://evil.com/xss.css');"),
				CreatedBy:   1,
			},
			wantErr:    true,
			errPattern: "@import",
		},
		{
			name: "template with -moz-binding in CSS",
			template: &models.Template{
				Name:        "Moz Binding CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("-moz-binding: url('http://evil.com/xss.xml#xss');"),
				CreatedBy:   1,
			},
			wantErr:    true,
			errPattern: "-moz-binding",
		},
		{
			name: "template with data:text/html in CSS",
			template: &models.Template{
				Name:        "Data HTML CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("background: url('data:text/html,<script>alert(1)</script>');"),
				CreatedBy:   1,
			},
			wantErr:    true,
			errPattern: "data:text/html",
		},
		{
			name: "template with complex safe CSS",
			template: &models.Template{
				Name:        "Complex Safe CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent: strPtr(`
					body {
						font-family: Arial, sans-serif;
						font-size: 16px;
						line-height: 1.6;
						color: #333;
					}
					.container {
						max-width: 600px;
						margin: 0 auto;
						padding: 20px;
					}
					.header {
						background-color: #007bff;
						color: white;
						border-radius: 8px;
					}
					@media (max-width: 768px) {
						.container {
							padding: 10px;
						}
					}
				`),
				CreatedBy: 1,
			},
			wantErr: false,
		},
		{
			name: "template with safe data URL image",
			template: &models.Template{
				Name:        "Safe Data URL",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("background: url('data:image/png;base64,iVBORw0KGgoAAAANS');"),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "template without CSS",
			template: &models.Template{
				Name:        "No CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  nil,
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "template with empty CSS",
			template: &models.Template{
				Name:        "Empty CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr(""),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "invite email template with safe CSS",
			template: &models.Template{
				Name:        "Invite Email CSS",
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1><p>{{.Invite.Name}}</p>",
				TextContent: strPtr("Event: {{.Event.Title}}"),
				CSSContent:  strPtr("h1 { color: #007bff; } p { margin: 10px 0; }"),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "confirmation page template with safe CSS",
			template: &models.Template{
				Name:        "Confirmation CSS",
				Type:        models.TemplateTypeConfirmationPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1><p>{{.RSVP.Response}}</p>",
				CSSContent:  strPtr(".success { color: green; font-weight: bold; }"),
				CreatedBy:   1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errPattern != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errPattern)
					return
				}
				errMsg := err.Error()
				if !contains(errMsg, tt.errPattern) {
					t.Errorf("Error message = %v, want to contain %q", errMsg, tt.errPattern)
				}
			}
		})
	}
}

func TestValidator_CSSValidation_EdgeCases(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	tests := []struct {
		name     string
		template *models.Template
		wantErr  bool
	}{
		{
			name: "CSS size exceeds limit",
			template: &models.Template{
				Name:        "Large CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr(generateLargeCSS(51 * 1024)),
				CreatedBy:   1,
			},
			wantErr: true,
		},
		{
			name: "CSS at size limit",
			template: &models.Template{
				Name:        "Max CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr(generateLargeCSS(50 * 1024)),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "multiple dangerous patterns",
			template: &models.Template{
				Name:        "Multiple Dangerous",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr("body { behavior: url(xss.htc); background: url(javascript:alert('xss')); }"),
				CreatedBy:   1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func generateLargeCSS(size int) string {
	css := ""
	for len(css) < size {
		css += ".class" + string(rune(len(css)%26+97)) + " { color: blue; }\n"
	}
	return css[:size]
}
