package email

import (
	"context"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type Service interface {
	SendConfirmationEmail(ctx context.Context, token string, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error
}
