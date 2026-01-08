package email

import (
	"context"
	"fmt"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/ics"
)

type RSVPAnswerData struct {
	Question string
	Answer   string
}

type RSVPConfirmationTemplateData struct {
	GuestName     string
	Response      string
	PlusOnes      int
	EventTitle    string
	EventDate     string
	EventLocation string
	UpdateURL     string
	Answers       []RSVPAnswerData
}

type confirmationService struct {
	renderer   TemplateRenderer
	emailQueue repositories.EmailQueueRepository
	icsGen     ics.Generator
}

func NewConfirmationService(
	renderer TemplateRenderer,
	emailQueue repositories.EmailQueueRepository,
	icsGen ics.Generator,
) Service {
	return &confirmationService{
		renderer:   renderer,
		emailQueue: emailQueue,
		icsGen:     icsGen,
	}
}

func (s *confirmationService) SendConfirmationEmail(
	ctx context.Context,
	token string,
	rsvp *models.RSVP,
	invite *models.Invite,
	event *models.Event,
	answers []*models.RSVPAnswer,
) error {
	templateData := s.prepareTemplateData(token, rsvp, invite, event, answers)

	htmlBody, err := s.renderer.RenderHTML(ctx, "rsvp_confirmation", templateData)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody, err := s.renderer.RenderText(ctx, "rsvp_confirmation", templateData)
	if err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	rsvpURL := fmt.Sprintf("https://example.com/rsvp/%s", token)
	icsContent, err := s.icsGen.Generate(event, rsvpURL)
	if err != nil {
		return fmt.Errorf("failed to generate ICS attachment: %w", err)
	}

	var toEmail string
	if invite.Email != nil {
		toEmail = *invite.Email
	}

	email := &models.EmailQueue{
		ToEmail:      toEmail,
		ToName:       invite.Name,
		Subject:      fmt.Sprintf("RSVP Confirmed: %s", event.Title),
		BodyText:     textBody,
		BodyHTML:     &htmlBody,
		Status:       models.EmailStatusPending,
		Attempts:     0,
		MaxAttempts:  3,
		ScheduledFor: time.Now(),
	}

	attachment := models.EmailAttachment{
		Filename:    "event.ics",
		ContentType: "text/calendar",
		Content:     icsContent,
	}

	if err := email.SetAttachments([]models.EmailAttachment{attachment}); err != nil {
		return fmt.Errorf("failed to set email attachments: %w", err)
	}

	if err := s.emailQueue.Create(ctx, email); err != nil {
		return fmt.Errorf("failed to queue confirmation email: %w", err)
	}

	return nil
}

func (s *confirmationService) prepareTemplateData(
	token string,
	rsvp *models.RSVP,
	invite *models.Invite,
	event *models.Event,
	answers []*models.RSVPAnswer,
) *RSVPConfirmationTemplateData {
	guestName := "Guest"
	if invite.Name != nil {
		guestName = *invite.Name
	}

	eventDate := event.StartTime.Format("Monday, January 2, 2006 at 3:04 PM")

	var eventLocation string
	if event.Location != nil {
		eventLocation = *event.Location
	}

	updateURL := fmt.Sprintf("https://example.com/rsvp/%s", token)

	data := &RSVPConfirmationTemplateData{
		GuestName:     guestName,
		Response:      string(rsvp.Response),
		PlusOnes:      rsvp.PlusOnes,
		EventTitle:    event.Title,
		EventDate:     eventDate,
		EventLocation: eventLocation,
		UpdateURL:     updateURL,
	}

	if len(answers) > 0 {
		answerData := make([]RSVPAnswerData, len(answers))
		for i, answer := range answers {
			answerValue := ""
			if answer.AnswerText != nil {
				answerValue = *answer.AnswerText
			} else if answer.AnswerOption != nil {
				answerValue = *answer.AnswerOption
			} else if answer.AnswerBoolean != nil {
				if *answer.AnswerBoolean {
					answerValue = "Yes"
				} else {
					answerValue = "No"
				}
			}
			answerData[i] = RSVPAnswerData{
				Question: fmt.Sprintf("Question %d", answer.QuestionID),
				Answer:   answerValue,
			}
		}
		data.Answers = answerData
	}

	return data
}
