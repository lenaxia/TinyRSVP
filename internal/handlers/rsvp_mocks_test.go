package handlers

// This file contains shared manual mock types used by test files that use
// the func-field pattern and cannot be easily migrated to gomock
// (e.g., color_override_integration_test.go, rsvp_questions_test.go,
// rsvp_theme_test.go, rsvp_unsubscribe_test.go, rsvp_csrf_integration_test.go,
// event_customization_e2e_test.go, rsvp_color_override_test.go).

import (
	"context"
	"errors"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/rsvp"
)

type mockRSVPInviteService struct {
	getInviteByTokenFunc func(ctx context.Context, token string) (*models.Invite, error)
	markViewedFunc       func(ctx context.Context, inviteID int64) error
	unsubscribeFunc      func(ctx context.Context, token string) error
}

func (m *mockRSVPInviteService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	if m.getInviteByTokenFunc != nil {
		return m.getInviteByTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRSVPInviteService) MarkInviteViewed(ctx context.Context, inviteID int64) error {
	if m.markViewedFunc != nil {
		return m.markViewedFunc(ctx, inviteID)
	}
	return nil
}

func (m *mockRSVPInviteService) UnsubscribeFromReminders(ctx context.Context, token string) error {
	if m.unsubscribeFunc != nil {
		return m.unsubscribeFunc(ctx, token)
	}
	return nil
}

type mockRSVPEventRepository struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockRSVPEventRepository) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRSVPEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRSVPEventRepository) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRSVPEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockRSVPEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockRSVPEventRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRSVPEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) ListWithStats(ctx context.Context, filters repositories.ListFilters) ([]*models.EventWithStats, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) CountEvents(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockRSVPEventRepository) GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
	return nil
}

func (m *mockRSVPEventRepository) DeleteComponentOverrides(ctx context.Context, eventID int64) error {
	return nil
}

type mockRSVPRSVPRepository struct {
	getByInviteIDFunc func(ctx context.Context, inviteID int64) (*models.RSVP, error)
}

func (m *mockRSVPRSVPRepository) Create(ctx context.Context, rsvpEntry *models.RSVP) error {
	return nil
}

func (m *mockRSVPRSVPRepository) GetByID(ctx context.Context, id int64) (*models.RSVP, error) {
	return nil, nil
}

func (m *mockRSVPRSVPRepository) GetByInviteID(ctx context.Context, inviteID int64) (*models.RSVP, error) {
	if m.getByInviteIDFunc != nil {
		return m.getByInviteIDFunc(ctx, inviteID)
	}
	return nil, &models.NotFoundError{Resource: "rsvp"}
}

func (m *mockRSVPRSVPRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.RSVP, error) {
	return nil, nil
}

func (m *mockRSVPRSVPRepository) Update(ctx context.Context, rsvpEntry *models.RSVP) error {
	return nil
}

func (m *mockRSVPRSVPRepository) GetStats(ctx context.Context, eventID int64) (*repositories.RSVPStats, error) {
	return nil, nil
}

func (m *mockRSVPRSVPRepository) GetByInviteIDs(ctx context.Context, inviteIDs []int64) ([]*models.RSVP, error) {
	return []*models.RSVP{}, nil
}

type mockRSVPQuestionRepository struct {
	getByEventIDFunc func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
}

func (m *mockRSVPQuestionRepository) Create(ctx context.Context, question *models.PreferenceQuestion) error {
	return nil
}

func (m *mockRSVPQuestionRepository) GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
	return nil, nil
}

func (m *mockRSVPQuestionRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	if m.getByEventIDFunc != nil {
		return m.getByEventIDFunc(ctx, eventID)
	}
	return []*models.PreferenceQuestion{}, nil
}

func (m *mockRSVPQuestionRepository) Update(ctx context.Context, question *models.PreferenceQuestion) error {
	return nil
}

func (m *mockRSVPQuestionRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRSVPQuestionRepository) UpdateDisplayOrder(ctx context.Context, eventID int64, questionIDs []int64) error {
	return nil
}

func (m *mockRSVPQuestionRepository) Reorder(ctx context.Context, eventID int64, questionIDs []int64) error {
	return nil
}

type mockAnswerRepository struct {
	getByRSVPIDFunc func(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error)
}

func (m *mockAnswerRepository) Create(ctx context.Context, answer *models.RSVPAnswer) error {
	return nil
}

func (m *mockAnswerRepository) GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
	if m.getByRSVPIDFunc != nil {
		return m.getByRSVPIDFunc(ctx, rsvpID)
	}
	return []*models.RSVPAnswer{}, nil
}

func (m *mockAnswerRepository) GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error) {
	return nil, nil
}

func (m *mockAnswerRepository) Update(ctx context.Context, answer *models.RSVPAnswer) error {
	return nil
}

func (m *mockAnswerRepository) DeleteByRSVPID(ctx context.Context, rsvpID int64) error {
	return nil
}

type mockRSVPService struct {
	submitRSVPFunc func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error)
	updateRSVPFunc func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error)
}

func (m *mockRSVPService) SubmitRSVP(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
	if m.submitRSVPFunc != nil {
		return m.submitRSVPFunc(ctx, token, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRSVPService) UpdateRSVP(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
	if m.updateRSVPFunc != nil {
		return m.updateRSVPFunc(ctx, token, req)
	}
	return nil, errors.New("not implemented")
}

func (m * mockRSVPEventRepository) GetDashboardStatsByCreator(ctx context.Context, creatorID int64) (*models.DashboardStats, error) {
	return &models.DashboardStats{}, nil
}

func (m * mockRSVPEventRepository) CountEventsByCreator(ctx context.Context, creatorID int64) (int, error) {
	return 0, nil
}
