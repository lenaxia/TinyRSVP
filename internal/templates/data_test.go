package templates

import (
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestInviteEmailData_Structure(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 6, 10, 23, 59, 59, 0, time.UTC)

	data := &InviteEmailData{
		RSVPURL:     "https://example.com/rsvp/token123",
		MaxPlusOnes: 2,
	}
	data.Event.Title = "Birthday Party"
	data.Event.Description = "Join us for cake!"
	data.Event.StartTime = startTime
	data.Event.EndTime = &endTime
	data.Event.Timezone = "America/Los_Angeles"
	data.Event.Location = "123 Main St"
	data.Event.RSVPDeadline = &deadline
	data.Invite.Name = "John Doe"
	data.Invite.Email = "john@example.com"

	if data.Event.Title != "Birthday Party" {
		t.Errorf("Expected Title 'Birthday Party', got '%s'", data.Event.Title)
	}
	if data.Event.Description != "Join us for cake!" {
		t.Errorf("Expected Description 'Join us for cake!', got '%s'", data.Event.Description)
	}
	if !data.Event.StartTime.Equal(startTime) {
		t.Errorf("Expected StartTime %v, got %v", startTime, data.Event.StartTime)
	}
	if data.Event.EndTime == nil || !data.Event.EndTime.Equal(endTime) {
		t.Errorf("Expected EndTime %v, got %v", endTime, data.Event.EndTime)
	}
	if data.Event.Timezone != "America/Los_Angeles" {
		t.Errorf("Expected Timezone 'America/Los_Angeles', got '%s'", data.Event.Timezone)
	}
	if data.Event.Location != "123 Main St" {
		t.Errorf("Expected Location '123 Main St', got '%s'", data.Event.Location)
	}
	if data.Event.RSVPDeadline == nil || !data.Event.RSVPDeadline.Equal(deadline) {
		t.Errorf("Expected RSVPDeadline %v, got %v", deadline, data.Event.RSVPDeadline)
	}
	if data.Invite.Name != "John Doe" {
		t.Errorf("Expected Invite.Name 'John Doe', got '%s'", data.Invite.Name)
	}
	if data.Invite.Email != "john@example.com" {
		t.Errorf("Expected Invite.Email 'john@example.com', got '%s'", data.Invite.Email)
	}
	if data.RSVPURL != "https://example.com/rsvp/token123" {
		t.Errorf("Expected RSVPURL 'https://example.com/rsvp/token123', got '%s'", data.RSVPURL)
	}
	if data.MaxPlusOnes != 2 {
		t.Errorf("Expected MaxPlusOnes 2, got %d", data.MaxPlusOnes)
	}
}

