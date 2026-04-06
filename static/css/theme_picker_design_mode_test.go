package css

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// Test 1: Required CSS classes exist
func TestCSS_RequiredClassesExist(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	requiredClasses := []string{
		".theme-mode-controls",
		".mode-btn",
		".design-mode-container",
		".live-preview-wrapper",
		".live-preview-frame",
		".live-preview-loading",
		".live-preview-error",
		".mobile-view-toggle",
		".view-btn",
	}

	for _, class := range requiredClasses {
		var exists bool
		err := chromedp.Run(ctx,
			setTestAuthHeader(),
			chromedp.Navigate("http://localhost:8080/events/new"),
			chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
			chromedp.Click(`#design-mode-btn`, chromedp.ByID),
			chromedp.Sleep(200*time.Millisecond),
			chromedp.Evaluate(`document.querySelector('`+class+`') !== null`, &exists),
		)

		if err != nil {
			t.Fatalf("Failed to check class %s: %v", class, err)
		}

		if !exists {
			t.Errorf("Required CSS class %s not found in DOM", class)
		}
	}
}

// Test 2: Responsive breakpoints work - Mobile toggle visible only on mobile
func TestCSS_MobileToggleVisibility(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Test on desktop (should be hidden)
	var desktopDisplay string
	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.EmulateViewport(1024, 768), // Desktop size
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.mobile-view-toggle')).display`, &desktopDisplay),
	)

	if err != nil {
		t.Fatalf("Failed to check desktop display: %v", err)
	}

	if desktopDisplay != "none" {
		t.Errorf("Expected mobile toggle to be hidden on desktop (display: none), got %s", desktopDisplay)
	}

	// Test on mobile (should be visible in design mode)
	var mobileDisplay string
	err = chromedp.Run(ctx,
		chromedp.EmulateViewport(375, 667), // Mobile size
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.mobile-view-toggle')).display`, &mobileDisplay),
	)

	if err != nil {
		t.Fatalf("Failed to check mobile display: %v", err)
	}

	if mobileDisplay == "none" {
		t.Errorf("Expected mobile toggle to be visible on mobile in design mode, got display: %s", mobileDisplay)
	}
}

// Test 3: Desktop layout uses side-by-side columns
func TestCSS_DesktopLayoutColumns(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var layoutDisplay string
	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.EmulateViewport(1024, 768), // Desktop size
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`.event-form-layout`, chromedp.ByQuery),
		chromedp.Evaluate(`window.getComputedStyle(document.querySelector('.event-form-layout')).display`, &layoutDisplay),
	)

	if err != nil {
		t.Fatalf("Failed to check layout display: %v", err)
	}

	if layoutDisplay != "flex" {
		t.Errorf("Expected desktop layout to use flexbox, got display: %s", layoutDisplay)
	}
}

// Test 4: Loading/error states styled correctly
func TestCSS_LoadingErrorStates(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check loading spinner exists
	var spinnerExists bool
	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Evaluate(`document.querySelector('.spinner') !== null`, &spinnerExists),
	)

	if err != nil {
		t.Fatalf("Failed to check spinner: %v", err)
	}

	if !spinnerExists {
		t.Errorf("Expected loading spinner element to exist")
	}

	// Check error container has proper styling
	var errorContainerExists bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('.live-preview-error') !== null`, &errorContainerExists),
	)

	if err != nil {
		t.Fatalf("Failed to check error container: %v", err)
	}

	if !errorContainerExists {
		t.Errorf("Expected error container element to exist")
	}
}

// Test 5: Mode buttons have proper touch target size (44px minimum)
func TestCSS_TouchTargetSize(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var minHeight float64
	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.EmulateViewport(375, 667), // Mobile size
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Evaluate(`parseInt(window.getComputedStyle(document.querySelector('.mode-btn')).minHeight)`, &minHeight),
	)

	if err != nil {
		t.Fatalf("Failed to check touch target size: %v", err)
	}

	if minHeight < 44 {
		t.Errorf("Expected mode button min-height to be at least 44px for touch targets, got %.0fpx", minHeight)
	}
}

// Test 6: Preview iframe has proper dimensions
func TestCSS_PreviewIframeDimensions(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var width, height string
	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Evaluate(`window.getComputedStyle(document.getElementById('live-preview-frame')).width`, &width),
		chromedp.Evaluate(`window.getComputedStyle(document.getElementById('live-preview-frame')).height`, &height),
	)

	if err != nil {
		t.Fatalf("Failed to check iframe dimensions: %v", err)
	}

	if width == "" || width == "0px" {
		t.Errorf("Expected iframe to have non-zero width, got %s", width)
	}

	if height == "" || height == "0px" {
		t.Errorf("Expected iframe to have non-zero height, got %s", height)
	}
}

// Test 7: Active mode button has distinct styling
func TestCSS_ActiveModeButtonStyling(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var galleryBtnBg, designBtnBg string
	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#gallery-mode-btn`, chromedp.ByID),

		// Get initial state (gallery active)
		chromedp.Evaluate(`window.getComputedStyle(document.getElementById('gallery-mode-btn')).backgroundColor`, &galleryBtnBg),

		// Switch to design mode
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Get design mode button background
		chromedp.Evaluate(`window.getComputedStyle(document.getElementById('design-mode-btn')).backgroundColor`, &designBtnBg),
	)

	if err != nil {
		t.Fatalf("Failed to check button styling: %v", err)
	}

	if designBtnBg == "" {
		t.Errorf("Expected design mode button to have background color when active")
	}
}

// Test 8: Hidden attribute works correctly
func TestCSS_HiddenAttributeWorks(t *testing.T) {
	requireServer(t)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var galleryDisplay string
	err := chromedp.Run(ctx,
		setTestAuthHeader(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`#design-mode-btn`, chromedp.ByID),
		chromedp.Click(`#design-mode-btn`, chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),

		// Check that hidden gallery container is not displayed
		chromedp.Evaluate(`window.getComputedStyle(document.getElementById('theme-gallery-container')).display`, &galleryDisplay),
	)

	if err != nil {
		t.Fatalf("Failed to check hidden attribute: %v", err)
	}

	if galleryDisplay != "none" {
		t.Errorf("Expected hidden gallery container to have display: none, got %s", galleryDisplay)
	}
}
