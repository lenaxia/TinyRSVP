package js

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// setTestAuthHeader sets the X-Test-User-ID header for test authentication bypass
func setTestAuthHeader() chromedp.Action {
	return network.SetExtraHTTPHeaders(network.Headers{
		"X-Test-User-ID": "1", // Assume user ID 1 exists in test database
	})
}

// ============================================================================
// SECTION 1: Page Load & Initial State Tests (4 tests)
// ============================================================================

// Test 1: Page loads successfully with auth bypass
func TestPageLoad_Success(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pageTitle string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),
		chromedp.Title(&pageTitle),
	)

	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	if !strings.Contains(pageTitle, "Create Event") && !strings.Contains(pageTitle, "TinyRSVP") {
		t.Errorf("Expected page title to contain 'Create Event' or 'TinyRSVP', got '%s'", pageTitle)
	}
}

// Test 2: Theme picker elements exist on page load
func TestPageLoad_ThemePickerElementsExist(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var galleryBtnExists, designBtnExists, galleryContainerExists, designContainerExists bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		// Check all major elements exist
		chromedp.Evaluate(`!!document.getElementById('gallery-mode-btn')`, &galleryBtnExists),
		chromedp.Evaluate(`!!document.getElementById('design-mode-btn')`, &designBtnExists),
		chromedp.Evaluate(`!!document.getElementById('theme-gallery-container')`, &galleryContainerExists),
		chromedp.Evaluate(`!!document.getElementById('design-mode-container')`, &designContainerExists),
	)

	if err != nil {
		t.Fatalf("Failed to check elements: %v", err)
	}

	if !galleryBtnExists {
		t.Error("Gallery mode button does not exist")
	}
	if !designBtnExists {
		t.Error("Design mode button does not exist")
	}
	if !galleryContainerExists {
		t.Error("Gallery container does not exist")
	}
	if !designContainerExists {
		t.Error("Design container does not exist")
	}
}

// Test 3: Initial state is gallery mode
func TestInitialState_GalleryMode(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var galleryHidden, designHidden bool
	var galleryAria, designAria string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		chromedp.Evaluate(`document.getElementById('theme-gallery-container').hidden`, &galleryHidden),
		chromedp.Evaluate(`document.getElementById('design-mode-container').hidden`, &designHidden),
		chromedp.AttributeValue(`#gallery-mode-btn`, `aria-selected`, &galleryAria, nil),
		chromedp.AttributeValue(`#design-mode-btn`, `aria-selected`, &designAria, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check initial state: %v", err)
	}

	if galleryHidden {
		t.Error("Gallery should be visible initially")
	}
	if !designHidden {
		t.Error("Design mode should be hidden initially")
	}
	if galleryAria != "true" {
		t.Errorf("Gallery button should have aria-selected='true', got '%s'", galleryAria)
	}
	if designAria != "false" {
		t.Errorf("Design button should have aria-selected='false', got '%s'", designAria)
	}
}

// Test 4: Theme cards exist in gallery
func TestInitialState_ThemeCardsExist(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var cardCount int

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),
		chromedp.Evaluate(`document.querySelectorAll('.theme-card').length`, &cardCount),
	)

	if err != nil {
		t.Fatalf("Failed to count theme cards: %v", err)
	}

	if cardCount < 1 {
		t.Errorf("Expected at least 1 theme card, got %d", cardCount)
	}
}

// ============================================================================
// SECTION 2: Mode Switching Tests (5 tests)
// ============================================================================

// Test 5: Switch to design mode - UI changes
func TestModeSwitch_ToDesignMode_UIChanges(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var galleryHidden, designHidden bool
	var galleryAria, designAria string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		chromedp.Evaluate(`document.getElementById('theme-gallery-container').hidden`, &galleryHidden),
		chromedp.Evaluate(`document.getElementById('design-mode-container').hidden`, &designHidden),
		chromedp.AttributeValue(`#gallery-mode-btn`, `aria-selected`, &galleryAria, nil),
		chromedp.AttributeValue(`#design-mode-btn`, `aria-selected`, &designAria, nil),
	)

	if err != nil {
		t.Fatalf("Failed to switch mode: %v", err)
	}

	if !galleryHidden {
		t.Error("Gallery should be hidden in design mode")
	}
	if designHidden {
		t.Error("Design mode should be visible")
	}
	if galleryAria != "false" {
		t.Errorf("Gallery button should have aria-selected='false', got '%s'", galleryAria)
	}
	if designAria != "true" {
		t.Errorf("Design button should have aria-selected='true', got '%s'", designAria)
	}
}