func TestRSVPPageData_Structure(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 6, 10, 23, 59, 59, 0, time.UTC)

	data := &RSVPPageData{
		Token:       "token123",
		MaxPlusOnes: 2,
		Questions: []QuestionData{
			{
				ID:           1,
				QuestionText: "Dietary restrictions?",
				QuestionType: "text",
				Required:     false,
				HelpText:     "Let us know",
				Options: []OptionData{
					{Value: "vegan", Label: "Vegan"},
					{Value: "vegetarian", Label: "Vegetarian"},
				},
			},
		},
	}
	data.Event.Title = "Birthday Party"
	data.Event.Description = "Join us for cake!"
	data.Event.StartTime = startTime
	data.Event.EndTime = &endTime
	data.Event.Timezone = "America/Los_Angeles"
	data.Event.Location = "123 Main St"
	data.Event.RSVPDeadline = &deadline

	if data.Event.Title != "Birthday Party" {
		t.Errorf("Expected Title 'Birthday Party', got '%s'", data.Event.Title)
	}
	if data.Token != "token123" {
		t.Errorf("Expected Token 'token123', got '%s'", data.Token)
	}
	if data.MaxPlusOnes != 2 {
		t.Errorf("Expected MaxPlusOnes 2, got %d", data.MaxPlusOnes)
	}
	if len(data.Questions) != 1 {
		t.Fatalf("Expected 1 question, got %d", len(data.Questions))
	}
	if data.Questions[0].ID != 1 {
		t.Errorf("Expected Question ID 1, got %d", data.Questions[0].ID)
	}
	if data.Questions[0].QuestionText != "Dietary restrictions?" {
		t.Errorf("Expected QuestionText 'Dietary restrictions?', got '%s'", data.Questions[0].QuestionText)
	}
	if data.Questions[0].QuestionType != "text" {
		t.Errorf("Expected QuestionType 'text', got '%s'", data.Questions[0].QuestionType)
	}
	if data.Questions[0].Required {
		t.Error("Expected Required false, got true")
	}
	if data.Questions[0].HelpText != "Let us know" {
		t.Errorf("Expected HelpText 'Let us know', got '%s'", data.Questions[0].HelpText)
	}
	if len(data.Questions[0].Options) != 2 {
		t.Fatalf("Expected 2 options, got %d", len(data.Questions[0].Options))
	}
	if data.Questions[0].Options[0].Value != "vegan" {
		t.Errorf("Expected Option Value 'vegan', got '%s'", data.Questions[0].Options[0].Value)
	}
	if data.Questions[0].Options[0].Label != "Vegan" {
		t.Errorf("Expected Option Label 'Vegan', got '%s'", data.Questions[0].Options[0].Label)
	}
}

func TestConfirmationPageData_Structure(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC)

	data := &ConfirmationPageData{
		Token: "token123",
		Answers: []AnswerData{
			{
				QuestionText:  "Dietary restrictions?",
				AnswerDisplay: "Vegan",
			},
		},
	}
	data.Event.Title = "Birthday Party"
	data.Event.Description = "Join us for cake!"
	data.Event.StartTime = startTime
	data.Event.EndTime = &endTime
	data.Event.Timezone = "America/Los_Angeles"
	data.Event.Location = "123 Main St"
	data.RSVP.Response = "yes"
	data.RSVP.PlusOnes = 2
	data.RSVP.Notes = "Looking forward to it!"

	if data.Event.Title != "Birthday Party" {
		t.Errorf("Expected Title 'Birthday Party', got '%s'", data.Event.Title)
	}
	if data.Token != "token123" {
		t.Errorf("Expected Token 'token123', got '%s'", data.Token)
	}
	if data.RSVP.Response != "yes" {
		t.Errorf("Expected RSVP.Response 'yes', got '%s'", data.RSVP.Response)
	}
	if data.RSVP.PlusOnes != 2 {
		t.Errorf("Expected RSVP.PlusOnes 2, got %d", data.RSVP.PlusOnes)
	}
	if data.RSVP.Notes != "Looking forward to it!" {
		t.Errorf("Expected RSVP.Notes 'Looking forward to it!', got '%s'", data.RSVP.Notes)
	}
	if len(data.Answers) != 1 {
		t.Fatalf("Expected 1 answer, got %d", len(data.Answers))
	}
	if data.Answers[0].QuestionText != "Dietary restrictions?" {
		t.Errorf("Expected QuestionText 'Dietary restrictions?', got '%s'", data.Answers[0].QuestionText)
	}
	if data.Answers[0].AnswerDisplay != "Vegan" {
		t.Errorf("Expected AnswerDisplay 'Vegan', got '%s'", data.Answers[0].AnswerDisplay)
	}
}

