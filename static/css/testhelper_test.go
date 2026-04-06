package css

import (
	"net/http"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// requireServer skips the test if the local dev server is not running at localhost:8080.
// This prevents browser/Chromedp tests from failing in CI or during unit-test runs.
func requireServer(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8080/")
	if err != nil {
		t.Skipf("Skipping browser test: localhost:8080 not reachable (%v)", err)
		return
	}
	resp.Body.Close()
}

// setTestAuthHeader sets the X-Test-User-ID header for test authentication bypass.
func setTestAuthHeader() chromedp.Action {
	return network.SetExtraHTTPHeaders(network.Headers{
		"X-Test-User-ID": "1",
	})
}
