package templates

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestXSSPrevention_Integration_AllTemplateTypes(t *testing.T) {
	engine := NewEngine()
	
	maliciousData := struct {
		Event struct {
			Title       string
			Description string
			Location    string
			StartTime   time.Time
		}
		Guest struct {
			Name  string
			Email string
		}
		RSVPURL string
	}{
		Event: struct {
			Title       string
			Description string
			Location    string
			StartTime   time.Time
		}{
			Title:       "<script>alert('xss')</script>",
			Description: "<img src=x onerror=alert('xss')>",
			Location:    "javascript:alert('xss')",
			StartTime:   time.Now(),
		},
		Guest: struct {
			Name  string
			Email string
		}{
			Name:  "<svg onload=alert('xss')>",
			Email: "test@example.com",
		},
		RSVPURL: "javascript:alert('xss')",
	}
	
	templates := []struct {
		name     string
		template string
		context  string
	}{
		{
			name:     "invite_email",
			template: "<h1>{{.Event.Title}}</h1><p>{{.Event.Description}}</p><a href='{{.RSVPURL}}'>RSVP</a>",
			context:  "email",
		},
		{
			name:     "rsvp_page",
			template: "<h1>{{.Event.Title}}</h1><p>Location: {{.Event.Location}}</p><p>Guest: {{.Guest.Name}}</p>",
			context:  "web",
		},
		{
			name:     "confirmation_page",
			template: "<h1>Thank you, {{.Guest.Name}}!</h1><p>Event: {{.Event.Title}}</p>",
			context:  "web",
		},
	}
	
	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			result, err := engine.ExecuteToString(tmpl, maliciousData)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}
			
			if strings.Contains(result, "<script>") && !strings.Contains(result, "&lt;script&gt;") {
				t.Errorf("%s: Script tags not escaped in output: %s", tt.name, result)
			}
			
			resultLower := strings.ToLower(result)
			if strings.Contains(resultLower, "onerror=alert") && !strings.Contains(result, "&lt;") {
				t.Errorf("%s: Event handlers not escaped in output: %s", tt.name, result)
			}
			
			if strings.Contains(resultLower, "onload=alert") && !strings.Contains(result, "&lt;") {
				t.Errorf("%s: Event handlers not escaped in output: %s", tt.name, result)
			}
			
			if strings.Contains(result, "<") && !strings.Contains(result, "&lt;") && !strings.Contains(result, "<!DOCTYPE") {
				t.Errorf("%s: Expected HTML escaping for user input: %s", tt.name, result)
			}
		})
	}
}

func TestXSSPrevention_Integration_RealWorldScenario(t *testing.T) {
	engine := NewEngine()
	
	desc := "Join us for fun! <img src=x onerror=alert('xss')>"
	loc := "123 Main St"
	event := &models.Event{
		Title:       "Company Party <script>alert('xss')</script>",
		Description: &desc,
		Location:    &loc,
		StartTime:   time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC),
	}
	
	template := `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Event.Title}}</title>
</head>
<body>
	<h1>{{.Event.Title}}</h1>
	<p>{{.Event.Description}}</p>
	<p>Location: {{.Event.Location}}</p>
	<p>Time: {{formatDateTime .Event.StartTime}}</p>
</body>
</html>
`
	
	data := struct {
		Event *models.Event
	}{
		Event: event,
	}
	
	tmpl, err := engine.Parse(template)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if strings.Contains(result, "<script>alert") && !strings.Contains(result, "&lt;script&gt;") {
		t.Error("XSS not prevented: unescaped script tag found in output")
	}
	
	if strings.Contains(result, "onerror=alert") && !strings.Contains(result, "&lt;") {
		t.Error("XSS not prevented: unescaped event handler found in output")
	}
	
	if !strings.Contains(result, "&lt;script&gt;") {
		t.Error("Expected script tags to be escaped")
	}
	
	if !strings.Contains(result, "&lt;img") {
		t.Error("Expected img tags to be escaped")
	}
}

func TestXSSPrevention_Integration_NoBypassFunctions(t *testing.T) {
	engine := NewEngine()
	
	dangerousFunctions := []string{
		"{{safeHTML .Content}}",
		"{{safeURL .URL}}",
		"{{safeCSS .CSS}}",
	}
	
	for _, funcCall := range dangerousFunctions {
		t.Run(funcCall, func(t *testing.T) {
			_, err := engine.Parse(funcCall)
			if err == nil {
				t.Errorf("Dangerous function %s should not be available", funcCall)
			}
		})
	}
}

func TestXSSPrevention_Integration_ContextAwareEscaping(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name     string
		template string
		data     interface{}
		mustHave []string
		mustNot  []string
	}{
		{
			name:     "HTML context",
			template: "<div>{{.Content}}</div>",
			data:     struct{ Content string }{Content: "<script>alert('xss')</script>"},
			mustHave: []string{"&lt;script&gt;", "&lt;/script&gt;"},
			mustNot:  []string{"<script>"},
		},
		{
			name:     "Attribute context",
			template: "<img alt='{{.Alt}}'>",
			data:     struct{ Alt string }{Alt: "\" onerror=\"alert('xss')"},
			mustHave: []string{"&#34;"},
			mustNot:  []string{"onerror=\"alert"},
		},
		{
			name:     "URL context",
			template: "<a href='{{.URL}}'>Link</a>",
			data:     struct{ URL string }{URL: "javascript:alert('xss')"},
			mustHave: []string{"#ZgotmplZ"},
			mustNot:  []string{"javascript:alert"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			result, err := engine.ExecuteToString(tmpl, tt.data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}
			
			for _, must := range tt.mustHave {
				if !strings.Contains(result, must) {
					t.Errorf("Expected output to contain %q, got: %s", must, result)
				}
			}
			
			for _, mustNot := range tt.mustNot {
				if strings.Contains(result, mustNot) {
					t.Errorf("Expected output to NOT contain %q, got: %s", mustNot, result)
				}
			}
		})
	}
}

func TestXSSPrevention_Integration_ServiceLevel(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine()
	
	maliciousTemplate := &models.Template{
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<h1>{{.Event.Title}}</h1><p>{{.Event.Description}}</p>",
	}
	
	desc := "<img src=x onerror=alert('xss')>"
	loc := "Test Location"
	maliciousEvent := &models.Event{
		Title:       "<script>alert('xss')</script>",
		Description: &desc,
		Location:    &loc,
		StartTime:   time.Now(),
	}
	
	data := struct {
		Event *models.Event
	}{
		Event: maliciousEvent,
	}
	
	tmpl, err := engine.Parse(maliciousTemplate.HTMLContent)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	_ = ctx
	
	if strings.Contains(result, "<script>") {
		t.Error("Service-level XSS prevention failed: script tag in output")
	}
	
	if strings.Contains(result, "onerror=") && !strings.Contains(result, "&lt;") {
		t.Error("Service-level XSS prevention failed: event handler in output")
	}
}
