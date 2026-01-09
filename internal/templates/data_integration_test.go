package templates

import (
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestDataIntegration_InviteEmail(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 6, 10, 23, 59, 59, 0, time.UTC)
	description := "Join us for cake!"
	location := "123 Main St"
	name := "John Doe"
	email := "john@example.com"

	event := &models.Event{
		ID:           1,
		Title:        "Birthday Party",
		Description:  &description,
		StartTime:    startTime,
		EndTime:      &endTime,
		Timezone:     "America/Los_Angeles",
		Location:     &location,
		RSVPDeadline: &deadline,
	}

	invite := &models.Invite{
		ID:          1,
		EventID:     1,
		Name:        &name,
		Email:       &email,
		MaxPlusOnes: 2,
	}

	rsvpURL := "https://example.com/rsvp/token123"

	data := BuildInviteEmailData(event, invite, rsvpURL)

	engine := NewEngine()
	tmplContent := `
<html>
<body>
	<h1>{{.Event.Title}}</h1>
	<p>{{.Event.Description}}</p>
	<p>When: {{.Event.StartTime}}</p>
	{{if .Event.EndTime}}<p>Until: {{.Event.EndTime}}</p>{{end}}
	<p>Timezone: {{.Event.Timezone}}</p>
	{{if .Event.Location}}<p>Where: {{.Event.Location}}</p>{{end}}
	{{if .Event.RSVPDeadline}}<p>RSVP by: {{.Event.RSVPDeadline}}</p>{{end}}
	<p>Dear {{.Invite.Name}},</p>
	<p>Email: {{.Invite.Email}}</p>
	<p>You can bring up to {{.MaxPlusOnes}} guests</p>
	<a href="{{.RSVPURL}}">RSVP Here</a>
</body>
</html>
`

	tmpl, err := engine.Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if !strings.Contains(result, "Birthday Party") {
		t.Error("Expected result to contain 'Birthday Party'")
	}
	if !strings.Contains(result, "John Doe") {
		t.Error("Expected result to contain 'John Doe'")
	}
	if !strings.Contains(result, rsvpURL) {
		t.Error("Expected result to contain RSVP URL")
	}
}

func TestDataIntegration_RSVPPage(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	description := "Join us for cake!"
	name := "John Doe"
	helpText := "Let us know"
	optionsJSON := `["vegan","vegetarian"]`

	event := &models.Event{
		ID:          1,
		Title:       "Birthday Party",
		Description: &description,
		StartTime:   startTime,
		Timezone:    "America/Los_Angeles",
	}

	invite := &models.Invite{
		ID:          1,
		EventID:     1,
		Name:        &name,
		MaxPlusOnes: 2,
	}

	questions := []*models.PreferenceQuestion{
		{
			ID:           1,
			EventID:      1,
			QuestionText: "Dietary restrictions?",
			QuestionType: models.QuestionTypeText,
			Required:     false,
			HelpText:     &helpText,
		},
		{
			ID:           2,
			EventID:      1,
			QuestionText: "Meal preference?",
			QuestionType: models.QuestionTypeSingleChoice,
			Required:     true,
			Options:      &optionsJSON,
		},
	}

	token := "token123"

	data := BuildRSVPPageData(event, invite, questions, token)

	engine := NewEngine()
	tmplContent := `
<html>
<body>
	<h1>{{.Event.Title}}</h1>
	<p>{{.Event.Description}}</p>
	<p>Token: {{.Token}}</p>
	<p>Max guests: {{.MaxPlusOnes}}</p>
	{{range .Questions}}
	<div>
		<label>{{.QuestionText}}{{if .Required}}*{{end}}</label>
		{{if .HelpText}}<small>{{.HelpText}}</small>{{end}}
		{{if eq .QuestionType "text"}}
		<input type="text" name="q_{{.ID}}">
		{{else if eq .QuestionType "single_choice"}}
		{{$qid := .ID}}
		{{range .Options}}
		<label><input type="radio" name="q_{{$qid}}" value="{{.Value}}">{{.Label}}</label>
		{{end}}
		{{end}}
	</div>
	{{end}}
</body>
</html>
`

	tmpl, err := engine.Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if !strings.Contains(result, "Birthday Party") {
		t.Error("Expected result to contain 'Birthday Party'")
	}
	if !strings.Contains(result, "Dietary restrictions?") {
		t.Error("Expected result to contain 'Dietary restrictions?'")
	}
	if !strings.Contains(result, "Meal preference?") {
		t.Error("Expected result to contain 'Meal preference?'")
	}
	if !strings.Contains(result, "vegan") {
		t.Error("Expected result to contain 'vegan'")
	}
}

