package templates

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestCreateTestData_InviteEmail(t *testing.T) {
	data := CreateTestData(models.TemplateTypeInviteEmail)

	if data == nil {
		t.Fatal("Expected non-nil data for invite_email")
	}

	inviteData, ok := data.(*InviteEmailData)
	if !ok {
		t.Fatalf("Expected *InviteEmailData, got %T", data)
	}

	if inviteData.Event.Title == "" {
		t.Error("Expected non-empty event title")
	}

	if inviteData.Event.Description == "" {
		t.Error("Expected non-empty event description")
	}

	if inviteData.Event.StartTime.IsZero() {
		t.Error("Expected non-zero start time")
	}

	if inviteData.Event.StartTime.Before(time.Now()) {
		t.Error("Expected start time to be in the future")
	}

	if inviteData.Event.EndTime == nil {
		t.Error("Expected non-nil end time")
	}

	if inviteData.Event.Timezone == "" {
		t.Error("Expected non-empty timezone")
	}

	if inviteData.Event.Location == "" {
		t.Error("Expected non-empty location")
	}

	if inviteData.Event.RSVPDeadline == nil {
		t.Error("Expected non-nil RSVP deadline")
	}

	if inviteData.Invite.Name == "" {
		t.Error("Expected non-empty invite name")
	}

	if inviteData.Invite.Email == "" {
		t.Error("Expected non-empty invite email")
	}

	if inviteData.RSVPURL == "" {
		t.Error("Expected non-empty RSVP URL")
	}

	if inviteData.MaxPlusOnes < 0 {
		t.Error("Expected non-negative max plus ones")
	}
}

func TestCreateTestData_RSVPPage(t *testing.T) {
	data := CreateTestData(models.TemplateTypeRSVPPage)

	if data == nil {
		t.Fatal("Expected non-nil data for rsvp_page")
	}

	rsvpData, ok := data.(*RSVPPageData)
	if !ok {
		t.Fatalf("Expected *RSVPPageData, got %T", data)
	}

	if rsvpData.Event.Title == "" {
		t.Error("Expected non-empty event title")
	}

	if rsvpData.Event.Description == "" {
		t.Error("Expected non-empty event description")
	}

	if rsvpData.Event.StartTime.IsZero() {
		t.Error("Expected non-zero start time")
	}

	if rsvpData.Event.EndTime == nil {
		t.Error("Expected non-nil end time")
	}

	if rsvpData.Event.Timezone == "" {
		t.Error("Expected non-empty timezone")
	}

	if rsvpData.Event.Location == "" {
		t.Error("Expected non-empty location")
	}

	if rsvpData.Event.RSVPDeadline == nil {
		t.Error("Expected non-nil RSVP deadline")
	}

	if rsvpData.Token == "" {
		t.Error("Expected non-empty token")
	}

	if rsvpData.MaxPlusOnes < 0 {
		t.Error("Expected non-negative max plus ones")
	}

	if len(rsvpData.Questions) == 0 {
		t.Error("Expected at least one question")
	}

	for i, q := range rsvpData.Questions {
		if q.ID == 0 {
			t.Errorf("Question %d: expected non-zero ID", i)
		}
		if q.QuestionText == "" {
			t.Errorf("Question %d: expected non-empty question text", i)
		}
		if q.QuestionType == "" {
			t.Errorf("Question %d: expected non-empty question type", i)
		}
	}
}

func TestCreateTestData_ConfirmationPage(t *testing.T) {
	data := CreateTestData(models.TemplateTypeConfirmationPage)

	if data == nil {
		t.Fatal("Expected non-nil data for confirmation_page")
	}

	confirmData, ok := data.(*ConfirmationPageData)
	if !ok {
		t.Fatalf("Expected *ConfirmationPageData, got %T", data)
	}

	if confirmData.Event.Title == "" {
		t.Error("Expected non-empty event title")
	}

	if confirmData.Event.Description == "" {
		t.Error("Expected non-empty event description")
	}

	if confirmData.Event.StartTime.IsZero() {
		t.Error("Expected non-zero start time")
	}

	if confirmData.Event.EndTime == nil {
		t.Error("Expected non-nil end time")
	}

	if confirmData.Event.Timezone == "" {
		t.Error("Expected non-empty timezone")
	}

	if confirmData.Event.Location == "" {
		t.Error("Expected non-empty location")
	}

	if confirmData.Token == "" {
		t.Error("Expected non-empty token")
	}

	if confirmData.RSVP.Response == "" {
		t.Error("Expected non-empty RSVP response")
	}

	if confirmData.RSVP.PlusOnes < 0 {
		t.Error("Expected non-negative plus ones")
	}

	if len(confirmData.Answers) == 0 {
		t.Error("Expected at least one answer")
	}

	for i, ans := range confirmData.Answers {
		if ans.QuestionText == "" {
			t.Errorf("Answer %d: expected non-empty question text", i)
		}
		if ans.AnswerDisplay == "" {
			t.Errorf("Answer %d: expected non-empty answer display", i)
		}
	}
}

func TestCreateTestData_InvalidType(t *testing.T) {
	data := CreateTestData(models.TemplateType("invalid"))

	if data != nil {
		t.Error("Expected nil data for invalid template type")
	}
}

func TestCreateTestData_AllTypesReturnUniqueData(t *testing.T) {
	types := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, templateType := range types {
		t.Run(string(templateType), func(t *testing.T) {
			data1 := CreateTestData(templateType)
			data2 := CreateTestData(templateType)

			if data1 == nil || data2 == nil {
				t.Fatal("Expected non-nil data")
			}

			if data1 == data2 {
				t.Error("Expected different instances for each call")
			}
		})
	}
}

