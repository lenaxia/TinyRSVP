package js

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestDateTimePickerSingleMode(t *testing.T) {
	requireServer(t)
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

func TestDateTimePickerRangeModeEndDisabledByDefault(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var toggleGroupDisplay, endToggleDisplay string
	var endCheckboxChecked bool
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.datetime-toggle-group')).display`, &toggleGroupDisplay),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.datetime-end-toggle')).display`, &endToggleDisplay),
		chromedp.Evaluate(`document.querySelector('.datetime-end-checkbox').checked`, &endCheckboxChecked),
	)

	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	if toggleGroupDisplay != "none" {
		t.Errorf("Expected toggle group hidden by default (end disabled), got display: %s", toggleGroupDisplay)
	}

	if endToggleDisplay == "none" {
		t.Errorf("Expected 'Add end time' checkbox visible for datetime-range mode, got display: %s", endToggleDisplay)
	}

	if endCheckboxChecked {
		t.Error("Expected 'Add end time' checkbox unchecked by default")
	}
}

func TestDateTimePickerEnablingEndShowsToggleGroup(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var toggleGroupDisplay string
	var activeContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Enable end time via the checkbox
		chromedp.Evaluate(`document.querySelector('.datetime-end-checkbox').click()`, nil),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.datetime-toggle-group')).display`, &toggleGroupDisplay),
		chromedp.Evaluate(`document.querySelector('.datetime-picker-content.active')?.dataset?.content || 'none'`, &activeContent),
	)

	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	if toggleGroupDisplay == "none" {
		t.Errorf("Expected toggle group visible after enabling end time, got display: %s", toggleGroupDisplay)
	}

	if activeContent != "end" {
		t.Errorf("Expected end pane active after enabling end time, got: %s", activeContent)
	}
}