func TestBuildInviteEmailData(t *testing.T) {
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

	if data == nil {
		t.Fatal("Expected non-nil data")
	}
	if data.Event.Title != "Birthday Party" {
		t.Errorf("Expected Title 'Birthday Party', got '%s'", data.Event.Title)
	}
	if data.Event.Description != "Join us for cake!" {
		t.Errorf("Expected Description 'Join us for cake!', got '%s'", data.Event.Description)
	}
	if !data.Event.StartTime.Equal(startTime) {
		t.Errorf("Expected StartTime %v, got %v", startTime, data.Event.StartTime)
	}
	if data.Event.EndTime == nil || !data.Event.EndTime.Equal(endTime) {
		t.Errorf("Expected EndTime %v, got %v", endTime, data.Event.EndTime)
	}
	if data.Event.Timezone != "America/Los_Angeles" {
		t.Errorf("Expected Timezone 'America/Los_Angeles', got '%s'", data.Event.Timezone)
	}
	if data.Event.Location != "123 Main St" {
		t.Errorf("Expected Location '123 Main St', got '%s'", data.Event.Location)
	}
	if data.Event.RSVPDeadline == nil || !data.Event.RSVPDeadline.Equal(deadline) {
		t.Errorf("Expected RSVPDeadline %v, got %v", deadline, data.Event.RSVPDeadline)
	}
	if data.Invite.Name != "John Doe" {
		t.Errorf("Expected Invite.Name 'John Doe', got '%s'", data.Invite.Name)
	}
	if data.Invite.Email != "john@example.com" {
		t.Errorf("Expected Invite.Email 'john@example.com', got '%s'", data.Invite.Email)
	}
	if data.RSVPURL != rsvpURL {
		t.Errorf("Expected RSVPURL '%s', got '%s'", rsvpURL, data.RSVPURL)
	}
	if data.MaxPlusOnes != 2 {
		t.Errorf("Expected MaxPlusOnes 2, got %d", data.MaxPlusOnes)
	}
}

func TestBuildInviteEmailData_NilFields(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)

	event := &models.Event{
		ID:        1,
		Title:     "Birthday Party",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
	}

	invite := &models.Invite{
		ID:          1,
		EventID:     1,
		MaxPlusOnes: 0,
	}

	rsvpURL := "https://example.com/rsvp/token123"

	data := BuildInviteEmailData(event, invite, rsvpURL)

	if data == nil {
		t.Fatal("Expected non-nil data")
	}
	if data.Event.Description != "" {
		t.Errorf("Expected empty Description, got '%s'", data.Event.Description)
	}
	if data.Event.EndTime != nil {
		t.Errorf("Expected nil EndTime, got %v", data.Event.EndTime)
	}
	if data.Event.Location != "" {
		t.Errorf("Expected empty Location, got '%s'", data.Event.Location)
	}
	if data.Event.RSVPDeadline != nil {
		t.Errorf("Expected nil RSVPDeadline, got %v", data.Event.RSVPDeadline)
	}
	if data.Invite.Name != "" {
		t.Errorf("Expected empty Invite.Name, got '%s'", data.Invite.Name)
	}
	if data.Invite.Email != "" {
		t.Errorf("Expected empty Invite.Email, got '%s'", data.Invite.Email)
	}
	if data.MaxPlusOnes != 0 {
		t.Errorf("Expected MaxPlusOnes 0, got %d", data.MaxPlusOnes)
	}
}