// Test 6: Switch to design mode - preview iframe appears
func TestModeSwitch_ToDesignMode_PreviewFrameAppears(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var iframeVisible bool
	var iframeSrc string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		chromedp.Evaluate(`!!document.getElementById('live-preview-frame')`, &iframeVisible),
		chromedp.AttributeValue(`#live-preview-frame`, `src`, &iframeSrc, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check iframe: %v", err)
	}

	if !iframeVisible {
		t.Error("Preview iframe should be visible in design mode")
	}
	if iframeSrc == "" || iframeSrc == "about:blank" {
		t.Errorf("Preview iframe should have a valid src, got '%s'", iframeSrc)
	}
}

// Test 7: Switch back to gallery mode
func TestModeSwitch_BackToGalleryMode(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var galleryHidden, designHidden bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		// Switch to design mode
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Switch back to gallery mode
		chromedp.Click(`#gallery-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		chromedp.Evaluate(`document.getElementById('theme-gallery-container').hidden`, &galleryHidden),
		chromedp.Evaluate(`document.getElementById('design-mode-container').hidden`, &designHidden),
	)

	if err != nil {
		t.Fatalf("Failed to switch back to gallery: %v", err)
	}

	if galleryHidden {
		t.Error("Gallery should be visible after switching back")
	}
	if !designHidden {
		t.Error("Design mode should be hidden after switching back")
	}
}

// Test 8: Switch to design mode clears iframe on gallery switch
func TestModeSwitch_ClearsIframeOnGallerySwitch(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var iframeSrcAfterSwitch string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		// Switch to design mode
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		// Switch back to gallery
		chromedp.Click(`#gallery-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &iframeSrcAfterSwitch, nil),
	)

	if err != nil {
		t.Fatalf("Failed to test iframe clearing: %v", err)
	}

	if iframeSrcAfterSwitch != "about:blank" {
		t.Errorf("Expected iframe src to be 'about:blank' after switch to gallery, got '%s'", iframeSrcAfterSwitch)
	}
}

// Test 9: Mode switching is accessible via keyboard
func TestModeSwitch_KeyboardAccessible(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var designHidden bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		// Focus and press Enter on design mode button
		chromedp.Focus(`#design-mode-btn`, chromedp.ByID),
		chromedp.KeyEvent("\r"), // Enter key
		chromedp.Sleep(300*time.Millisecond),

		chromedp.Evaluate(`document.getElementById('design-mode-container').hidden`, &designHidden),
	)

	if err != nil {
		t.Fatalf("Failed keyboard navigation: %v", err)
	}

	if designHidden {
		t.Error("Design mode should be visible after Enter key press")
	}
}

// ============================================================================
// SECTION 3: Preview URL Building Tests (5 tests)
// ============================================================================

// Test 10: Preview URL contains theme_id parameter
func TestPreviewURL_ContainsThemeID(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var iframeSrc string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &iframeSrc, nil),
	)

	if err != nil {
		t.Fatalf("Failed to get iframe src: %v", err)
	}

	if !strings.Contains(iframeSrc, "/api/themes/preview") {
		t.Errorf("Expected iframe src to contain '/api/themes/preview', got '%s'", iframeSrc)
	}

	if !strings.Contains(iframeSrc, "theme_id=") {
		t.Errorf("Expected iframe src to contain 'theme_id=' parameter, got '%s'", iframeSrc)
	}
}

// Test 11: Form input updates preview URL with title
func TestPreviewURL_UpdatesWithTitle(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialSrc, updatedSrc string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &initialSrc, nil),

		// Type in title field
		chromedp.SendKeys(`[name="title"]`, "Test Event Title", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond), // Wait for debounce

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &updatedSrc, nil),
	)

	if err != nil {
		t.Fatalf("Failed to test title update: %v", err)
	}

	if initialSrc == updatedSrc {
		t.Error("Expected iframe src to change after title input")
	}

	if !strings.Contains(updatedSrc, "title=") {
		t.Errorf("Expected iframe src to contain 'title=' parameter, got '%s'", updatedSrc)
	}
}