func TestDateTimePickerToggleGroupHiddenOnOpen(t *testing.T) {
	requireServer(t)
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
		// Wait for panel to close (CSS transition ~300ms + buffer)
		chromedp.Sleep(600*time.Millisecond),
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
	requireServer(t)
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
	requireServer(t)
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
		chromedp.Evaluate(`document.querySelector('.datetime-end-checkbox').click()`, nil),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Click(`.datetime-toggle-btn[data-mode="end"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Use JS click to bypass chromedp viewport visibility checks for end pane elements
		chromedp.Evaluate(`(function(){var d=document.querySelector('.datetime-picker-content[data-content="end"] .calendar-day:not(.other-month):not(.disabled)');if(d){d.click();return true;}return false;})()`, nil),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`(function(){var t=document.querySelector('.datetime-picker-content[data-content="end"] .time-option');if(t){t.click();return true;}return false;})()`, nil),
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
	requireServer(t)
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
		chromedp.Evaluate(`document.querySelector('.datetime-end-checkbox').click()`, nil),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Click(`.datetime-toggle-btn[data-mode="end"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Use JS click to bypass chromedp viewport visibility checks for end pane elements
		chromedp.Evaluate(`(function(){var d=document.querySelector('.datetime-picker-content[data-content="end"] .calendar-day:not(.other-month):not(.disabled)');if(d){d.click();return true;}return false;})()`, nil),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`(function(){var t=document.querySelector('.datetime-picker-content[data-content="end"] .time-option');if(t){t.click();return true;}return false;})()`, nil),
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

// TestDateTimePickerDeadlineSavesToDeadlineInput is a regression test for the
// bug where saving the RSVP deadline picker would write to start_time instead
// of rsvp_deadline_input (due to a hardcoded inputId check in saveDateTime).
func TestDateTimePickerDeadlineSavesToDeadlineInput(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var deadlineValue, startValue string
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
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Value(`#rsvp_deadline_input`, &deadlineValue, chromedp.ByID),
		chromedp.Value(`#start_time`, &startValue, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}
	if deadlineValue == "" {
		t.Error("rsvp_deadline_input should have a value after saving the deadline picker")
	}
	if startValue != "" {
		t.Errorf("start_time should NOT have been written when saving the deadline picker, got: %s", startValue)
	}
}

// TestDateTimePickerEventSaveDoesNotTouchDeadlineInput is a regression test
// verifying that saving the event datetime picker does not overwrite
// rsvp_deadline_input (double-fire from multiple listener registrations).
func TestDateTimePickerEventSaveDoesNotTouchDeadlineInput(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var startValue, deadlineValue string
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
		chromedp.Click(`.datetime-picker-save`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Value(`#start_time`, &startValue, chromedp.ByID),
		chromedp.Value(`#rsvp_deadline_input`, &deadlineValue, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}
	if startValue == "" {
		t.Error("start_time should have a value after saving the event datetime picker")
	}
	if deadlineValue != "" {
		t.Errorf("rsvp_deadline_input should NOT have been written when saving the event picker, got: %s", deadlineValue)
	}
}

// TestDateTimePickerDeadlineAfterEventPickerUsedEndMode is a regression test for
// the bug where: user sets event start+end time (leaving panel in 'end' mode),
// then opens the RSVP deadline picker. Because openPanel previously only called
// switchMode when showEndTime=true, the panel stayed in 'end' mode. The user
// saw the stale [data-content="end"] panel and their calendar/time clicks went
// to the event instance, writing the deadline date into end_time instead of
// rsvp_deadline_input.
//
// This test verifies the fix: openPanel now always calls switchMode(currentMode),
// so [data-content="start"] is always activated when the single-mode deadline
// picker opens, regardless of what mode a prior picker left the panel in.
func TestDateTimePickerDeadlineAfterEventPickerUsedEndMode(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Verify that after opening the RSVP deadline picker, the [data-content="start"]
	// pane is active (not "end"), regardless of prior picker state. We set the
	// panel to end-mode via JS to simulate the prior picker leaving it in that state.
	var activeContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Directly set the panel to end-mode state (simulating prior event picker use)
		chromedp.Evaluate(`
			document.querySelectorAll('.datetime-picker-content').forEach(c => {
				c.classList.toggle('active', c.dataset.content === 'end');
			});
		`, nil),
		chromedp.Sleep(100*time.Millisecond),

		// Open the RSVP deadline picker — fix must activate [data-content="start"]
		chromedp.Click(`#rsvp_deadline_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),

		// Check which content pane is active
		chromedp.Evaluate(`
			document.querySelector('.datetime-picker-content.active')?.dataset?.content || 'none'
		`, &activeContent),
	)
	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}
	if activeContent != "start" {
		t.Errorf("expected [data-content='start'] to be active after opening deadline picker, got %q", activeContent)
	}
}

// TestDateTimePickerUncheckingEndClearsEndTime verifies that enabling end time,
// saving (which writes end_time), then unchecking the "Add end time" checkbox
// and saving again clears the end_time input. This satisfies the requirement
// that an end datetime can be removed after being set.
func TestDateTimePickerUncheckingEndClearsEndTime(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var endValueAfterSet, endValueAfterClear string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(400*time.Millisecond),

		// Pick start date + time
		chromedp.Click(`.calendar-day:not(.other-month):not(.disabled)`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// Enable end and pick an end date (second available day) + time
		chromedp.Evaluate(`document.querySelector('.datetime-end-checkbox').click()`, nil),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Evaluate(`(function(){var d=document.querySelectorAll('.datetime-picker-content[data-content="end"] .calendar-day:not(.other-month):not(.disabled)');if(d.length>1){d[1].click();return true;}return false;})()`, nil),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`(function(){var t=document.querySelector('.datetime-picker-content[data-content="end"] .time-option:not(.disabled)');if(t){t.click();return true;}return false;})()`, nil),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Click(`.datetime-picker-save`, chromedp.ByQuery),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Value(`#end_time`, &endValueAfterSet, chromedp.ByID),

		// Reopen, uncheck end, save — end_time must be cleared
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Evaluate(`document.querySelector('.datetime-end-checkbox').click()`, nil),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click(`.datetime-picker-save`, chromedp.ByQuery),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Value(`#end_time`, &endValueAfterClear, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}
	if endValueAfterSet == "" {
		t.Error("end_time should be populated after enabling end and saving")
	}
	if endValueAfterClear != "" {
		t.Errorf("end_time should be cleared after unchecking end and saving, got: %s", endValueAfterClear)
	}
}

// TestDateTimePickerEndCalendarDisablesBeforeStart verifies that once a start
// date is chosen and end time enabled, days strictly before the start date are
// rendered as disabled in the end pane (cannot select end before start).
func TestDateTimePickerEndCalendarDisablesBeforeStart(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// JS that picks a start day, enables end, then reports whether any
	// non-other-month day earlier than the selected start day is clickable
	// (i.e. lacks the 'disabled' class). The picker is correct when the
	// returned count of such clickably-invalid days is 0.
	var invalidClickableDays int64
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(400*time.Millisecond),

		// Pick a start day that is NOT the first of the month, so earlier
		// same-month days exist to test against.
		chromedp.Evaluate(`(function(){
			var days = document.querySelectorAll('.datetime-picker-content[data-content="start"] .calendar-day:not(.other-month):not(.disabled)');
			for (var i = days.length - 1; i >= 0; i--) {
				var n = parseInt(days[i].textContent, 10);
				if (n > 2) { days[i].click(); return true; }
			}
			return false;
		})()`, nil),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click(`.time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.Evaluate(`document.querySelector('.datetime-end-checkbox').click()`, nil),
		chromedp.Sleep(400*time.Millisecond),

		chromedp.Evaluate(`(function(){
			var endDays = document.querySelectorAll('.datetime-picker-content[data-content="end"] .calendar-day:not(.other-month)');
			var startText = (document.querySelector('.datetime-picker-content[data-content="start"] .calendar-day.selected')||{}).textContent;
			var startNum = parseInt(startText, 10);
			var bad = 0;
			endDays.forEach(function(d){
				var n = parseInt(d.textContent, 10);
				if (!isNaN(startNum) && !isNaN(n) && n < startNum && !d.classList.contains('disabled')) { bad++; }
			});
			return bad;
		})()`, &invalidClickableDays),
	)
	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}
	if invalidClickableDays != 0 {
		t.Errorf("Expected no clickable end days before start date, got %d", invalidClickableDays)
	}
}

// TestDateTimePickerMovingStartAfterEndClearsEnd verifies that when an end time
// is set, moving the start date to a later day than the end (which would make
// end < start) clears the end selection on save, rather than persisting an
// invalid range that errors on submit.
func TestDateTimePickerMovingStartAfterEndClearsEnd(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var endValueAfterMove string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/static/datetime_picker_test.html"),
		chromedp.WaitVisible(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.Click(`#event_datetime_trigger`, chromedp.ByID),
		chromedp.WaitVisible(`.datetime-picker-panel.open`, chromedp.ByQuery),
		chromedp.Sleep(400*time.Millisecond),

		// Pick an early start day
		chromedp.Evaluate(`(function(){var d=document.querySelector('.datetime-picker-content[data-content="start"] .calendar-day:not(.other-month):not(.disabled)');if(d){d.click();return true;}return false;})()`, nil),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Click(`.time-option`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// Enable end, pick the very next day as end
		chromedp.Evaluate(`document.querySelector('.datetime-end-checkbox').click()`, nil),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Evaluate(`(function(){var d=document.querySelectorAll('.datetime-picker-content[data-content="end"] .calendar-day:not(.other-month):not(.disabled)');if(d.length>1){d[1].click();return true;}return false;})()`, nil),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`(function(){var t=document.querySelector('.datetime-picker-content[data-content="end"] .time-option:not(.disabled)');if(t){t.click();return true;}return false;})()`, nil),
		chromedp.Sleep(200*time.Millisecond),

		// Switch back to start and pick a day far in the future (last available)
		chromedp.Click(`.datetime-toggle-btn[data-mode="start"]`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`(function(){var d=document.querySelectorAll('.datetime-picker-content[data-content="start"] .calendar-day:not(.other-month):not(.disabled)');if(d.length){d[d.length-1].click();return true;}return false;})()`, nil),
		chromedp.Sleep(300*time.Millisecond),

		chromedp.Click(`.datetime-picker-save`, chromedp.ByQuery),
		chromedp.Sleep(400*time.Millisecond),
		chromedp.Value(`#end_time`, &endValueAfterMove, chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}
	if endValueAfterMove != "" {
		t.Errorf("end_time should be cleared after moving start past end, got: %s", endValueAfterMove)
	}
}
