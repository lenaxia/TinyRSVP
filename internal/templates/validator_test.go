package templates

import (
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestNewValidator(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	if validator == nil {
		t.Fatal("NewValidator() returned nil")
	}
}

func TestGetAllowedVariables(t *testing.T) {
	tests := []struct {
		name         string
		templateType models.TemplateType
		wantVars     []string
	}{
		{
			name:         "invite email template",
			templateType: models.TemplateTypeInviteEmail,
			wantVars: []string{
				"Event.Title",
				"Event.Description",
				"Event.StartTime",
				"Event.EndTime",
				"Event.Timezone",
				"Event.Location",
				"Event.RSVPDeadline",
				"Invite.Name",
				"Invite.Email",
				"RSVPURL",
				"MaxPlusOnes",
			},
		},
		{
			name:         "rsvp page template",
			templateType: models.TemplateTypeRSVPPage,
			wantVars: []string{
				"Event.Title",
				"Event.Description",
				"Event.StartTime",
				"Event.EndTime",
				"Event.Timezone",
				"Event.Location",
				"Event.RSVPDeadline",
				"Token",
				"MaxPlusOnes",
				"RSVP.Response",
				"RSVP.PlusOnes",
				"Questions",
				"ID",
				"QuestionText",
				"QuestionType",
				"Required",
				"HelpText",
				"Options",
				"Value",
				"Label",
			},
		},
		{
			name:         "confirmation page template",
			templateType: models.TemplateTypeConfirmationPage,
			wantVars: []string{
				"Event.Title",
				"Event.Description",
				"Event.StartTime",
				"Event.EndTime",
				"Event.Timezone",
				"Event.Location",
				"Event.RSVPDeadline",
				"Token",
				"RSVP.Response",
				"RSVP.PlusOnes",
				"RSVP.Notes",
				"Answers",
				"QuestionText",
				"AnswerDisplay",
			},
		},
		{
			name:         "unknown template type returns common vars",
			templateType: "unknown",
			wantVars: []string{
				"Event.Title",
				"Event.Description",
				"Event.StartTime",
				"Event.EndTime",
				"Event.Timezone",
				"Event.Location",
				"Event.RSVPDeadline",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAllowedVariables(tt.templateType)

			if len(got) != len(tt.wantVars) {
				t.Errorf("getAllowedVariables() returned %d vars, want %d", len(got), len(tt.wantVars))
			}

			gotMap := make(map[string]bool)
			for _, v := range got {
				gotMap[v] = true
			}

			for _, want := range tt.wantVars {
				if !gotMap[want] {
					t.Errorf("getAllowedVariables() missing variable: %s", want)
				}
			}
		})
	}
}

func TestValidator_ValidateSize(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	tests := []struct {
		name     string
		content  string
		maxBytes int
		wantErr  bool
	}{
		{
			name:     "within limit",
			content:  "Hello World",
			maxBytes: 100,
			wantErr:  false,
		},
		{
			name:     "at boundary",
			content:  string(make([]byte, 100)),
			maxBytes: 100,
			wantErr:  false,
		},
		{
			name:     "exceeds limit by 1",
			content:  string(make([]byte, 101)),
			maxBytes: 100,
			wantErr:  true,
		},
		{
			name:     "far exceeds limit",
			content:  string(make([]byte, 200)),
			maxBytes: 100,
			wantErr:  true,
		},
		{
			name:     "empty content",
			content:  "",
			maxBytes: 100,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSize(tt.content, tt.maxBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateSyntax(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	tests := []struct {
		name         string
		content      string
		templateType models.TemplateType
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid template",
			content:      "<h1>{{.Event.Title}}</h1>",
			templateType: models.TemplateTypeRSVPPage,
			wantErr:      false,
		},
		{
			name:         "parse error - unclosed action",
			content:      "{{.Event.Title",
			templateType: models.TemplateTypeRSVPPage,
			wantErr:      true,
			errContains:  "syntax error",
		},
		{
			name:         "parse error - invalid syntax",
			content:      "{{.Event.Title}}{{",
			templateType: models.TemplateTypeRSVPPage,
			wantErr:      true,
			errContains:  "syntax error",
		},
		{
			name:         "valid with functions",
			content:      "{{.Event.Title | upper}}",
			templateType: models.TemplateTypeRSVPPage,
			wantErr:      false,
		},
		{
			name:         "valid with if statement",
			content:      "{{if .Event.Title}}{{.Event.Title}}{{end}}",
			templateType: models.TemplateTypeRSVPPage,
			wantErr:      false,
		},
		{
			name:         "valid invite email",
			content:      "<p>Hello {{.Invite.Name}}</p>",
			templateType: models.TemplateTypeInviteEmail,
			wantErr:      false,
		},
		{
			name:         "valid confirmation page",
			content:      "<p>Response: {{.RSVP.Response}}</p>",
			templateType: models.TemplateTypeConfirmationPage,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSyntax(tt.content, tt.templateType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSyntax() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateSyntax() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidator_ValidateVariables(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	tests := []struct {
		name        string
		content     string
		allowedVars []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "all variables allowed",
			content:     "{{.Event.Title}} {{.Event.Location}}",
			allowedVars: []string{"Event.Title", "Event.Location"},
			wantErr:     false,
		},
		{
			name:        "undefined variable",
			content:     "{{.UndefinedVar}}",
			allowedVars: []string{"Event.Title"},
			wantErr:     true,
			errContains: "Undefined variable",
		},
		{
			name:        "nested variable allowed",
			content:     "{{.Event.Title}}",
			allowedVars: []string{"Event.Title", "Event.Location"},
			wantErr:     false,
		},
		{
			name:        "nested variable not allowed",
			content:     "{{.Event.Secret}}",
			allowedVars: []string{"Event.Title", "Event.Location"},
			wantErr:     true,
			errContains: "Event.Secret",
		},
		{
			name:        "multiple variables some undefined",
			content:     "{{.Event.Title}} {{.BadVar}}",
			allowedVars: []string{"Event.Title"},
			wantErr:     true,
			errContains: "BadVar",
		},
		{
			name:        "variable in if statement",
			content:     "{{if .Event.Title}}yes{{end}}",
			allowedVars: []string{"Event.Title"},
			wantErr:     false,
		},
		{
			name:        "variable in range statement",
			content:     "{{range .Questions}}{{.}}{{end}}",
			allowedVars: []string{"Questions"},
			wantErr:     false,
		},
		{
			name:        "empty template",
			content:     "",
			allowedVars: []string{"Event.Title"},
			wantErr:     false,
		},
		{
			name:        "no variables used",
			content:     "<h1>Static Content</h1>",
			allowedVars: []string{"Event.Title"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVariables(tt.content, tt.allowedVars)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateVariables() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidator_ValidateTemplate(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	textContent := "Event: {{.Event.Title}}"
	cssContent := ".container { color: red; }"
	largeHTML := string(make([]byte, 101*1024))
	largeText := string(make([]byte, 51*1024))
	largeCSS := string(make([]byte, 51*1024))

	tests := []struct {
		name        string
		template    *models.Template
		wantErr     bool
		errContains string
	}{
		{
			name: "valid invite email template",
			template: &models.Template{
				Name:        "Custom Invite",
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1><p>{{.Invite.Name}}</p>",
				TextContent: &textContent,
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "valid rsvp page template",
			template: &models.Template{
				Name:        "Custom RSVP",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "valid with css",
			template: &models.Template{
				Name:        "Styled Template",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  &cssContent,
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "html size exceeded",
			template: &models.Template{
				Name:        "Large Template",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: largeHTML,
				CreatedBy:   1,
			},
			wantErr:     true,
			errContains: "exceeds",
		},
		{
			name: "text size exceeded",
			template: &models.Template{
				Name:        "Large Text",
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				TextContent: &largeText,
				CreatedBy:   1,
			},
			wantErr:     true,
			errContains: "exceeds",
		},
		{
			name: "css size exceeded",
			template: &models.Template{
				Name:        "Large CSS",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  &largeCSS,
				CreatedBy:   1,
			},
			wantErr:     true,
			errContains: "exceeds",
		},
		{
			name: "syntax error in html",
			template: &models.Template{
				Name:        "Broken HTML",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "{{.Event.Title",
				CreatedBy:   1,
			},
			wantErr:     true,
			errContains: "parse",
		},
		{
			name: "undefined variable in html",
			template: &models.Template{
				Name:        "Bad Variable",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "{{.UndefinedVar}}",
				CreatedBy:   1,
			},
			wantErr:     true,
			errContains: "Undefined variable",
		},
		{
			name: "missing name fails model validation",
			template: &models.Template{
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>Test</h1>",
				CreatedBy:   1,
			},
			wantErr:     true,
			errContains: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateTemplate() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