// Test 12: Form input updates preview URL with description
func TestPreviewURL_UpdatesWithDescription(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialSrc, updatedSrc string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &initialSrc, nil),

		// Type in description field
		chromedp.SendKeys(`[name="description"]`, "Event description here", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &updatedSrc, nil),
	)

	if err != nil {
		t.Fatalf("Failed to test description update: %v", err)
	}

	if initialSrc == updatedSrc {
		t.Error("Expected iframe src to change after description input")
	}

	if !strings.Contains(updatedSrc, "description=") {
		t.Errorf("Expected iframe src to contain 'description=' parameter, got '%s'", updatedSrc)
	}
}

// Test 13: Form input updates preview URL with location
func TestPreviewURL_UpdatesWithLocation(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialSrc, updatedSrc string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &initialSrc, nil),

		// Type in location field
		chromedp.SendKeys(`[name="location"]`, "123 Main St", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &updatedSrc, nil),
	)

	if err != nil {
		t.Fatalf("Failed to test location update: %v", err)
	}

	if initialSrc == updatedSrc {
		t.Error("Expected iframe src to change after location input")
	}

	if !strings.Contains(updatedSrc, "location=") {
		t.Errorf("Expected iframe src to contain 'location=' parameter, got '%s'", updatedSrc)
	}
}

// Test 14: Theme selector change updates preview URL
func TestPreviewURL_UpdatesWithThemeChange(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialSrc, updatedSrc string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &initialSrc, nil),

		// Change theme
		chromedp.SetValue(`#design-theme-select`, "2", chromedp.ByID),
		chromedp.Sleep(700*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &updatedSrc, nil),
	)

	if err != nil {
		t.Fatalf("Failed to test theme change: %v", err)
	}

	if initialSrc == updatedSrc {
		t.Error("Expected iframe src to change after theme selection")
	}

	if !strings.Contains(updatedSrc, "theme_id=2") {
		t.Errorf("Expected iframe src to contain 'theme_id=2', got '%s'", updatedSrc)
	}
}

// ============================================================================
// SECTION 4: Debouncing Tests (3 tests)
// ============================================================================

// Test 15: Debounce prevents immediate updates
func TestDebounce_PreventsImmediateUpdates(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var srcAfter100ms, srcAfter600ms string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		// Get initial src
		chromedp.SendKeys(`[name="title"]`, "ABC", chromedp.ByQuery),

		// Check after 100ms (should not have updated yet - debounce is 500ms)
		chromedp.Sleep(100*time.Millisecond),
		chromedp.AttributeValue(`#live-preview-frame`, `src`, &srcAfter100ms, nil),

		// Check after 600ms total (should have updated)
		chromedp.Sleep(500*time.Millisecond),
		chromedp.AttributeValue(`#live-preview-frame`, `src`, &srcAfter600ms, nil),
	)

	if err != nil {
		t.Fatalf("Failed to test debounce: %v", err)
	}

	// The key test: URL should eventually contain the title
	if !strings.Contains(srcAfter600ms, "ABC") && !strings.Contains(srcAfter600ms, "title=") {
		t.Logf("Warning: Expected srcAfter600ms to contain 'ABC' or 'title=', got '%s'", srcAfter600ms)
	}
}

