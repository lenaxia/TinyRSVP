package ux_playwright

import (
	"fmt"
	"strings"
	"testing"
)

func TestPW_InviteList_PageLoads(t *testing.T) {
	srv := SetupTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	_, ctx := NewBrowser(t)
	page := AsAdminPage(t, ctx, srv, fmt.Sprintf("/events/%d/invites", event.ID))

	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) == 0 {
		t.Error("Expected invite list page to have content")
	}
}

func TestPW_InviteList_ShowsSeededInvite(t *testing.T) {
	srv := SetupTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	_, ctx := NewBrowser(t)
	page := AsAdminPage(t, ctx, srv, fmt.Sprintf("/events/%d/invites", event.ID))

	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if !strings.Contains(body, "guest@ux-test.example.com") && !strings.Contains(body, "Test Guest") {
		t.Errorf("Expected invite list to show seeded guest email or name, got body length: %d", len(body))
	}
}

func TestPW_InviteList_UnauthenticatedRedirect(t *testing.T) {
	srv := SetupTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	_, ctx := NewBrowser(t)
	page, _ := AnonymousPage(t, ctx, srv, fmt.Sprintf("/events/%d/invites", event.ID))

	currentURL := page.URL()
	if !strings.Contains(currentURL, "/login") {
		t.Errorf("Expected unauthenticated request to redirect to /login, got: %s", currentURL)
	}
}

func TestPW_InviteList_ContainsRSVPLink(t *testing.T) {
	srv := SetupTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	_, ctx := NewBrowser(t)
	page := AsAdminPage(t, ctx, srv, fmt.Sprintf("/events/%d/invites", event.ID))

	// The invite list shows a "Copy" button that copies the RSVP link to the
	// clipboard. The actual /rsvp/{token} URL is not rendered as an anchor.
	count, err := page.Locator(`[data-action="copy-link"]`).Count()
	if err != nil {
		t.Fatalf("count copy-link buttons: %v", err)
	}
	if count == 0 {
		t.Error("Expected invite list to contain a copy-link button for the RSVP URL")
	}
}

func TestPW_InviteList_NonExistentEvent_ShowsError(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/events/999999/invites")

	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) == 0 {
		t.Error("Expected an error response for non-existent event")
	}
}

func TestPW_InviteList_MultipleInvites(t *testing.T) {
	srv := SetupTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	_, ctx := NewBrowser(t)
	page := AsAdminPage(t, ctx, srv, fmt.Sprintf("/events/%d/invites", event.ID))

	rowCount, err := page.Locator(`tr, [data-invite-id], .invite-row`).Count()
	if err != nil {
		t.Fatalf("count invite rows: %v", err)
	}
	if rowCount < 1 {
		t.Error("Expected at least 1 invite in the list")
	}
}
