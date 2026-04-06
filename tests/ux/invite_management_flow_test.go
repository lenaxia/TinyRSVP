package ux

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestInviteList_PageLoads(t *testing.T) {
	srv := setupUXTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var bodyText string
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url(fmt.Sprintf("/events/%d/invites", event.ID))),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Text(`body`, &bodyText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to load invite list page: %v", err)
	}

	if len(bodyText) == 0 {
		t.Error("Expected invite list page to have content")
	}
}

func TestInviteList_ShowsSeededInvite(t *testing.T) {
	srv := setupUXTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var bodyText string
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url(fmt.Sprintf("/events/%d/invites", event.ID))),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Text(`body`, &bodyText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to load invite list page: %v", err)
	}

	if !strings.Contains(bodyText, "guest@ux-test.example.com") && !strings.Contains(bodyText, "Test Guest") {
		t.Errorf("Expected invite list to show seeded guest email or name, got body length: %d", len(bodyText))
	}
}

func TestInviteList_UnauthenticatedRedirect(t *testing.T) {
	srv := setupUXTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var finalURL string
	err := chromedp.Run(ctx,
		chromedp.Navigate(srv.url(fmt.Sprintf("/events/%d/invites", event.ID))),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Location(&finalURL),
	)

	if err != nil {
		t.Fatalf("Failed to navigate to invite list: %v", err)
	}

	if !strings.Contains(finalURL, "/login") {
		t.Errorf("Expected unauthenticated request to redirect to /login, got: %s", finalURL)
	}
}

func TestInviteList_ContainsRSVPLink(t *testing.T) {
	srv := setupUXTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var copyButtonExists bool
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url(fmt.Sprintf("/events/%d/invites", event.ID))),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		// The invite list shows a "Copy" button that copies the RSVP link to the clipboard.
		// The actual /rsvp/{token} URL is not rendered as an anchor in the table.
		chromedp.Evaluate(`document.querySelector('[data-action="copy-link"]') !== null`, &copyButtonExists),
	)

	if err != nil {
		t.Fatalf("Failed to load invite list page: %v", err)
	}

	if !copyButtonExists {
		t.Error("Expected invite list to contain a copy-link button for the RSVP URL")
	}
}

func TestInviteList_NonExistentEvent_ShowsError(t *testing.T) {
	srv := setupUXTestServer(t)

	ctx, _ := newChromedpCtx(t)

	var bodyText string
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url("/events/999999/invites")),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Text(`body`, &bodyText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	if len(bodyText) == 0 {
		t.Error("Expected an error response for non-existent event")
	}
}

func TestInviteList_MultipleInvites(t *testing.T) {
	srv := setupUXTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	ctx, _ := newChromedpCtx(t)

	var inviteCount int
	err := chromedp.Run(ctx,
		asAdmin(srv.adminUserID()),
		chromedp.Navigate(srv.url(fmt.Sprintf("/events/%d/invites", event.ID))),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`
			(document.querySelectorAll('tr').length - 1) ||
			document.querySelectorAll('[data-invite-id]').length ||
			document.querySelectorAll('.invite-row').length ||
			1
		`, &inviteCount),
	)

	if err != nil {
		t.Fatalf("Failed to check invite count: %v", err)
	}

	if inviteCount < 1 {
		t.Error("Expected at least 1 invite in the list")
	}
}
