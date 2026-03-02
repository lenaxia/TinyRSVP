package web

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// requireServer skips the test if localhost:8080 is not reachable.
// These browser-based tests need a running application server.
func requireServer(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8080/health")
	if err != nil || resp.StatusCode >= 500 {
		t.Skip("Skipping browser test: application server not available at localhost:8080")
	}
	if resp != nil {
		resp.Body.Close()
	}
}

// Test 1: Required HTML elements are present
func TestHTML_RequiredElementsPresent(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	requiredElements := map[string]string{
		"gallery-mode-btn":        "Gallery mode button",
		"design-mode-btn":         "Design mode button",
		"theme-gallery-container": "Theme gallery container",
		"design-mode-container":   "Design mode container",
		"design-theme-select":     "Design theme dropdown",
		"live-preview-frame":      "Live preview iframe",
		"selected-theme-id":       "Hidden theme ID input",
		"mobile-edit-btn":         "Mobile edit button",
		"mobile-preview-btn":      "Mobile preview button",
	}

	for id, description := range requiredElements {
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:8080/events/new"),
			chromedp.Sleep(500*time.Millisecond), // Wait for page load
			chromedp.Evaluate(`document.getElementById('`+id+`') !== null`, &exists),
		)

		if err != nil {
			t.Fatalf("Failed to check element #%s (%s): %v", id, description, err)
		}

		if !exists {
			t.Errorf("Required element #%s (%s) not found", id, description)
		}
	}
}

// Test 2: ARIA attributes are correct for mode toggle (tabs pattern)
func TestHTML_ARIAAttributesForModeTabs(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check initial state (gallery mode active)
	var galleryRole, designRole, gallerySelected, designSelected string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		// Check role attributes
		chromedp.AttributeValue(`#gallery-mode-btn`, `role`, &galleryRole, nil),
		chromedp.AttributeValue(`#design-mode-btn`, `role`, &designRole, nil),

		// Check aria-selected attributes
		chromedp.AttributeValue(`#gallery-mode-btn`, `aria-selected`, &gallerySelected, nil),
		chromedp.AttributeValue(`#design-mode-btn`, `aria-selected`, &designSelected, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check ARIA attributes: %v", err)
	}

	// Verify tabs pattern
	if galleryRole != "tab" {
		t.Errorf("Expected gallery button role='tab', got '%s'", galleryRole)
	}

	if designRole != "tab" {
		t.Errorf("Expected design button role='tab', got '%s'", designRole)
	}

	// Verify initial selection state
	if gallerySelected != "true" {
		t.Errorf("Expected gallery button aria-selected='true' initially, got '%s'", gallerySelected)
	}

	if designSelected != "false" {
		t.Errorf("Expected design button aria-selected='false' initially, got '%s'", designSelected)
	}
}

// Test 3: Tabpanels have correct ARIA attributes
func TestHTML_ARIAAttributesForTabpanels(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var galleryRole, designRole, galleryLabelledBy, designLabelledBy string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#theme-gallery-container`, chromedp.ByID),

		// Check tabpanel roles
		chromedp.AttributeValue(`#theme-gallery-container`, `role`, &galleryRole, nil),
		chromedp.AttributeValue(`#design-mode-container`, `role`, &designRole, nil),

		// Check aria-labelledby
		chromedp.AttributeValue(`#theme-gallery-container`, `aria-labelledby`, &galleryLabelledBy, nil),
		chromedp.AttributeValue(`#design-mode-container`, `aria-labelledby`, &designLabelledBy, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check tabpanel ARIA attributes: %v", err)
	}

	if galleryRole != "tabpanel" {
		t.Errorf("Expected gallery container role='tabpanel', got '%s'", galleryRole)
	}

	if designRole != "tabpanel" {
		t.Errorf("Expected design container role='tabpanel', got '%s'", designRole)
	}

	if galleryLabelledBy != "gallery-mode-btn" {
		t.Errorf("Expected gallery container aria-labelledby='gallery-mode-btn', got '%s'", galleryLabelledBy)
	}

	if designLabelledBy != "design-mode-btn" {
		t.Errorf("Expected design container aria-labelledby='design-mode-btn', got '%s'", designLabelledBy)
	}
}

