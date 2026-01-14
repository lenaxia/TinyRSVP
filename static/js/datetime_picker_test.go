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

func TestDateTimePickerSingleModeNoEndDateInDisplay(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var triggerText string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#rsvp_deadline_trigger`, chromedp.ByID),
		chromedp.Click(`#rsvp_deadline_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`.calendar-day:not(.other-month):not(.disabled)`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-picker-save`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text(`#rsvp_deadline_trigger .datetime-trigger-text`, &triggerText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	if len(triggerText) == 0 {
		t.Error("Expected trigger text to be set after selecting date")
	}

	if containsEndDateSeparator(triggerText) {
		t.Errorf("Expected single-date mode trigger text to NOT contain end date separator ' - ', got: %s", triggerText)
	}
}

func TestDateTimePickerRangeModeShowsEndDateInDisplay(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var triggerText string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`.calendar-day:not(.other-month):not(.disabled)`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-toggle-btn[data-mode="end"]`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click(`.datetime-picker-content[data-content="end"] .calendar-day:not(.other-month):not(.disabled)`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-picker-content[data-content="end"] .time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-picker-save`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text(`#event_datetime_trigger .datetime-trigger-text`, &triggerText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	if len(triggerText) == 0 {
		t.Error("Expected trigger text to be set after selecting dates")
	}

	if !containsEndDateSeparator(triggerText) {
		t.Errorf("Expected range-mode trigger text to contain end date separator ' - ', got: %s", triggerText)
	}
}

func TestDateTimePickerSingleModeNoEndDateAfterRangeMode(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var eventTriggerText, rsvpTriggerText string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`.calendar-day:not(.other-month):not(.disabled)`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-toggle-btn[data-mode="end"]`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click(`.datetime-picker-content[data-content="end"] .calendar-day:not(.other-month):not(.disabled)`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-picker-content[data-content="end"] .time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-picker-save`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text(`#event_datetime_trigger .datetime-trigger-text`, &eventTriggerText, chromedp.ByQuery),
		chromedp.Click(`#rsvp_deadline_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`.calendar-day:not(.other-month):not(.disabled)`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-picker-save`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Text(`#rsvp_deadline_trigger .datetime-trigger-text`, &rsvpTriggerText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	if !containsEndDateSeparator(eventTriggerText) {
		t.Errorf("Expected event trigger (range mode) to contain end date separator ' - ', got: %s", eventTriggerText)
	}

	if containsEndDateSeparator(rsvpTriggerText) {
		t.Errorf("Expected RSVP trigger (single mode) to NOT contain end date separator ' - ' even after using range mode, got: %s", rsvpTriggerText)
	}
}

func containsEndDateSeparator(text string) bool {
	if len(text) == 0 || text[0:1] == "C" {
		return false
	}
	
	for i := 0; i < len(text)-2; i++ {
		if text[i:i+3] == " - " {
			return true
		}
	}
	return false
}
