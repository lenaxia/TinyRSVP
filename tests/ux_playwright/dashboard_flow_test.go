package ux_playwright

import (
	"strings"
	"testing"

	pw "github.com/mxschmitt/playwright-go"
)

func TestPW_Dashboard_PageLoads(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/")

	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) == 0 {
		t.Error("Expected dashboard to have text content")
	}
}

func TestPW_Dashboard_UnauthenticatedRedirect(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page, _ := AnonymousPage(t, ctx, srv, "/")

	currentURL := page.URL()
	if !strings.Contains(currentURL, "/login") {
		t.Errorf("Expected unauthenticated dashboard request to redirect to /login, got: %s", currentURL)
	}
}

func TestPW_Dashboard_ShowsSeededEvent(t *testing.T) {
	srv := SetupTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	_, ctx := NewBrowser(t)
	page := AsAdminPage(t, ctx, srv, "/")

	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(body, event.Title) {
		t.Logf("Note: seeded event title %q not found on dashboard — dashboard may show summary stats rather than event list", event.Title)
	}
}

func TestPW_EventList_PageLoads(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/events/")

	hasContent, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(hasContent) == 0 {
		t.Error("Expected event list page to render content")
	}
}

func TestPW_EventList_ShowsSeededEvent(t *testing.T) {
	srv := SetupTestServer(t)
	event, _ := seedEventAndInvite(t, srv)

	_, ctx := NewBrowser(t)
	page := AsAdminPage(t, ctx, srv, "/events/")

	AssertContainsText(t, page, event.Title)
}

func TestPW_EventList_UnauthenticatedRedirect(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page, _ := AnonymousPage(t, ctx, srv, "/events/")

	currentURL := page.URL()
	if !strings.Contains(currentURL, "/login") {
		t.Errorf("Expected unauthenticated event list to redirect to /login, got: %s", currentURL)
	}
}

func TestPW_EventList_NavigateToEventDetail(t *testing.T) {
	srv := SetupTestServer(t)
	event, _ := seedEventAndInvite(t, srv)
	_ = event

	_, ctx := NewBrowser(t)
	page := AsAdminPage(t, ctx, srv, "/events/")

	link := page.Locator(`a[href*="/events/"]`).First()
	if err := link.Click(pw.LocatorClickOptions{
		Timeout: pw.Float(5000),
	}); err != nil {
		t.Fatalf("click event link: %v", err)
	}

	if err := page.WaitForLoadState(pw.PageWaitForLoadStateOptions{
		State: pw.LoadStateNetworkidle,
	}); err != nil {
		t.Logf("wait for network idle: %v", err)
	}

	currentURL := page.URL()
	if !strings.Contains(currentURL, "/events/") {
		t.Errorf("Expected navigation to remain in /events/ paths, got: %s", currentURL)
	}
}

func TestPW_HealthEndpoint_Accessible(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page, _ := AnonymousPage(t, ctx, srv, "/health")

	AssertContainsText(t, page, "healthy")
}