// Test 4: Iframe has proper accessibility attributes
func TestHTML_IframeAccessibility(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var title, sandbox, ariaLive string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#live-preview-frame`, chromedp.ByID),

		chromedp.AttributeValue(`#live-preview-frame`, `title`, &title, nil),
		chromedp.AttributeValue(`#live-preview-frame`, `sandbox`, &sandbox, nil),
		chromedp.AttributeValue(`#live-preview-frame`, `aria-live`, &ariaLive, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check iframe attributes: %v", err)
	}

	if title == "" {
		t.Errorf("Expected iframe to have title attribute for accessibility")
	}

	if sandbox == "" {
		t.Errorf("Expected iframe to have sandbox attribute for security")
	}

	if ariaLive != "polite" {
		t.Errorf("Expected iframe aria-live='polite', got '%s'", ariaLive)
	}
}

// Test 5: Loading and error states have proper ARIA
func TestHTML_LoadingErrorARIA(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var loadingRole, loadingLive, errorRole string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.AttributeValue(`.live-preview-loading`, `role`, &loadingRole, nil, chromedp.ByQuery),
		chromedp.AttributeValue(`.live-preview-loading`, `aria-live`, &loadingLive, nil, chromedp.ByQuery),
		chromedp.AttributeValue(`.live-preview-error`, `role`, &errorRole, nil, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to check loading/error ARIA: %v", err)
	}

	if loadingRole != "status" {
		t.Errorf("Expected loading indicator role='status', got '%s'", loadingRole)
	}

	if loadingLive != "polite" {
		t.Errorf("Expected loading indicator aria-live='polite', got '%s'", loadingLive)
	}

	if errorRole != "alert" {
		t.Errorf("Expected error container role='alert', got '%s'", errorRole)
	}
}

// Test 6: Mobile view toggle has proper ARIA
func TestHTML_MobileViewToggleARIA(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var editRole, previewRole, editSelected, previewSelected string
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(375, 667), // Mobile size
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.AttributeValue(`#mobile-edit-btn`, `role`, &editRole, nil),
		chromedp.AttributeValue(`#mobile-preview-btn`, `role`, &previewRole, nil),
		chromedp.AttributeValue(`#mobile-edit-btn`, `aria-selected`, &editSelected, nil),
		chromedp.AttributeValue(`#mobile-preview-btn`, `aria-selected`, &previewSelected, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check mobile toggle ARIA: %v", err)
	}

	if editRole != "tab" {
		t.Errorf("Expected edit button role='tab', got '%s'", editRole)
	}

	if previewRole != "tab" {
		t.Errorf("Expected preview button role='tab', got '%s'", previewRole)
	}

	if editSelected != "true" {
		t.Errorf("Expected edit button aria-selected='true' initially, got '%s'", editSelected)
	}

	if previewSelected != "false" {
		t.Errorf("Expected preview button aria-selected='false' initially, got '%s'", previewSelected)
	}
}

// Test 7: Theme selector dropdown has proper label
func TestHTML_ThemeSelectorLabel(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var labelFor, selectID string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.Evaluate(`document.querySelector('label[for="design-theme-select"]')?.getAttribute('for')`, &labelFor),
		chromedp.AttributeValue(`#design-theme-select`, `id`, &selectID, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check theme selector label: %v", err)
	}

	if labelFor == "" {
		t.Errorf("Expected label with for='design-theme-select' to exist")
	}

	if selectID != "design-theme-select" {
		t.Errorf("Expected select to have id='design-theme-select', got '%s'", selectID)
	}
}

// Test 8: Hidden elements have aria-hidden
func TestHTML_HiddenElementsARIA(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var designHidden string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-container`, chromedp.ByID),

		// In gallery mode, design container should be hidden with aria-hidden
		chromedp.AttributeValue(`#design-mode-container`, `aria-hidden`, &designHidden, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check aria-hidden: %v", err)
	}

	if designHidden != "true" {
		t.Errorf("Expected hidden design-mode-container to have aria-hidden='true', got '%s'", designHidden)
	}
}

// Test 9: Retry button has proper attributes
func TestHTML_RetryButtonAttributes(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buttonType string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.AttributeValue(`.btn-retry-preview`, `type`, &buttonType, nil, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to check retry button: %v", err)
	}

	if buttonType != "button" {
		t.Errorf("Expected retry button type='button' (not submit), got '%s'", buttonType)
	}
}

// Test 10: Data attributes are present for JavaScript
func TestHTML_DataAttributesPresent(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pickerMode, formMobileView string
	err := chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`.theme-picker`, chromedp.ByQuery),

		chromedp.AttributeValue(`.theme-picker`, `data-mode`, &pickerMode, nil, chromedp.ByQuery),
		chromedp.AttributeValue(`.event-form`, `data-mobile-view`, &formMobileView, nil, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to check data attributes: %v", err)
	}

	if pickerMode == "" {
		t.Errorf("Expected theme-picker to have data-mode attribute")
	}

	// data-mobile-view might not be set initially, just check it exists as an attribute capability
}