func TestBuildRSVPPageData(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 6, 10, 23, 59, 59, 0, time.UTC)
	description := "Join us for cake!"
	location := "123 Main St"
	name := "John Doe"
	email := "john@example.com"
	helpText := "Let us know"
	optionsJSON := `["vegan","vegetarian","none"]`

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

	if data == nil {
		t.Fatal("Expected non-nil data")
	}
	if data.Event.Title != "Birthday Party" {
		t.Errorf("Expected Title 'Birthday Party', got '%s'", data.Event.Title)
	}
	if data.Token != token {
		t.Errorf("Expected Token '%s', got '%s'", token, data.Token)
	}
	if data.MaxPlusOnes != 2 {
		t.Errorf("Expected MaxPlusOnes 2, got %d", data.MaxPlusOnes)
	}
	if len(data.Questions) != 2 {
		t.Fatalf("Expected 2 questions, got %d", len(data.Questions))
	}

	q1 := data.Questions[0]
	if q1.ID != 1 {
		t.Errorf("Expected Question 1 ID 1, got %d", q1.ID)
	}
	if q1.QuestionText != "Dietary restrictions?" {
		t.Errorf("Expected QuestionText 'Dietary restrictions?', got '%s'", q1.QuestionText)
	}
	if q1.QuestionType != "text" {
		t.Errorf("Expected QuestionType 'text', got '%s'", q1.QuestionType)
	}
	if q1.Required {
		t.Error("Expected Required false, got true")
	}
	if q1.HelpText != "Let us know" {
		t.Errorf("Expected HelpText 'Let us know', got '%s'", q1.HelpText)
	}
	if len(q1.Options) != 0 {
		t.Errorf("Expected 0 options for text question, got %d", len(q1.Options))
	}

	q2 := data.Questions[1]
	if q2.ID != 2 {
		t.Errorf("Expected Question 2 ID 2, got %d", q2.ID)
	}
	if q2.QuestionText != "Meal preference?" {
		t.Errorf("Expected QuestionText 'Meal preference?', got '%s'", q2.QuestionText)
	}
	if q2.QuestionType != "single_choice" {
		t.Errorf("Expected QuestionType 'single_choice', got '%s'", q2.QuestionType)
	}
	if !q2.Required {
		t.Error("Expected Required true, got false")
	}
	if q2.HelpText != "" {
		t.Errorf("Expected empty HelpText, got '%s'", q2.HelpText)
	}
	if len(q2.Options) != 3 {
		t.Fatalf("Expected 3 options, got %d", len(q2.Options))
	}
	if q2.Options[0].Value != "vegan" || q2.Options[0].Label != "vegan" {
		t.Errorf("Expected option 0 vegan/vegan, got %s/%s", q2.Options[0].Value, q2.Options[0].Label)
	}
	if q2.Options[1].Value != "vegetarian" || q2.Options[1].Label != "vegetarian" {
		t.Errorf("Expected option 1 vegetarian/vegetarian, got %s/%s", q2.Options[1].Value, q2.Options[1].Label)
	}
	if q2.Options[2].Value != "none" || q2.Options[2].Label != "none" {
		t.Errorf("Expected option 2 none/none, got %s/%s", q2.Options[2].Value, q2.Options[2].Label)
	}
}

func TestBuildRSVPPageData_NoQuestions(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	description := "Join us for cake!"
	name := "John Doe"

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
		MaxPlusOnes: 0,
	}

	data := BuildRSVPPageData(event, invite, nil, "token123")

	if data == nil {
		t.Fatal("Expected non-nil data")
	}
	if len(data.Questions) != 0 {
		t.Errorf("Expected 0 questions, got %d", len(data.Questions))
	}
}