func TestPreviewTemplate_ValidInviteEmail(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	req := &PreviewRequest{
		Type:        models.TemplateTypeInviteEmail,
		HTMLContent: "<h1>{{.Event.Title}}</h1><p>{{.Invite.Name}}</p>",
	}

	resp, err := service.PreviewTemplate(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.HTMLPreview == "" {
		t.Error("Expected non-empty HTML preview")
	}

	if !strings.Contains(resp.HTMLPreview, "Sample Event") {
		t.Error("Expected HTML preview to contain event title")
	}

	if !strings.Contains(resp.HTMLPreview, "John Doe") {
		t.Error("Expected HTML preview to contain invite name")
	}
}

func TestPreviewTemplate_ValidRSVPPage(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	req := &PreviewRequest{
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<h1>{{.Event.Title}}</h1><p>Token: {{.Token}}</p>",
	}

	resp, err := service.PreviewTemplate(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.HTMLPreview == "" {
		t.Error("Expected non-empty HTML preview")
	}

	if !strings.Contains(resp.HTMLPreview, "Sample Event") {
		t.Error("Expected HTML preview to contain event title")
	}

	if !strings.Contains(resp.HTMLPreview, "sample-token-preview") {
		t.Error("Expected HTML preview to contain token")
	}
}

func TestPreviewTemplate_ValidConfirmationPage(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	req := &PreviewRequest{
		Type:        models.TemplateTypeConfirmationPage,
		HTMLContent: "<h1>{{.Event.Title}}</h1><p>Response: {{.RSVP.Response}}</p>",
	}

	resp, err := service.PreviewTemplate(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.HTMLPreview == "" {
		t.Error("Expected non-empty HTML preview")
	}

	if !strings.Contains(resp.HTMLPreview, "Sample Event") {
		t.Error("Expected HTML preview to contain event title")
	}

	if !strings.Contains(resp.HTMLPreview, "yes") {
		t.Error("Expected HTML preview to contain RSVP response")
	}
}

func TestPreviewTemplate_InvalidSyntax(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	req := &PreviewRequest{
		Type:        models.TemplateTypeInviteEmail,
		HTMLContent: "{{.Event.Title",
	}

	_, err := service.PreviewTemplate(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for invalid syntax")
	}

	validationErr, ok := err.(*models.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "html_content" {
		t.Errorf("Expected field 'html_content', got '%s'", validationErr.Field)
	}
}

func TestPreviewTemplate_UndefinedVariable(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	req := &PreviewRequest{
		Type:        models.TemplateTypeInviteEmail,
		HTMLContent: "{{.UndefinedVariable}}",
	}

	_, err := service.PreviewTemplate(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for undefined variable")
	}

	validationErr, ok := err.(*models.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "html_content" {
		t.Errorf("Expected field 'html_content', got '%s'", validationErr.Field)
	}
}

func TestPreviewTemplate_WithTextContent(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	textContent := "Event: {{.Event.Title}}\nInvite: {{.Invite.Name}}"
	req := &PreviewRequest{
		Type:        models.TemplateTypeInviteEmail,
		HTMLContent: "<h1>{{.Event.Title}}</h1>",
		TextContent: &textContent,
	}

	resp, err := service.PreviewTemplate(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.HTMLPreview == "" {
		t.Error("Expected non-empty HTML preview")
	}

	if resp.TextPreview == "" {
		t.Error("Expected non-empty text preview")
	}

	if !strings.Contains(resp.TextPreview, "Sample Event") {
		t.Error("Expected text preview to contain event title")
	}

	if !strings.Contains(resp.TextPreview, "John Doe") {
		t.Error("Expected text preview to contain invite name")
	}
}

func TestPreviewTemplate_InvalidTextContent(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	textContent := "{{.Event.Title"
	req := &PreviewRequest{
		Type:        models.TemplateTypeInviteEmail,
		HTMLContent: "<h1>{{.Event.Title}}</h1>",
		TextContent: &textContent,
	}

	_, err := service.PreviewTemplate(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for invalid text content syntax")
	}

	validationErr, ok := err.(*models.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "text_content" {
		t.Errorf("Expected field 'text_content', got '%s'", validationErr.Field)
	}
}

func TestPreviewTemplate_NilRequest(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	_, err := service.PreviewTemplate(context.Background(), nil)
	if err == nil {
		t.Fatal("Expected error for nil request")
	}
}

func TestPreviewTemplate_InvalidType(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	req := &PreviewRequest{
		Type:        models.TemplateType("invalid"),
		HTMLContent: "<h1>Test</h1>",
	}

	_, err := service.PreviewTemplate(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for invalid template type")
	}

	validationErr, ok := err.(*models.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "type" {
		t.Errorf("Expected field 'type', got '%s'", validationErr.Field)
	}
}

func TestPreviewTemplate_EmptyHTMLContent(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	req := &PreviewRequest{
		Type:        models.TemplateTypeInviteEmail,
		HTMLContent: "",
	}

	_, err := service.PreviewTemplate(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for empty HTML content")
	}
}

func TestPreviewTemplate_WithTemplateFunctions(t *testing.T) {
	validator := NewValidator(NewEngine())
	service := NewService(nil, validator)

	req := &PreviewRequest{
		Type:        models.TemplateTypeInviteEmail,
		HTMLContent: "<h1>{{.Event.Title | upper}}</h1><p>{{formatDateTime .Event.StartTime}}</p>",
	}

	resp, err := service.PreviewTemplate(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if !strings.Contains(resp.HTMLPreview, "SAMPLE EVENT") {
		t.Error("Expected HTML preview to contain uppercased event title")
	}
}
