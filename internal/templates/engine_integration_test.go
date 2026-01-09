package templates

import (
	"strings"
	"testing"
	"time"
)

type Event struct {
	ID          int64
	Title       string
	Description string
	Location    string
	StartTime   time.Time
	EndTime     time.Time
}

type Invite struct {
	ID        int64
	EventID   int64
	Email     string
	Name      string
	Token     string
	CreatedAt time.Time
}

type RSVP struct {
	ID         int64
	InviteID   int64
	Attending  bool
	GuestCount int
	Message    string
	CreatedAt  time.Time
}

func TestEngine_Integration_EventTemplate(t *testing.T) {
	engine := NewEngine()
	
	event := Event{
		ID:          1,
		Title:       "Summer BBQ Party",
		Description: "Join us for food, fun, and friends!",
		Location:    "123 Main Street, Anytown, USA",
		StartTime:   time.Date(2026, 7, 15, 17, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2026, 7, 15, 21, 0, 0, 0, time.UTC),
	}
	
	tmplStr := `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
</head>
<body>
	<h1>{{.Title | upper}}</h1>
	<p>{{.Description}}</p>
	<div>
		<strong>When:</strong> {{formatDate .StartTime "Monday, January 2, 2006"}} at {{formatTime .StartTime}}
	</div>
	<div>
		<strong>Where:</strong> {{.Location}}
	</div>
</body>
</html>
`
	
	tmpl, err := engine.Parse(tmplStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	result, err := engine.ExecuteToString(tmpl, event)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if !strings.Contains(result, "SUMMER BBQ PARTY") {
		t.Error("Expected uppercase title")
	}
	if !strings.Contains(result, "Join us for food, fun, and friends!") {
		t.Error("Expected description")
	}
	if !strings.Contains(result, "July 15, 2026") {
		t.Errorf("Expected formatted date, got: %s", result)
	}
	if !strings.Contains(result, "5:00 PM") {
		t.Errorf("Expected formatted time, got: %s", result)
	}
	if !strings.Contains(result, "123 Main Street") {
		t.Error("Expected location")
	}
}

func TestEngine_Integration_InviteEmailTemplate(t *testing.T) {
	engine := NewEngine()
	
	data := struct {
		Event  Event
		Invite Invite
	}{
		Event: Event{
			Title:     "Birthday Celebration",
			StartTime: time.Date(2026, 8, 20, 19, 0, 0, 0, time.UTC),
			Location:  "The Grand Hall",
		},
		Invite: Invite{
			Name:  "John Doe",
			Email: "john@example.com",
			Token: "abc123xyz",
		},
	}
	
	tmplStr := `
Dear {{.Invite.Name}},

You're invited to {{.Event.Title}}!

Event Details:
- Date: {{formatDateTime .Event.StartTime}}
- Location: {{.Event.Location}}

Please RSVP using your personal link:
https://example.com/rsvp/{{.Invite.Token}}

We hope to see you there!
`
	
	tmpl, err := engine.Parse(tmplStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if !strings.Contains(result, "Dear John Doe") {
		t.Error("Expected personalized greeting")
	}
	if !strings.Contains(result, "Birthday Celebration") {
		t.Error("Expected event title")
	}
	if !strings.Contains(result, "August 20, 2026 at 7:00 PM") {
		t.Errorf("Expected formatted date/time, got: %s", result)
	}
	if !strings.Contains(result, "https://example.com/rsvp/abc123xyz") {
		t.Error("Expected RSVP link with token")
	}
}

func TestEngine_Integration_RSVPConfirmationTemplate(t *testing.T) {
	engine := NewEngine()
	
	data := struct {
		Event Event
		RSVP  RSVP
		Name  string
	}{
		Event: Event{
			Title:     "Annual Gala",
			StartTime: time.Date(2026, 9, 10, 18, 30, 0, 0, time.UTC),
			Location:  "City Convention Center",
		},
		RSVP: RSVP{
			Attending:  true,
			GuestCount: 2,
			Message:    "Looking forward to it!",
		},
		Name: "Jane Smith",
	}
	
	tmplStr := `
<html>
<head><title>RSVP Confirmation</title></head>
<body>
	<h1>Thank You, {{.Name}}!</h1>
	<p>Your RSVP has been confirmed.</p>
	
	{{if .RSVP.Attending}}
	<div>
		<p><strong>You are attending:</strong> {{.Event.Title}}</p>
		<p><strong>Number of guests:</strong> {{.RSVP.GuestCount}}</p>
		<p><strong>Date:</strong> {{formatDate .Event.StartTime "January 2, 2006"}}</p>
		<p><strong>Time:</strong> {{formatTime .Event.StartTime}}</p>
		<p><strong>Location:</strong> {{.Event.Location}}</p>
		{{if .RSVP.Message}}
		<p><strong>Your message:</strong> {{.RSVP.Message}}</p>
		{{end}}
	</div>
	{{else}}
	<p>We're sorry you can't make it.</p>
	{{end}}
</body>
</html>
`
	
	tmpl, err := engine.Parse(tmplStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if !strings.Contains(result, "Thank You, Jane Smith!") {
		t.Error("Expected personalized thank you")
	}
	if !strings.Contains(result, "Annual Gala") {
		t.Errorf("Expected event title, got: %s", result)
	}
	if !strings.Contains(result, "2") {
		t.Errorf("Expected guest count, got: %s", result)
	}
	if !strings.Contains(result, "September 10, 2026") {
		t.Error("Expected formatted date")
	}
	if !strings.Contains(result, "6:30 PM") {
		t.Error("Expected formatted time")
	}
	if !strings.Contains(result, "Looking forward to it!") {
		t.Error("Expected custom message")
	}
}

func TestEngine_Integration_XSSInUserData(t *testing.T) {
	engine := NewEngine()
	
	event := Event{
		Title:       "<script>alert('xss')</script>Malicious Event",
		Description: "<img src=x onerror=alert('xss')>",
		Location:    "javascript:alert('xss')",
	}
	
	tmplStr := `
<html>
<body>
	<h1>{{.Title}}</h1>
	<p>{{.Description}}</p>
	<p>Location: {{.Location}}</p>
</body>
</html>
`
	
	tmpl, err := engine.Parse(tmplStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	result, err := engine.ExecuteToString(tmpl, event)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if strings.Contains(result, "<script>") && !strings.Contains(result, "&lt;script&gt;") {
		t.Error("XSS not prevented in title")
	}
	if strings.Contains(result, "onerror=") && !strings.Contains(result, "onerror=alert") {
		t.Error("XSS not prevented in description")
	}
	if strings.Contains(result, "javascript:alert") && !strings.Contains(result, "&#") {
		t.Error("XSS not prevented in location")
	}
	
	if !strings.Contains(result, "&lt;") || !strings.Contains(result, "&gt;") {
		t.Error("Expected HTML escaping")
	}
}

func TestEngine_Integration_ComplexNestedData(t *testing.T) {
	engine := NewEngine()
	
	data := struct {
		Events []Event
		Title  string
	}{
		Title: "Upcoming Events",
		Events: []Event{
			{
				Title:     "Event 1",
				StartTime: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
			},
			{
				Title:     "Event 2",
				StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
			},
			{
				Title:     "Event 3",
				StartTime: time.Date(2026, 7, 1, 18, 0, 0, 0, time.UTC),
			},
		},
	}
	
	tmplStr := `
<html>
<body>
	<h1>{{.Title | upper}}</h1>
	<ul>
	{{range .Events}}
		<li>
			<strong>{{.Title}}</strong> - {{formatDate .StartTime "Jan 2, 2006"}} at {{formatTime .StartTime}}
		</li>
	{{end}}
	</ul>
</body>
</html>
`
	
	tmpl, err := engine.Parse(tmplStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if !strings.Contains(result, "UPCOMING EVENTS") {
		t.Error("Expected uppercase title")
	}
	if !strings.Contains(result, "Event 1") || !strings.Contains(result, "Event 2") || !strings.Contains(result, "Event 3") {
		t.Error("Expected all events")
	}
	if !strings.Contains(result, "Jun 1, 2026") {
		t.Error("Expected formatted date for event 1")
	}
	if !strings.Contains(result, "10:00 AM") {
		t.Error("Expected formatted time for event 1")
	}
}

func TestEngine_Integration_TruncateAndDefault(t *testing.T) {
	engine := NewEngine()
	
	data := struct {
		LongText  string
		ShortText string
		EmptyText string
	}{
		LongText:  "This is a very long description that should be truncated to fit in the preview",
		ShortText: "Short",
		EmptyText: "",
	}
	
	tmplStr := `
<div>
	<p>Long: {{truncate .LongText 20}}</p>
	<p>Short: {{truncate .ShortText 20}}</p>
	<p>Empty: {{default .EmptyText "No description provided"}}</p>
</div>
`
	
	tmpl, err := engine.Parse(tmplStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if !strings.Contains(result, "This is a very long ...") {
		t.Error("Expected truncated long text")
	}
	if !strings.Contains(result, "Short: Short") {
		t.Error("Expected untruncated short text")
	}
	if !strings.Contains(result, "No description provided") {
		t.Error("Expected default value for empty text")
	}
}

func TestEngine_Integration_NoUnsafeFunctions(t *testing.T) {
	engine := NewEngine()
	
	dangerousFunctions := []struct {
		name     string
		template string
	}{
		{
			name:     "safeHTML",
			template: "<div>{{safeHTML .Content}}</div>",
		},
		{
			name:     "safeURL",
			template: "<a href='{{safeURL .URL}}'>Link</a>",
		},
		{
			name:     "safeCSS",
			template: "<style>{{safeCSS .CSS}}</style>",
		},
	}
	
	for _, tt := range dangerousFunctions {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Parse(tt.template)
			if err == nil {
				t.Errorf("Expected error when using dangerous function %s, but parsing succeeded", tt.name)
			}
			if err != nil && !strings.Contains(err.Error(), "not defined") {
				t.Errorf("Expected 'not defined' error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func BenchmarkEngine_Integration_ComplexTemplate(b *testing.B) {
	engine := NewEngine()
	
	data := struct {
		Event  Event
		Invite Invite
	}{
		Event: Event{
			Title:       "Test Event",
			Description: "Test Description",
			Location:    "Test Location",
			StartTime:   time.Now(),
		},
		Invite: Invite{
			Name:  "Test User",
			Email: "test@example.com",
			Token: "test123",
		},
	}
	
	tmplStr := `
<!DOCTYPE html>
<html>
<head><title>{{.Event.Title}}</title></head>
<body>
	<h1>{{.Event.Title | upper}}</h1>
	<p>{{.Event.Description}}</p>
	<p>Date: {{formatDateTime .Event.StartTime}}</p>
	<p>Location: {{.Event.Location}}</p>
	<p>Dear {{.Invite.Name}},</p>
	<p>RSVP: https://example.com/rsvp/{{.Invite.Token}}</p>
</body>
</html>
`
	
	tmpl, _ := engine.Parse(tmplStr)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ExecuteToString(tmpl, data)
	}
}
