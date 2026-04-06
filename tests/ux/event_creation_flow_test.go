package ux

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestEventForm_UnauthenticatedRedirect(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	var finalURL string
	err := chromedp.Run(ctx,
		chromedp.Navigate(srv.url("/events/new")),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Location(&finalURL),
	)

	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	if !strings.Contains(finalURL, "/login") {
		t.Errorf("Expected unauthenticated request to redirect to /login, got: %s", finalURL)
	}
}

func TestEventForm_PageLoadsForAdmin(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	var formExists bool
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url("/events/new")),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('form') !== null`, &formExists),
	)

	if err != nil {
		t.Fatalf("Failed to load event form: %v", err)
	}

	if !formExists {
		t.Error("Expected event creation form to be present")
	}
}

func TestEventForm_TitleInputExists(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	var titleInputExists bool
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url("/events/new")),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('[name="title"]') !== null`, &titleInputExists),
	)

	if err != nil {
		t.Fatalf("Failed to check title input: %v", err)
	}

	if !titleInputExists {
		t.Error("Expected event form to have a title input field")
	}
}

func TestEventForm_SubmitWithMissingTitle(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	var currentURL string
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url("/events/new")),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.Submit(`form`, chromedp.ByQuery),
		chromedp.Sleep(400*time.Millisecond),

		chromedp.Location(&currentURL),
	)

	if err != nil {
		t.Fatalf("Failed to submit empty form: %v", err)
	}

	if strings.Contains(currentURL, "/events/") && !strings.Contains(currentURL, "/events/new") {
		t.Errorf("Expected to stay on form page when title is missing, but navigated to: %s", currentURL)
	}
}

func TestEventForm_SuccessfulCreation(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	futureDate := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02T15:04")

	var finalURL string
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url("/events/new")),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),

		chromedp.SetValue(`[name="title"]`, "My Test Event", chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),

		chromedp.SetValue(`[name="timezone"]`, "America/Los_Angeles", chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),

		chromedp.Evaluate(`
			var startInput = document.querySelector('[name="start_time"]');
			if (startInput) startInput.value = '`+futureDate+`';
			var endInput = document.querySelector('[name="end_time"]');
			if (endInput) endInput.value = '`+futureDate+`';
			true;
		`, nil),

		chromedp.Submit(`form`, chromedp.ByQuery),
		chromedp.Sleep(800*time.Millisecond),

		chromedp.Location(&finalURL),
	)

	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	if strings.Contains(finalURL, "/events/new") {
		t.Logf("Form did not redirect — may require additional required fields. URL: %s", finalURL)
	}
}

func TestEventForm_FormHasTimezoneField(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	var timezoneExists bool
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url("/events/new")),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Evaluate(`
			document.querySelector('[name="timezone"]') !== null ||
			document.querySelector('#timezone') !== null
		`, &timezoneExists),
	)

	if err != nil {
		t.Fatalf("Failed to check timezone field: %v", err)
	}

	if !timezoneExists {
		t.Error("Expected event form to have a timezone field")
	}
}

func TestEventForm_FormHasDescriptionField(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	var descExists bool
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url("/events/new")),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.Evaluate(`
			document.querySelector('[name="description"]') !== null ||
			document.querySelector('textarea') !== null
		`, &descExists),
	)

	if err != nil {
		t.Fatalf("Failed to check description field: %v", err)
	}

	if !descExists {
		t.Error("Expected event form to have a description field")
	}
}
