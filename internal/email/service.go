package email

import (
	"context"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type Service interface {
	SendConfirmationEmail(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error
}

type MockService struct {
	SendConfirmationEmailFunc func(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error
	SendConfirmationEmailCalls int
}

func (m *MockService) SendConfirmationEmail(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error {
	m.SendConfirmationEmailCalls++
	if m.SendConfirmationEmailFunc != nil {
		return m.SendConfirmationEmailFunc(ctx, rsvp, invite, event, answers)
	}
	return nil
}
