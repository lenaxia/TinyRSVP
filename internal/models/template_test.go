package models

import (
	"errors"
	"testing"
)

func TestTemplateType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		tt   TemplateType
		want bool
	}{
		{
			name: "valid invite email",
			tt:   TemplateTypeInviteEmail,
			want: true,
		},
		{
			name: "valid rsvp page",
			tt:   TemplateTypeRSVPPage,
			want: true,
		},
		{
			name: "valid confirmation page",
			tt:   TemplateTypeConfirmationPage,
			want: true,
		},
		{
			name: "invalid type",
			tt:   "invalid_type",
			want: false,
		},
		{
			name: "empty type",
			tt:   "",
			want: false,
		},
		{
			name: "random string",
			tt:   "random",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tt.IsValid(); got != tt.want {
				t.Errorf("TemplateType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateType_String(t *testing.T) {
	tests := []struct {
		name string
		tt   TemplateType
		want string
	}{
		{
			name: "invite email",
			tt:   TemplateTypeInviteEmail,
			want: "invite_email",
		},
		{
			name: "rsvp page",
			tt:   TemplateTypeRSVPPage,
			want: "rsvp_page",
		},
		{
			name: "confirmation page",
			tt:   TemplateTypeConfirmationPage,
			want: "confirmation_page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tt.String(); got != tt.want {
				t.Errorf("TemplateType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplate_Validate(t *testing.T) {
	textContent := "Event: {{.Event.Title}}"
	cssContent := ".container { color: red; }"
	eventID := int64(1)

	tests := []struct {
		name     string
		template *Template
		wantErr  bool
		errField string
	}{
		{
			name: "valid invite email template",
			template: &Template{
				Name:        "Custom Invite",
				Type:        TemplateTypeInviteEmail,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				TextContent: &textContent,
				CreatedBy:   1,
				Category:    CategoryPlain,
			},
			wantErr: false,
		},
		{
			name: "valid rsvp page template",
			template: &Template{
				Name:        "Custom RSVP",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    CategoryPlain,
			},
			wantErr: false,
		},
		{
			name: "valid template with all fields",
			template: &Template{
				EventID:     &eventID,
				Name:        "Full Template",
				Type:        TemplateTypeConfirmationPage,
				Description: "A complete template",
				HTMLContent: "<html>{{.Event.Title}}</html>",
				TextContent: &textContent,
				CSSContent:  &cssContent,
				IsDefault:   false,
				IsActive:    true,
				Version:     1,
				CreatedBy:   1,
				Category:    CategoryPlain,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			template: &Template{
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html></html>",
				CreatedBy:   1,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "name too short",
			template: &Template{
				Name:        "ab",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html></html>",
				CreatedBy:   1,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "name too long",
			template: &Template{
				Name:        "a" + string(make([]byte, 100)),
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html></html>",
				CreatedBy:   1,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "invalid type",
			template: &Template{
				Name:        "Test Template",
				Type:        "invalid_type",
				HTMLContent: "<html></html>",
				CreatedBy:   1,
			},
			wantErr:  true,
			errField: "type",
		},
		{
			name: "missing html content",
			template: &Template{
				Name:      "Test Template",
				Type:      TemplateTypeRSVPPage,
				CreatedBy: 1,
			},
			wantErr:  true,
			errField: "html_content",
		},
		{
			name: "email template missing text content",
			template: &Template{
				Name:        "Email Template",
				Type:        TemplateTypeInviteEmail,
				HTMLContent: "<html></html>",
				CreatedBy:   1,
			},
			wantErr:  true,
			errField: "text_content",
		},
		{
			name: "page template with nil text content is valid",
			template: &Template{
				Name:        "Page Template",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html></html>",
				TextContent: nil,
				CreatedBy:   1,
				Category:    CategoryPlain,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Template.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errField != "" {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if valErr.Field != tt.errField {
					t.Errorf("Expected error for field %s, got %s", tt.errField, valErr.Field)
				}
			}
		})
	}
}

func TestTemplate_Validate_EdgeCases(t *testing.T) {
	t.Run("name with exactly 3 characters", func(t *testing.T) {
		template := &Template{
			Name:        "abc",
			Type:        TemplateTypeRSVPPage,
			HTMLContent: "<html></html>",
			CreatedBy:   1,
			Category:    CategoryPlain,
		}
		if err := template.Validate(); err != nil {
			t.Errorf("Expected valid template with 3 character name, got error: %v", err)
		}
	})

	t.Run("name with exactly 100 characters", func(t *testing.T) {
		name := ""
		for i := 0; i < 100; i++ {
			name += "a"
		}
		template := &Template{
			Name:        name,
			Type:        TemplateTypeRSVPPage,
			HTMLContent: "<html></html>",
			CreatedBy:   1,
			Category:    CategoryPlain,
		}
		if err := template.Validate(); err != nil {
			t.Errorf("Expected valid template with 100 character name, got error: %v", err)
		}
	})

	t.Run("confirmation page with text content", func(t *testing.T) {
		textContent := "Confirmation text"
		template := &Template{
			Name:        "Confirmation Template",
			Type:        TemplateTypeConfirmationPage,
			HTMLContent: "<html></html>",
			TextContent: &textContent,
			CreatedBy:   1,
			Category:    CategoryPlain,
		}
		if err := template.Validate(); err != nil {
			t.Errorf("Expected valid confirmation page with text content, got error: %v", err)
		}
	})
}
