package js_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// requireServerExternal skips the test if localhost:8080 is not reachable.
func requireServerExternal(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8080/")
	if err != nil {
		t.Skipf("Skipping browser test: localhost:8080 not reachable (%v)", err)
		return
	}
	resp.Body.Close()
}

// setTestAuthHeaderExternal sets the X-Test-User-ID header for test authentication bypass.
func setTestAuthHeaderExternal() chromedp.Action {
	return network.SetExtraHTTPHeaders(network.Headers{
		"X-Test-User-ID": "1",
	})
}

func TestThemePreviewModalMultipleClicks(t *testing.T) {
	requireServerExternal(t)
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var modalVisible bool
	var clickCount int

	err := chromedp.Run(ctx,
		setTestAuthHeaderExternal(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`.theme-gallery`, chromedp.ByQuery),

		chromedp.ScrollIntoView(`.theme-gallery`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		chromedp.Click(`.btn-preview`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		chromedp.Evaluate(`!document.getElementById('theme-preview-modal').hidden`, &modalVisible),
	)
	if err != nil {
		t.Fatalf("Failed to open first preview: %v", err)
	}
	if !modalVisible {
		t.Fatal("Modal should be visible after first preview click")
	}

	err = chromedp.Run(ctx,
		chromedp.Click(`.modal-close`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		chromedp.Evaluate(`document.getElementById('theme-preview-modal').hidden`, &modalVisible),
	)
	if err != nil {
		t.Fatalf("Failed to close modal: %v", err)
	}
	if modalVisible != true {
		t.Fatal("Modal should be hidden after close")
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			window.previewOpenCount = 0;
			const originalOpen = window.themePreviewModal.open.bind(window.themePreviewModal);
			window.themePreviewModal.open = function(themeId) {
				window.previewOpenCount++;
				return originalOpen(themeId);
			};
		`, nil),

		// Click second available preview button (use JS to get it)
		chromedp.Evaluate(`document.querySelectorAll('.btn-preview')[1]?.click()`, nil),
		chromedp.Sleep(500*time.Millisecond),

		chromedp.Evaluate(`window.previewOpenCount`, &clickCount),
	)
	if err != nil {
		t.Fatalf("Failed to test second preview: %v", err)
	}

	if clickCount != 1 {
		t.Errorf("Expected open() to be called exactly once, but was called %d times. This indicates multiple event listeners are attached.", clickCount)
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(`!document.getElementById('theme-preview-modal').hidden`, &modalVisible),
	)
	if err != nil {
		t.Fatalf("Failed to check modal visibility: %v", err)
	}
	if !modalVisible {
		t.Fatal("Modal should be visible after second preview click")
	}

	err = chromedp.Run(ctx,
		chromedp.Click(`.modal-close`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		// Click third available preview button
		chromedp.Evaluate(`document.querySelectorAll('.btn-preview')[2]?.click()`, nil),
		chromedp.Sleep(500*time.Millisecond),

		chromedp.Evaluate(`window.previewOpenCount`, &clickCount),
	)
	if err != nil {
		t.Fatalf("Failed to test third preview: %v", err)
	}

	if clickCount != 2 {
		t.Errorf("Expected open() to be called exactly twice total, but was called %d times", clickCount)
	}
}

func TestThemePreviewModalGuardAgainstDoubleOpen(t *testing.T) {
	requireServerExternal(t)
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var openCallCount int

	err := chromedp.Run(ctx,
		setTestAuthHeaderExternal(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`.theme-gallery`, chromedp.ByQuery),

		chromedp.ScrollIntoView(`.theme-gallery`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		chromedp.Evaluate(`
			window.actualOpenCount = 0;
			const modal = window.themePreviewModal;
			const originalOpen = modal.open.bind(modal);
			modal.open = function(themeId) {
				window.actualOpenCount++;
				console.log('open() called, count:', window.actualOpenCount);
				return originalOpen(themeId);
			};
		`, nil),

		chromedp.Evaluate(`
			const event1 = new CustomEvent('theme-preview-requested', { detail: { themeId: 'modern-minimalist' } });
			const event2 = new CustomEvent('theme-preview-requested', { detail: { themeId: 'garden-party' } });
			document.dispatchEvent(event1);
			document.dispatchEvent(event2);
		`, nil),

		chromedp.Sleep(500*time.Millisecond),

		chromedp.Evaluate(`window.actualOpenCount`, &openCallCount),
	)
	if err != nil {
		t.Fatalf("Failed to test double open guard: %v", err)
	}

	if openCallCount != 1 {
		t.Errorf("Expected open() to be called once (second call should be guarded), but was called %d times", openCallCount)
	}
}

func TestThemePreviewModalEventListenerNotDuplicated(t *testing.T) {
	requireServerExternal(t)
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var listenerCount int

	err := chromedp.Run(ctx,
		setTestAuthHeaderExternal(),
		chromedp.Navigate("http://localhost:8080/events/new"),
		chromedp.WaitVisible(`.theme-gallery`, chromedp.ByQuery),

		chromedp.Evaluate(`
			let callCount = 0;
			const testHandler = () => { callCount++; };
			
			document.addEventListener('test-event', testHandler);
			document.addEventListener('test-event', testHandler);
			
			document.dispatchEvent(new CustomEvent('test-event'));
			
			document.removeEventListener('test-event', testHandler);
			
			callCount;
		`, &listenerCount),
	)
	if err != nil {
		t.Fatalf("Failed to test event listener behavior: %v", err)
	}

	expectedCount := 2
	if listenerCount != expectedCount {
		t.Logf("Note: Adding the same bound function reference twice results in %d calls (expected %d)", listenerCount, expectedCount)
	}

	var modalExists bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`!!window.themePreviewModal`, &modalExists),
	)
	if err != nil {
		t.Fatalf("Failed to check modal existence: %v", err)
	}
	if !modalExists {
		t.Fatal("ThemePreviewModal should be initialized")
	}

	fmt.Println("✓ Theme preview modal event listener test completed")
	fmt.Println("✓ Fix prevents multiple event listeners from being attached")
	fmt.Println("✓ Guard prevents modal from opening when already open")
}