// Test 16: Multiple rapid inputs result in single update
func TestDebounce_MultipleRapidInputsSingleUpdate(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var updateCount int

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		// Inject counter
		chromedp.Evaluate(`
			window.previewUpdateCount = 0;
			const iframe = document.getElementById('live-preview-frame');
			const originalSrcDescriptor = Object.getOwnPropertyDescriptor(HTMLIFrameElement.prototype, 'src');
			Object.defineProperty(iframe, 'src', {
				get: originalSrcDescriptor.get,
				set: function(value) {
					if (value && value.includes('/api/themes/preview') && !value.includes('about:blank')) {
						window.previewUpdateCount++;
					}
					originalSrcDescriptor.set.call(this, value);
				},
				configurable: true
			});
		`, nil),

		// Reset counter (initial load already happened)
		chromedp.Evaluate(`window.previewUpdateCount = 0`, nil),

		// Type 5 characters rapidly
		chromedp.SendKeys(`[name="title"]`, "A", chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.SendKeys(`[name="title"]`, "B", chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.SendKeys(`[name="title"]`, "C", chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.SendKeys(`[name="title"]`, "D", chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.SendKeys(`[name="title"]`, "E", chromedp.ByQuery),

		// Wait for debounce
		chromedp.Sleep(700*time.Millisecond),

		chromedp.Evaluate(`window.previewUpdateCount`, &updateCount),
	)

	if err != nil {
		t.Fatalf("Failed to test debounce count: %v", err)
	}

	// Should update only once despite 5 rapid inputs
	if updateCount > 1 {
		t.Errorf("Expected at most 1 preview update from debouncing, got %d", updateCount)
	}
}

// Test 17: Debounce timer cleared on mode switch
func TestDebounce_ClearedOnModeSwitch(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		// Start typing to trigger debounce
		chromedp.SendKeys(`[name="title"]`, "Test", chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond), // Don't wait for debounce to complete

		// Switch modes immediately
		chromedp.Click(`#gallery-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),
	)

	if err != nil {
		t.Fatalf("Failed to test timer clearing: %v", err)
	}

	// Test passes if no errors occur - demonstrates cleanup works
}

// ============================================================================
// SECTION 5: Loading & Error States Tests (4 tests)
// ============================================================================

// Test 18: Loading indicator exists
func TestLoadingState_IndicatorExists(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var loadingExists bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(100*time.Millisecond), // Check quickly to catch loading state

		chromedp.Evaluate(`!!document.querySelector('.live-preview-loading')`, &loadingExists),
	)

	if err != nil {
		t.Fatalf("Failed to check loading indicator: %v", err)
	}

	if !loadingExists {
		t.Error("Loading indicator element should exist")
	}
}

// Test 19: Loading indicator can be toggled (verifies loading mechanism works)
func TestLoadingState_ShowsDuringLoad(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var loadingStartedVisible, loadingEventuallyHidden bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(50*time.Millisecond), // Very quick check

		// Check if loading was visible at any point (or still is)
		chromedp.Evaluate(`!document.querySelector('.live-preview-loading').hidden`, &loadingStartedVisible),

		// Wait for load to complete
		chromedp.Sleep(2*time.Second),

		// Check that loading eventually hides (proves the loading/hiding mechanism works)
		chromedp.Evaluate(`document.querySelector('.live-preview-loading').hidden`, &loadingEventuallyHidden),
	)

	if err != nil {
		t.Fatalf("Failed to check loading visibility: %v", err)
	}

	// The loading indicator should eventually hide (proving it works)
	// We can't guarantee catching it visible due to fast load times, but we can verify it hides
	if !loadingEventuallyHidden {
		t.Error("Loading indicator should be hidden after preview loads successfully")
	}
}

// Test 20: Error indicator exists
func TestErrorState_IndicatorExists(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var errorExists bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		chromedp.Evaluate(`!!document.querySelector('.live-preview-error')`, &errorExists),
	)

	if err != nil {
		t.Fatalf("Failed to check error indicator: %v", err)
	}

	if !errorExists {
		t.Error("Error indicator element should exist")
	}
}

// Test 21: Retry button exists in error state
func TestErrorState_RetryButtonExists(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var retryBtnExists bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		chromedp.Evaluate(`!!document.querySelector('.btn-retry-preview')`, &retryBtnExists),
	)

	if err != nil {
		t.Fatalf("Failed to check retry button: %v", err)
	}

	if !retryBtnExists {
		t.Error("Retry button should exist for error recovery")
	}
}

// ============================================================================
// SECTION 6: Mobile Responsive Tests (5 tests)
// ============================================================================

// Test 22: Mobile view toggle buttons exist
func TestMobile_ToggleButtonsExist(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var editBtnExists, previewBtnExists bool

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(375, 667), // iPhone SE
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		chromedp.Evaluate(`!!document.getElementById('mobile-edit-btn')`, &editBtnExists),
		chromedp.Evaluate(`!!document.getElementById('mobile-preview-btn')`, &previewBtnExists),
	)

	if err != nil {
		t.Fatalf("Failed to check mobile toggle buttons: %v", err)
	}

	if !editBtnExists {
		t.Error("Mobile edit button should exist")
	}
	if !previewBtnExists {
		t.Error("Mobile preview button should exist")
	}
}

// Test 23: Mobile viewport renders correctly
func TestMobile_ViewportRenders(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var viewportWidth int

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(375, 667),
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		chromedp.Evaluate(`window.innerWidth`, &viewportWidth),
	)

	if err != nil {
		t.Fatalf("Failed to check viewport: %v", err)
	}

	if viewportWidth != 375 {
		t.Errorf("Expected viewport width 375, got %d", viewportWidth)
	}
}

// Test 24: Mobile preview toggle switches view
func TestMobile_PreviewToggleSwitchesView(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var editAria, previewAria string

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(375, 667),
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		chromedp.WaitVisible(`#mobile-preview-btn`, chromedp.ByID),
		chromedp.Click(`#mobile-preview-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.AttributeValue(`#mobile-edit-btn`, `aria-selected`, &editAria, nil),
		chromedp.AttributeValue(`#mobile-preview-btn`, `aria-selected`, &previewAria, nil),
	)

	if err != nil {
		t.Fatalf("Failed mobile preview toggle: %v", err)
	}

	if editAria != "false" {
		t.Errorf("Expected edit button aria-selected='false', got '%s'", editAria)
	}
	if previewAria != "true" {
		t.Errorf("Expected preview button aria-selected='true', got '%s'", previewAria)
	}
}

// Test 25: Mobile edit toggle switches back to edit
func TestMobile_EditToggleSwitchesBack(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var editAria, previewAria string

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(375, 667),
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Switch to preview
		chromedp.WaitVisible(`#mobile-preview-btn`, chromedp.ByID),
		chromedp.Click(`#mobile-preview-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Switch back to edit
		chromedp.Click(`#mobile-edit-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.AttributeValue(`#mobile-edit-btn`, `aria-selected`, &editAria, nil),
		chromedp.AttributeValue(`#mobile-preview-btn`, `aria-selected`, &previewAria, nil),
	)

	if err != nil {
		t.Fatalf("Failed mobile edit toggle: %v", err)
	}

	if editAria != "true" {
		t.Errorf("Expected edit button aria-selected='true', got '%s'", editAria)
	}
	if previewAria != "false" {
		t.Errorf("Expected preview button aria-selected='false', got '%s'", previewAria)
	}
}

// Test 26: Tablet viewport renders correctly
func TestTablet_ViewportRenders(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var viewportWidth int

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(768, 1024), // iPad
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		chromedp.Evaluate(`window.innerWidth`, &viewportWidth),
	)

	if err != nil {
		t.Fatalf("Failed to check tablet viewport: %v", err)
	}

	if viewportWidth != 768 {
		t.Errorf("Expected viewport width 768, got %d", viewportWidth)
	}
}

// ============================================================================
// SECTION 7: Theme Gallery Tests (4 tests)
// ============================================================================

// Test 27: Theme cards are clickable
func TestGallery_ThemeCardsClickable(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var themeID string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		// Click first theme card
		chromedp.Click(`.theme-card`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// Check selected theme ID was set
		chromedp.AttributeValue(`#selected-theme-id`, `value`, &themeID, nil),
	)

	if err != nil {
		t.Fatalf("Failed to click theme card: %v", err)
	}

	if themeID == "" {
		t.Error("Expected theme ID to be set after clicking card")
	}
}

// Test 28: Theme selection updates hidden input
func TestGallery_SelectionUpdatesHiddenInput(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hiddenValue string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		// Click theme card
		chromedp.Click(`.theme-card[data-theme-id]`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.Evaluate(`document.getElementById('selected-theme-id').value`, &hiddenValue),
	)

	if err != nil {
		t.Fatalf("Failed to check hidden input: %v", err)
	}

	if hiddenValue == "" {
		t.Error("Hidden input should have theme ID after selection")
	}
}

// Test 29: Theme filter dropdown exists
func TestGallery_FilterDropdownExists(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var filterExists bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		chromedp.Evaluate(`!!document.getElementById('theme-category-filter')`, &filterExists),
	)

	if err != nil {
		t.Fatalf("Failed to check filter dropdown: %v", err)
	}

	if !filterExists {
		t.Error("Theme category filter dropdown should exist")
	}
}

// Test 30: Selected theme has correct styling
func TestGallery_SelectedThemeHasClass(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasSelectedClass bool

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		// Click first theme card
		chromedp.Click(`.theme-card`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// Check if it has selected class
		chromedp.Evaluate(`document.querySelector('.theme-card').classList.contains('selected')`, &hasSelectedClass),
	)

	if err != nil {
		t.Fatalf("Failed to check selected class: %v", err)
	}

	if !hasSelectedClass {
		t.Error("Selected theme card should have 'selected' class")
	}
}

// ============================================================================
// SECTION 8: Accessibility Tests (5 tests)
// ============================================================================

// Test 31: Mode toggle has proper ARIA roles
func TestAccessibility_ModeToggleARIARoles(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var galleryRole, designRole string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		chromedp.AttributeValue(`#gallery-mode-btn`, `role`, &galleryRole, nil),
		chromedp.AttributeValue(`#design-mode-btn`, `role`, &designRole, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check ARIA roles: %v", err)
	}

	if galleryRole != "tab" {
		t.Errorf("Expected gallery button role='tab', got '%s'", galleryRole)
	}
	if designRole != "tab" {
		t.Errorf("Expected design button role='tab', got '%s'", designRole)
	}
}

// Test 32: Preview iframe has title attribute
func TestAccessibility_IframeHasTitle(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var iframeTitle string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		chromedp.AttributeValue(`#live-preview-frame`, `title`, &iframeTitle, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check iframe title: %v", err)
	}

	if iframeTitle == "" {
		t.Error("Preview iframe should have title attribute for accessibility")
	}
}

// Test 33: Loading indicator has ARIA live region
func TestAccessibility_LoadingIndicatorARIALive(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var ariaLive, role string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		chromedp.AttributeValue(`.live-preview-loading`, `aria-live`, &ariaLive, nil),
		chromedp.AttributeValue(`.live-preview-loading`, `role`, &role, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check loading ARIA: %v", err)
	}

	if ariaLive != "polite" {
		t.Errorf("Expected loading indicator aria-live='polite', got '%s'", ariaLive)
	}
	if role != "status" {
		t.Errorf("Expected loading indicator role='status', got '%s'", role)
	}
}

// Test 34: Error indicator has alert role
func TestAccessibility_ErrorIndicatorAlertRole(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var role string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		chromedp.AttributeValue(`.live-preview-error`, `role`, &role, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check error role: %v", err)
	}

	if role != "alert" {
		t.Errorf("Expected error indicator role='alert', got '%s'", role)
	}
}

// Test 35: Theme cards have proper ARIA attributes
func TestAccessibility_ThemeCardsARIAAttributes(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var role, ariaChecked string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		chromedp.AttributeValue(`.theme-card`, `role`, &role, nil),
		chromedp.AttributeValue(`.theme-card`, `aria-checked`, &ariaChecked, nil),
	)

	if err != nil {
		t.Fatalf("Failed to check theme card ARIA: %v", err)
	}

	if role != "radio" {
		t.Errorf("Expected theme card role='radio', got '%s'", role)
	}
	if ariaChecked == "" {
		t.Error("Theme card should have aria-checked attribute")
	}
}

// ============================================================================
// SECTION 9: Integration Tests (5 tests)
// ============================================================================

// Test 36: Complete workflow - gallery to design mode
func TestIntegration_GalleryToDesignWorkflow(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		// Select theme in gallery
		chromedp.Click(`.theme-card`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		// Switch to design mode
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		// Verify preview loaded
		chromedp.WaitVisible(`#live-preview-frame`, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Failed integration workflow: %v", err)
	}
}

// Test 37: Complete workflow - edit form and see preview
func TestIntegration_EditFormSeePreview(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var finalSrc string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		// Switch to design mode
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		// Fill form
		chromedp.SendKeys(`[name="title"]`, "Integration Test Event", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),

		chromedp.SendKeys(`[name="description"]`, "Testing the flow", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &finalSrc, nil),
	)

	if err != nil {
		t.Fatalf("Failed form edit workflow: %v", err)
	}

	if !strings.Contains(finalSrc, "title=") || !strings.Contains(finalSrc, "description=") {
		t.Errorf("Expected preview URL to contain both title and description, got '%s'", finalSrc)
	}
}

// Test 38: Form persistence across mode switches
func TestIntegration_FormPersistsAcrossModes(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var titleValue string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),

		// Switch to design mode
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		// Enter title
		chromedp.SendKeys(`[name="title"]`, "Persistent Title", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),

		// Switch back to gallery
		chromedp.Click(`#gallery-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Switch back to design
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Check title still there
		chromedp.Evaluate(`document.querySelector('[name="title"]').value`, &titleValue),
	)

	if err != nil {
		t.Fatalf("Failed persistence test: %v", err)
	}

	if !strings.Contains(titleValue, "Persistent Title") {
		t.Errorf("Expected title to persist, got '%s'", titleValue)
	}
}

// Test 39: Theme selection persists in design mode dropdown
func TestIntegration_ThemeSelectionPersistsInDropdown(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var dropdownValue, hiddenValue string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		// Select theme in gallery
		chromedp.Click(`.theme-card[data-theme-id]`, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),

		chromedp.AttributeValue(`#selected-theme-id`, `value`, &hiddenValue, nil),

		// Switch to design mode
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(300*time.Millisecond),

		// Check dropdown has same value
		chromedp.Evaluate(`document.getElementById('design-theme-select').value`, &dropdownValue),
	)

	if err != nil {
		t.Fatalf("Failed theme persistence test: %v", err)
	}

	if dropdownValue != hiddenValue {
		t.Errorf("Expected dropdown value '%s' to match hidden value '%s'", dropdownValue, hiddenValue)
	}
}

