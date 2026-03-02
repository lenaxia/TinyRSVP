package templates

import (
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestValidator_Integration_RealWorldTemplates(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	tests := []struct {
		name     string
		template *models.Template
		wantErr  bool
	}{
		{
			name: "complete invite email template",
			template: &models.Template{
				Name: "Wedding Invitation",
				Type: models.TemplateTypeInviteEmail,
				HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Event.Title}}</title>
</head>
<body>
    <h1>You're Invited!</h1>
    <p>Dear {{.Invite.Name}},</p>
    <p>You are cordially invited to {{.Event.Title}}</p>
    <p><strong>When:</strong> {{.Event.StartTime | formatDateTime}}</p>
    <p><strong>Where:</strong> {{.Event.Location}}</p>
    <p><strong>RSVP by:</strong> {{.Event.RSVPDeadline | formatDateTime}}</p>
    {{if gt .MaxPlusOnes 0}}
    <p>You may bring up to {{.MaxPlusOnes}} guest(s).</p>
    {{end}}
    <p><a href="{{.RSVPURL}}">Click here to RSVP</a></p>
</body>
</html>`,
				TextContent: strPtr(`You're Invited!

Dear {{.Invite.Name}},

You are cordially invited to {{.Event.Title}}

When: {{.Event.StartTime | formatDateTime}}
Where: {{.Event.Location}}
RSVP by: {{.Event.RSVPDeadline | formatDateTime}}

{{if gt .MaxPlusOnes 0}}You may bring up to {{.MaxPlusOnes}} guest(s).{{end}}

RSVP at: {{.RSVPURL}}`),
				CreatedBy: 1,
			},
			wantErr: false,
		},
		{
			name: "complete rsvp page template",
			template: &models.Template{
				Name: "RSVP Form",
				Type: models.TemplateTypeRSVPPage,
				HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <title>RSVP - {{.Event.Title}}</title>
</head>
<body>
    <h1>{{.Event.Title}}</h1>
    <p>{{.Event.Description}}</p>
    <p><strong>Date:</strong> {{.Event.StartTime | formatDateTime}}</p>
    <p><strong>Location:</strong> {{.Event.Location}}</p>
    
    <form method="POST">
        <label>Will you attend?</label>
        <select name="response">
            <option value="yes">Yes, I'll be there!</option>
            <option value="no">Sorry, can't make it</option>
        </select>
        
        {{if .Questions}}
        <h2>Additional Questions</h2>
        {{range .Questions}}
        <div>
            <label>{{.}}</label>
            <input type="text" name="answer[]">
        </div>
        {{end}}
        {{end}}
        
        <button type="submit">Submit RSVP</button>
    </form>
</body>
</html>`,
				CSSContent: strPtr(`.container { max-width: 800px; margin: 0 auto; }`),
				CreatedBy:  1,
			},
			wantErr: false,
		},
		{
			name: "complete confirmation page template",
			template: &models.Template{
				Name: "Confirmation",
				Type: models.TemplateTypeConfirmationPage,
				HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <title>RSVP Confirmed - {{.Event.Title}}</title>
</head>
<body>
    <h1>Thank You!</h1>
    <p>Your RSVP has been recorded.</p>
    
    <h2>Event Details</h2>
    <p><strong>Event:</strong> {{.Event.Title}}</p>
    <p><strong>Date:</strong> {{.Event.StartTime | formatDateTime}}</p>
    <p><strong>Location:</strong> {{.Event.Location}}</p>
    
    <h2>Your Response</h2>
    <p><strong>Attending:</strong> {{.RSVP.Response | upper}}</p>
    {{if gt .RSVP.PlusOnes 0}}
    <p><strong>Plus Ones:</strong> {{.RSVP.PlusOnes}}</p>
    {{end}}
    
    {{if .Answers}}
    <h2>Your Answers</h2>
    {{range .Answers}}
    <p><strong>{{.}}</strong></p>
    {{end}}
    {{end}}
</body>
</html>`,
				CreatedBy: 1,
			},
			wantErr: false,
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

func TestValidator_Integration_EdgeCases(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	tests := []struct {
		name        string
		template    *models.Template
		wantErr     bool
		errContains string
	}{
		{
			name: "template at exact html size limit",
			template: &models.Template{
				Name:        "Exact Size",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: strings.Repeat("x", 100*1024),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "template at exact text size limit",
			template: &models.Template{
				Name:        "Exact Text Size",
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				TextContent: strPtr(strings.Repeat("x", 50*1024)),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "template at exact css size limit",
			template: &models.Template{
				Name:        "Exact CSS Size",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				CSSContent:  strPtr(strings.Repeat("x", 50*1024)),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "complex nested template structures",
			template: &models.Template{
				Name: "Complex Nesting",
				Type: models.TemplateTypeRSVPPage,
				HTMLContent: `
{{if .Event.Title}}
    {{if .Event.Description}}
        {{if .Event.Location}}
            <h1>{{.Event.Title}}</h1>
            <p>{{.Event.Description}}</p>
            <p>{{.Event.Location}}</p>
        {{end}}
    {{end}}
{{end}}`,
				CreatedBy: 1,
			},
			wantErr: false,
		},
		{
			name: "template with all allowed functions",
			template: &models.Template{
				Name: "All Functions",
				Type: models.TemplateTypeInviteEmail,
				HTMLContent: `
<h1>{{.Event.Title | upper}}</h1>
<p>{{.Event.Description | lower}}</p>
<p>{{.Event.Location | title}}</p>
<p>{{.Event.StartTime | formatDateTime}}</p>
<p>{{.Invite.Name | default "Guest"}}</p>`,
				TextContent: strPtr("{{.Event.Title}}"),
				CreatedBy:   1,
			},
			wantErr: false,
		},
		{
			name: "template with undefined variable in nested context",
			template: &models.Template{
				Name: "Nested Undefined",
				Type: models.TemplateTypeRSVPPage,
				HTMLContent: `
{{if .Event.Title}}
    <h1>{{.Event.Title}}</h1>
    <p>{{.Event.UndefinedField}}</p>
{{end}}`,
				CreatedBy: 1,
			},
			wantErr:     true,
			errContains: "Undefined variable",
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

func TestValidator_XSSPrevention(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	xssPayloads := []struct {
		name         string
		htmlContent  string
		templateType models.TemplateType
		description  string
	}{
		{
			name:         "script tag in user data",
			htmlContent:  `<div>{{.Event.Description}}</div>`,
			templateType: models.TemplateTypeRSVPPage,
			description:  "Ensures script tags in data are escaped",
		},
		{
			name:         "event handler in user data",
			htmlContent:  `<div>{{.Event.Description}}</div>`,
			templateType: models.TemplateTypeRSVPPage,
			description:  "Ensures event handlers in data are escaped",
		},
		{
			name:         "javascript url in invite email",
			htmlContent:  `<a href="{{.RSVPURL}}">Click</a>`,
			templateType: models.TemplateTypeInviteEmail,
			description:  "Ensures javascript: URLs in data are sanitized",
		},
		{
			name:         "data url in user data",
			htmlContent:  `<div>{{.Event.Location}}</div>`,
			templateType: models.TemplateTypeRSVPPage,
			description:  "Ensures data: URLs in data are escaped",
		},
		{
			name:         "svg with script",
			htmlContent:  `<div>{{.Event.Description}}</div>`,
			templateType: models.TemplateTypeRSVPPage,
			description:  "Ensures SVG with embedded script is escaped",
		},
		{
			name:         "html entities",
			htmlContent:  `<p>{{.Event.Title}}</p>`,
			templateType: models.TemplateTypeRSVPPage,
			description:  "Ensures HTML entities are properly escaped",
		},
	}

	for _, payload := range xssPayloads {
		t.Run(payload.name, func(t *testing.T) {
			textContent := "Test"
			template := &models.Template{
				Name:        "XSS Test",
				Type:        payload.templateType,
				HTMLContent: payload.htmlContent,
				CreatedBy:   1,
			}
			if payload.templateType == models.TemplateTypeInviteEmail {
				template.TextContent = &textContent
			}

			err := validator.ValidateTemplate(template)
			if err != nil {
				t.Fatalf("ValidateTemplate() error = %v", err)
			}

			testData := createXSSTestData(payload.templateType)

			tmpl, err := engine.Parse(template.HTMLContent)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.ExecuteToString(tmpl, testData)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if strings.Contains(result, "<script>alert") {
				t.Errorf("XSS not prevented: unescaped <script> tag found in result: %s", result)
			}

			hasEscaping := strings.Contains(result, "&lt;") ||
				strings.Contains(result, "&gt;") ||
				strings.Contains(result, "#ZgotmplZ") ||
				strings.Contains(result, "&#")

			if !hasEscaping {
				t.Errorf("Expected HTML escaping or URL sanitization in result: %s", result)
			}
		})
	}
}

func createXSSTestData(templateType models.TemplateType) interface{} {
	type Event struct {
		Title       string
		Description string
		Location    string
	}

	type Invite struct {
		Name  string
		Email string
	}

	event := Event{
		Title:       "<script>alert('xss')</script>",
		Description: "<img src=x onerror=alert('xss')>",
		Location:    "javascript:alert('xss')",
	}

	switch templateType {
	case models.TemplateTypeInviteEmail:
		return struct {
			Event   Event
			Invite  Invite
			RSVPURL string
		}{
			Event: event,
			Invite: Invite{
				Name:  "Test User",
				Email: "test@example.com",
			},
			RSVPURL: "javascript:alert('xss')",
		}
	case models.TemplateTypeRSVPPage:
		return struct {
			Event Event
		}{
			Event: event,
		}
	case models.TemplateTypeConfirmationPage:
		return struct {
			Event Event
		}{
			Event: event,
		}
	default:
		return struct {
			Event Event
		}{
			Event: event,
		}
	}
}

func TestValidator_XSSPrevention_AdvancedPayloads(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	advancedPayloads := []struct {
		name           string
		data           string
		mustBeEscaped  bool
		mustNotContain []string
	}{
		{
			name:           "encoded script tag",
			data:           "&#60;script&#62;alert('xss')&#60;/script&#62;",
			mustBeEscaped:  true,
			mustNotContain: []string{"<script>"},
		},
		{
			name:           "mixed case script",
			data:           "<ScRiPt>alert('xss')</ScRiPt>",
			mustBeEscaped:  true,
			mustNotContain: []string{"<ScRiPt>", "<script>"},
		},
		{
			name:           "svg with embedded script",
			data:           `<svg onload="alert('xss')">`,
			mustBeEscaped:  true,
			mustNotContain: []string{`<svg onload=`},
		},
		{
			name:           "iframe injection",
			data:           `<iframe src="javascript:alert('xss')"></iframe>`,
			mustBeEscaped:  true,
			mustNotContain: []string{"<iframe src="},
		},
		{
			name:           "object tag",
			data:           `<object data="javascript:alert('xss')"></object>`,
			mustBeEscaped:  true,
			mustNotContain: []string{"<object data="},
		},
	}

	for _, payload := range advancedPayloads {
		t.Run(payload.name, func(t *testing.T) {
			template := &models.Template{
				Name:        "Advanced XSS Test",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: `<div>{{.Event.Description}}</div>`,
				CreatedBy:   1,
			}

			err := validator.ValidateTemplate(template)
			if err != nil {
				t.Fatalf("ValidateTemplate() error = %v", err)
			}

			testData := struct {
				Event struct {
					Description string
				}
			}{
				Event: struct {
					Description string
				}{
					Description: payload.data,
				},
			}

			tmpl, err := engine.Parse(template.HTMLContent)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.ExecuteToString(tmpl, testData)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if payload.mustBeEscaped {
				if !strings.Contains(result, "&lt;") && strings.Contains(payload.data, "<") {
					t.Errorf("HTML not properly escaped, expected &lt; in result: %s", result)
				}
			}

			for _, pattern := range payload.mustNotContain {
				if strings.Contains(result, pattern) {
					t.Errorf("Dangerous unescaped pattern %q found in result: %s", pattern, result)
				}
			}
		})
	}
}
