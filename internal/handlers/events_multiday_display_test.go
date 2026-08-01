package handlers

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

func TestEventWebHandlers_EditEventForm_MultiDayDisplay(t *testing.T) {
	tests := []struct {
		name            string
		event           *models.Event
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "single day event shows only time for end",
			event: &models.Event{
				ID:        1,
				Title:     "Single Day Event",
				StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
				EndTime:   testutil.TimePtr(time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC)),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusDraft,
				CreatedBy: 1,
				Version:   1,
			},
			wantContains: []string{
				"Jun 15, 2026 at 2:00 PM",
				"- 6:00 PM",
			},
			wantNotContains: []string{
				"Jun 15, 2026 at 6:00 PM",
			},
		},
		{
			name: "multi-day event shows date for end",
			event: &models.Event{
				ID:        2,
				Title:     "Multi Day Event",
				StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
				EndTime:   testutil.TimePtr(time.Date(2026, 6, 17, 18, 0, 0, 0, time.UTC)),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusDraft,
				CreatedBy: 1,
				Version:   1,
			},
			wantContains: []string{
				"Jun 15, 2026 at 2:00 PM",
				"Jun 17, 2026 at 6:00 PM",
			},
		},
		{
			name: "event without end time",
			event: &models.Event{
				ID:        3,
				Title:     "No End Time Event",
				StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
				EndTime:   nil,
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusDraft,
				CreatedBy: 1,
				Version:   1,
			},
			wantContains: []string{
				"Jun 15, 2026 at 2:00 PM",
			},
			wantNotContains: []string{
				" - ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			mockService.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
				return tt.event, nil
			}

			tmpl := template.Must(template.New("event_form.html").Parse(`
				<!DOCTYPE html>
				<html>
				<body>
					<h1>Edit Event</h1>
					<div class="datetime-display">
						{{if and (not .Event.StartTime.IsZero) .Event.Timezone}}
							{{.Event.StartTime.Format "Jan 2, 2006 at 3:04 PM"}}{{if .Event.EndTime}}{{if ne (.Event.StartTime.Format "2006-01-02") (.Event.EndTime.Format "2006-01-02")}} - {{.Event.EndTime.Format "Jan 2, 2006 at 3:04 PM"}}{{else}} - {{.Event.EndTime.Format "3:04 PM"}}{{end}}{{end}} ({{.Event.Timezone}})
						{{end}}
					</div>
				</body>
				</html>
			`))

			handlers := NewEventWebHandlers(mockService, nil, nil, tmpl, tmpl, tmpl)

			req := httptest.NewRequest("GET", "/events/1/edit", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), &models.User{ID: 1, Role: models.RoleEventManager})
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.EditEventForm(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
			}

			body := w.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("Body should contain %q, but doesn't. Body: %s", want, body)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(body, notWant) {
					t.Errorf("Body should not contain %q, but does. Body: %s", notWant, body)
				}
			}
		})
	}
}
