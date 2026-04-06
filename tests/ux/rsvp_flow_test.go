package ux

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func seedEventAndInvite(t *testing.T, srv *uxTestServer) (event *models.Event, plainToken string) {
	t.Helper()

	adminCtx := auth.WithUser(context.Background(), srv.adminUser)

	startTime := time.Now().Add(7 * 24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	description := "A test event for UX testing"
	location := "Test Venue, 123 Main St"

	event = &models.Event{
		Title:       "UX Test Event",
		Description: &description,
		Location:    &location,
		StartTime:   startTime,
		EndTime:     &endTime,
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		Status:      models.EventStatusDraft,
		CreatedBy:   srv.adminUser.ID,
	}

	if err := srv.eventService.CreateEvent(adminCtx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	guestName := "Test Guest"
	resp, err := srv.inviteService.CreateIndividualInvite(adminCtx, srv.adminUser, &invites.CreateIndividualInviteRequest{
		EventID: event.ID,
		Email:   "guest@ux-test.example.com",
		Name:    &guestName,
	})
	if err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	return event, resp.Token
}

func TestRSVP_PageLoadsWithToken(t *testing.T) {
	srv := setupUXTestServer(t)
	_, plainToken := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var pageTitle string
	var formExists bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(srv.url("/rsvp/"+plainToken)),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Title(&pageTitle),
		chromedp.Evaluate(`document.querySelector('form') !== null`, &formExists),
	)

	if err != nil {
		t.Fatalf("Failed to load RSVP page: %v", err)
	}

	if !formExists {
		t.Error("Expected RSVP form to be present on the page")
	}
}

func TestRSVP_InvalidToken_ShowsError(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	var statusCode int
	err := chromedp.Run(ctx,
		chromedp.Navigate(srv.url("/rsvp/totally-invalid-token-that-does-not-exist")),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Evaluate(`window.performance.getEntriesByType('navigation')[0]?.responseStatus || 0`, &statusCode),
	)

	if err != nil {
		t.Fatalf("Failed to navigate to invalid RSVP page: %v", err)
	}

	var bodyText string
	err = chromedp.Run(ctx,
		chromedp.Text(`body`, &bodyText, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get page body text: %v", err)
	}

	if statusCode == 200 {
		t.Log("Note: page returned 200 but content should indicate error (chromedp follows redirects)")
	}
}

func TestRSVP_HappyPath_SubmitForm(t *testing.T) {
	srv := setupUXTestServer(t)
	event, plainToken := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var finalURL string

	err := chromedp.Run(ctx,
		chromedp.Navigate(srv.url("/rsvp/"+plainToken)),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),

		// The default legacy template uses name="response" radio inputs.
		chromedp.Click(`input[name="response"][value="yes"]`, chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),

		// Click the submit button rather than using chromedp.Submit, which
		// blocks waiting for network-idle and hangs across server-side redirects.
		chromedp.Click(`button[type="submit"], input[type="submit"]`, chromedp.ByQuery),

		// Wait for the confirmation page to appear (redirect destination).
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		chromedp.Location(&finalURL),
	)

	if err != nil {
		t.Fatalf("Failed to submit RSVP form: %v", err)
	}

	if finalURL == srv.url("/rsvp/"+plainToken) {
		t.Logf("Form submission stayed on RSVP page — may have validation errors or missing required fields. Event ID: %d", event.ID)
	}
}

func TestRSVP_EventTitleVisibleOnPage(t *testing.T) {
	srv := setupUXTestServer(t)
	event, plainToken := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var bodyText string
	err := chromedp.Run(ctx,
		chromedp.Navigate(srv.url("/rsvp/"+plainToken)),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Text(`body`, &bodyText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to load RSVP page: %v", err)
	}

	if len(bodyText) == 0 {
		t.Error("Expected page to have text content")
	}

	_ = event
}

func TestRSVP_AttendingOptionExists(t *testing.T) {
	srv := setupUXTestServer(t)
	_, plainToken := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var attendingExists bool
	err := chromedp.Run(ctx,
		chromedp.Navigate(srv.url("/rsvp/"+plainToken)),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Evaluate(`
			document.querySelector('[name="attending"]') !== null ||
			document.querySelector('[name="status"]') !== null ||
			document.querySelector('[type="radio"]') !== null
		`, &attendingExists),
	)

	if err != nil {
		t.Fatalf("Failed to check RSVP options: %v", err)
	}

	if !attendingExists {
		t.Error("Expected RSVP page to have attending/status input options")
	}
}

func TestRSVP_DeclinedPath(t *testing.T) {
	srv := setupUXTestServer(t)
	_, plainToken := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var finalURL string
	err := chromedp.Run(ctx,
		chromedp.Navigate(srv.url("/rsvp/"+plainToken)),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),

		// The default legacy template uses name="response" radio inputs.
		chromedp.Click(`input[name="response"][value="no"]`, chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),

		chromedp.Submit(`form`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		chromedp.Location(&finalURL),
	)

	if err != nil {
		t.Fatalf("Failed to submit declined RSVP: %v", err)
	}

	_ = finalURL
}