func TestBuildConfirmationPageData(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC)
	description := "Join us for cake!"
	location := "123 Main St"
	answerText := "Vegan"
	answerOption := "chicken"
	answerBool := true

	event := &models.Event{
		ID:          1,
		Title:       "Birthday Party",
		Description: &description,
		StartTime:   startTime,
		EndTime:     &endTime,
		Timezone:    "America/Los_Angeles",
		Location:    &location,
	}

	rsvp := &models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 2,
	}

	questions := map[int64]*models.PreferenceQuestion{
		1: {ID: 1, QuestionText: "Dietary restrictions?"},
		2: {ID: 2, QuestionText: "Meal preference?"},
		3: {ID: 3, QuestionText: "Attending ceremony?"},
	}

	answers := []*models.RSVPAnswer{
		{
			ID:         1,
			RSVPID:     1,
			QuestionID: 1,
			AnswerText: &answerText,
		},
		{
			ID:           2,
			RSVPID:       1,
			QuestionID:   2,
			AnswerOption: &answerOption,
		},
		{
			ID:            3,
			RSVPID:        1,
			QuestionID:    3,
			AnswerBoolean: &answerBool,
		},
	}

	token := "token123"

	data := BuildConfirmationPageData(event, rsvp, answers, questions, token)

	if data == nil {
		t.Fatal("Expected non-nil data")
	}
	if data.Event.Title != "Birthday Party" {
		t.Errorf("Expected Title 'Birthday Party', got '%s'", data.Event.Title)
	}
	if data.Token != token {
		t.Errorf("Expected Token '%s', got '%s'", token, data.Token)
	}
	if data.RSVP.Response != "yes" {
		t.Errorf("Expected RSVP.Response 'yes', got '%s'", data.RSVP.Response)
	}
	if data.RSVP.PlusOnes != 2 {
		t.Errorf("Expected RSVP.PlusOnes 2, got %d", data.RSVP.PlusOnes)
	}
	if len(data.Answers) != 3 {
		t.Fatalf("Expected 3 answers, got %d", len(data.Answers))
	}
	if data.Answers[0].QuestionText != "Dietary restrictions?" {
		t.Errorf("Expected QuestionText 'Dietary restrictions?', got '%s'", data.Answers[0].QuestionText)
	}
	if data.Answers[0].AnswerDisplay != "Vegan" {
		t.Errorf("Expected Answer 0 'Vegan', got '%s'", data.Answers[0].AnswerDisplay)
	}
	if data.Answers[1].QuestionText != "Meal preference?" {
		t.Errorf("Expected QuestionText 'Meal preference?', got '%s'", data.Answers[1].QuestionText)
	}
	if data.Answers[1].AnswerDisplay != "chicken" {
		t.Errorf("Expected Answer 1 'chicken', got '%s'", data.Answers[1].AnswerDisplay)
	}
	if data.Answers[2].QuestionText != "Attending ceremony?" {
		t.Errorf("Expected QuestionText 'Attending ceremony?', got '%s'", data.Answers[2].QuestionText)
	}
	if data.Answers[2].AnswerDisplay != "Yes" {
		t.Errorf("Expected Answer 2 'Yes', got '%s'", data.Answers[2].AnswerDisplay)
	}
}

func TestBuildConfirmationPageData_NoAnswers(t *testing.T) {
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	description := "Join us for cake!"

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
		Response: models.RSVPResponseNo,
		PlusOnes: 0,
	}

	questions := map[int64]*models.PreferenceQuestion{}

	data := BuildConfirmationPageData(event, rsvp, nil, questions, "token123")

	if data == nil {
		t.Fatal("Expected non-nil data")
	}
	if data.RSVP.Response != "no" {
		t.Errorf("Expected RSVP.Response 'no', got '%s'", data.RSVP.Response)
	}
	if len(data.Answers) != 0 {
		t.Errorf("Expected 0 answers, got %d", len(data.Answers))
	}
}

func TestFormatAnswer_Text(t *testing.T) {
	text := "Vegan"
	answer := &models.RSVPAnswer{
		AnswerText: &text,
	}

	result := formatAnswer(answer)
	if result != "Vegan" {
		t.Errorf("Expected 'Vegan', got '%s'", result)
	}
}

func TestFormatAnswer_Option(t *testing.T) {
	option := "chicken"
	answer := &models.RSVPAnswer{
		AnswerOption: &option,
	}

	result := formatAnswer(answer)
	if result != "chicken" {
		t.Errorf("Expected 'chicken', got '%s'", result)
	}
}

func TestFormatAnswer_BooleanTrue(t *testing.T) {
	boolVal := true
	answer := &models.RSVPAnswer{
		AnswerBoolean: &boolVal,
	}

	result := formatAnswer(answer)
	if result != "Yes" {
		t.Errorf("Expected 'Yes', got '%s'", result)
	}
}

func TestFormatAnswer_BooleanFalse(t *testing.T) {
	boolVal := false
	answer := &models.RSVPAnswer{
		AnswerBoolean: &boolVal,
	}

	result := formatAnswer(answer)
	if result != "No" {
		t.Errorf("Expected 'No', got '%s'", result)
	}
}

func TestFormatAnswer_Empty(t *testing.T) {
	answer := &models.RSVPAnswer{}

	result := formatAnswer(answer)
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}
