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
	renderer     TemplateRenderer
	emailQueue   repositories.EmailQueueRepository
	icsGen       ics.Generator
	baseURL      string
	questionRepo repositories.QuestionRepository // optional; nil causes answers to be omitted from email
}

func NewConfirmationService(
	renderer TemplateRenderer,
	emailQueue repositories.EmailQueueRepository,
	icsGen ics.Generator,
	baseURL string,
) Service {
	return &confirmationService{
		renderer:   renderer,
		emailQueue: emailQueue,
		icsGen:     icsGen,
		baseURL:    baseURL,
	}
}

// NewConfirmationServiceWithQuestions creates a confirmationService that looks
// up question text from the repository when building confirmation emails.
// Use this in production so guests see real question labels in their confirmation.
// Answers for questions not found (e.g. deleted after RSVP) are silently omitted.
func NewConfirmationServiceWithQuestions(
	renderer TemplateRenderer,
	emailQueue repositories.EmailQueueRepository,
	icsGen ics.Generator,
	baseURL string,
	questionRepo repositories.QuestionRepository,
) Service {
	return &confirmationService{
		renderer:     renderer,
		emailQueue:   emailQueue,
		icsGen:       icsGen,
		baseURL:      baseURL,
		questionRepo: questionRepo,
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
	// Build a question ID -> text map when a repository is available.
	// One query fetches all questions for the event; answers for questions
	// not found (deleted since RSVP) are silently omitted from the email.
	questionTexts := make(map[int64]string)
	if s.questionRepo != nil && len(answers) > 0 {
		questions, err := s.questionRepo.GetByEventID(ctx, event.ID)
		if err == nil {
			for _, q := range questions {
				questionTexts[q.ID] = q.QuestionText
			}
		}
		// On error, questionTexts remains empty and all answers will be omitted —
		// a non-fatal degradation (confirmation email still sends without Q&A section).
	}

	templateData := s.prepareTemplateData(token, rsvp, invite, event, answers, questionTexts)

	htmlBody, err := s.renderer.RenderHTML(ctx, "rsvp_confirmation", templateData)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody, err := s.renderer.RenderText(ctx, "rsvp_confirmation", templateData)
	if err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	rsvpURL := fmt.Sprintf("%s/rsvp/%s", s.baseURL, token)
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
	questionTexts map[int64]string,
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

	updateURL := fmt.Sprintf("%s/rsvp/%s", s.baseURL, token)

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
		answerData := make([]RSVPAnswerData, 0, len(answers))
		for _, answer := range answers {
			questionLabel, ok := questionTexts[answer.QuestionID]
			if !ok || questionLabel == "" {
				// Question not found (deleted since RSVP) or no repo available — skip.
				continue
			}

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
			answerData = append(answerData, RSVPAnswerData{
				Question: questionLabel,
				Answer:   answerValue,
			})
		}
		if len(answerData) > 0 {
			data.Answers = answerData
		}
	}

	return data
}