func TestDataIntegration_ConfirmationPage(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	description := "Join us for cake!"
	answerText := "Vegan"

	event := &models.Event{
		ID:          1,
		Title:       "Birthday Party",
		Description: &description,
		StartTime:   startTime,
		Timezone:    "America/Los_Angeles",
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 2,
	}

	questions := map[int64]*models.PreferenceQuestion{
		1: {ID: 1, QuestionText: "Dietary restrictions?"},
	}

	answers := []*models.RSVPAnswer{
		{
			ID:         1,
			RSVPID:     1,
			QuestionID: 1,
			AnswerText: &answerText,
		},
	}

	token := "token123"

	data := BuildConfirmationPageData(event, rsvp, answers, questions, token)

	engine := NewEngine()
	tmplContent := `
<html>
<body>
	<h1>{{.Event.Title}}</h1>
	<p>{{.Event.Description}}</p>
	<p>Token: {{.Token}}</p>
	<p>Response: {{.RSVP.Response}}</p>
	<p>Guests: {{.RSVP.PlusOnes}}</p>
	{{range .Answers}}
	<div>
		<strong>{{.QuestionText}}</strong>: {{.AnswerDisplay}}
	</div>
	{{end}}
</body>
</html>
`

	tmpl, err := engine.Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if !strings.Contains(result, "Birthday Party") {
		t.Error("Expected result to contain 'Birthday Party'")
	}
	if !strings.Contains(result, "yes") {
		t.Error("Expected result to contain 'yes'")
	}
	if !strings.Contains(result, "Dietary restrictions?") {
		t.Error("Expected result to contain 'Dietary restrictions?'")
	}
	if !strings.Contains(result, "Vegan") {
		t.Error("Expected result to contain 'Vegan'")
	}
}

func TestDataIntegration_ValidatorCompatibility(t *testing.T) {
	engine := NewEngine()
	validator := NewValidator(engine)

	t.Run("InviteEmail variables match whitelist", func(t *testing.T) {
		tmplContent := `
{{.Event.Title}}
{{.Event.Description}}
{{.Event.StartTime}}
{{.Event.EndTime}}
{{.Event.Timezone}}
{{.Event.Location}}
{{.Event.RSVPDeadline}}
{{.Invite.Name}}
{{.Invite.Email}}
{{.RSVPURL}}
{{.MaxPlusOnes}}
`
		err := validator.ValidateSyntax(tmplContent, models.TemplateTypeInviteEmail)
		if err != nil {
			t.Errorf("InviteEmail template validation failed: %v", err)
		}

		allowedVars := getAllowedVariables(models.TemplateTypeInviteEmail)
		err = validator.ValidateVariables(tmplContent, allowedVars)
		if err != nil {
			t.Errorf("InviteEmail variable validation failed: %v", err)
		}
	})

	t.Run("RSVPPage variables match whitelist", func(t *testing.T) {
		tmplContent := `
{{.Event.Title}}
{{.Event.Description}}
{{.Event.StartTime}}
{{.Event.EndTime}}
{{.Event.Timezone}}
{{.Event.Location}}
{{.Event.RSVPDeadline}}
{{.Token}}
{{.MaxPlusOnes}}
{{range .Questions}}
{{.ID}}
{{.QuestionText}}
{{.QuestionType}}
{{.Required}}
{{.HelpText}}
{{range .Options}}
{{.Value}}
{{.Label}}
{{end}}
{{end}}
`
		err := validator.ValidateSyntax(tmplContent, models.TemplateTypeRSVPPage)
		if err != nil {
			t.Errorf("RSVPPage template validation failed: %v", err)
		}

		allowedVars := getAllowedVariables(models.TemplateTypeRSVPPage)
		err = validator.ValidateVariables(tmplContent, allowedVars)
		if err != nil {
			t.Errorf("RSVPPage variable validation failed: %v", err)
		}
	})

	t.Run("ConfirmationPage variables match whitelist", func(t *testing.T) {
		tmplContent := `
{{.Event.Title}}
{{.Event.Description}}
{{.Event.StartTime}}
{{.Event.EndTime}}
{{.Event.Timezone}}
{{.Event.Location}}
{{.Token}}
{{.RSVP.Response}}
{{.RSVP.PlusOnes}}
{{.RSVP.Notes}}
{{range .Answers}}
{{.QuestionText}}
{{.AnswerDisplay}}
{{end}}
`
		err := validator.ValidateSyntax(tmplContent, models.TemplateTypeConfirmationPage)
		if err != nil {
			t.Errorf("ConfirmationPage template validation failed: %v", err)
		}

		allowedVars := getAllowedVariables(models.TemplateTypeConfirmationPage)
		err = validator.ValidateVariables(tmplContent, allowedVars)
		if err != nil {
			t.Errorf("ConfirmationPage variable validation failed: %v", err)
		}
	})
}