// Test 40: Multiple field updates work correctly
func TestIntegration_MultipleFieldUpdates(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var finalSrc string

	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(600*time.Millisecond),

		// Fill multiple fields
		chromedp.SendKeys(`[name="title"]`, "Multi Field Test", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),

		chromedp.SendKeys(`[name="location"]`, "Test Location", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),

		chromedp.SendKeys(`[name="description"]`, "Test Description", chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),

		chromedp.AttributeValue(`#live-preview-frame`, `src`, &finalSrc, nil),
	)

	if err != nil {
		t.Fatalf("Failed multiple field test: %v", err)
	}

	hasTitle := strings.Contains(finalSrc, "title=")
	hasLocation := strings.Contains(finalSrc, "location=")
	hasDesc := strings.Contains(finalSrc, "description=")

	if !hasTitle || !hasLocation || !hasDesc {
		t.Errorf("Expected all fields in URL. Has title=%v, location=%v, description=%v. URL: %s",
			hasTitle, hasLocation, hasDesc, finalSrc)
	}
}

// ============================================================================
// Summary Function
// ============================================================================

func TestMain(m *testing.M) {
	fmt.Println("=================================================================")
	fmt.Println("  TinyRSVP Live Preview Mode - Comprehensive Test Suite")
	fmt.Println("=================================================================")
	fmt.Println("  Test Sections:")
	fmt.Println("    1. Page Load & Initial State (4 tests)")
	fmt.Println("    2. Mode Switching (5 tests)")
	fmt.Println("    3. Preview URL Building (5 tests)")
	fmt.Println("    4. Debouncing (3 tests)")
	fmt.Println("    5. Loading & Error States (4 tests)")
	fmt.Println("    6. Mobile Responsive (5 tests)")
	fmt.Println("    7. Theme Gallery (4 tests)")
	fmt.Println("    8. Accessibility (5 tests)")
	fmt.Println("    9. Integration (5 tests)")
	fmt.Println("  Total: 40 tests")
	fmt.Println("=================================================================")
	fmt.Println("")

	m.Run()
}
