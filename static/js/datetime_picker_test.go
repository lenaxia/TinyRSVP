package js

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestDateTimePickerSingleMode(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var toggleGroupDisplay string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#rsvp_deadline_trigger`, chromedp.ByID),
		chromedp.Click(`#rsvp_deadline_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.datetime-toggle-group')).display`, &toggleGroupDisplay),
	)

	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	if toggleGroupDisplay != "none" {
		t.Errorf("Expected toggle group display to be 'none' for datetime-single mode, got: %s", toggleGroupDisplay)
	}
}

func TestDateTimePickerRangeMode(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var toggleGroupDisplay string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.datetime-toggle-group')).display`, &toggleGroupDisplay),
	)

	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	if toggleGroupDisplay == "none" {
		t.Errorf("Expected toggle group to be visible for datetime-range mode, got display: %s", toggleGroupDisplay)
	}
}

func TestDateTimePickerToggleGroupHiddenOnOpen(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var toggleGroupDisplay string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`.datetime-picker-close`, chromedp.ByQuery),
		chromedp.WaitNotVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Click(`#rsvp_deadline_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.datetime-toggle-group')).display`, &toggleGroupDisplay),
	)

	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	if toggleGroupDisplay != "none" {
		t.Errorf("Expected toggle group display to be 'none' after switching from range to single mode, got: %s", toggleGroupDisplay)
	}
}
