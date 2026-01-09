package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestInviteListIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	inviteListContent, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	variables := string(variablesContent)
	inviteList := string(inviteListContent)

	requiredVariables := []string{
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-5",
		"--spacing-6",
		"--spacing-8",
		"--spacing-10",
		"--spacing-12",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-text-tertiary",
		"--color-background",
		"--color-surface",
		"--color-border",
		"--color-border-focus",
		"--color-primary-600",
		"--color-primary-100",
		"--color-success",
		"--color-success-light",
		"--color-warning",
		"--color-warning-light",
		"--color-info",
		"--color-info-light",
		"--color-error",
		"--color-error-light",
		"--font-size-xs",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-size-xl",
		"--font-size-2xl",
		"--font-size-4xl",
		"--font-weight-medium",
		"--font-weight-semibold",
		"--font-weight-bold",
		"--radius-md",
		"--radius-lg",
		"--radius-full",
	}

	for _, variable := range requiredVariables {
		if !strings.Contains(variables, variable+":") {
			t.Errorf("variables.css missing required variable: %s", variable)
		}
	}

	usedVariables := []string{
		"var(--spacing-",
		"var(--color-",
		"var(--font-",
		"var(--radius-",
	}

	for _, pattern := range usedVariables {
		if !strings.Contains(inviteList, pattern) {
			t.Errorf("invite_list.css should use variable pattern: %s", pattern)
		}
	}
}

func TestInviteListIntegrationWithGrid(t *testing.T) {
	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	inviteListContent, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	grid := string(gridContent)
	inviteList := string(inviteListContent)

	if !strings.Contains(grid, "display: grid") {
		t.Error("grid.css should define grid display")
	}

	if !strings.Contains(inviteList, "display: grid") {
		t.Error("invite_list.css should use CSS Grid for layout")
	}

	if !strings.Contains(inviteList, "grid-template-columns") {
		t.Error("invite_list.css should use grid-template-columns")
	}
}

func TestInviteListHTTPServing(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(content)
	})

	req := httptest.NewRequest("GET", "/static/css/invite_list.css", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %s", contentType)
	}
}

func TestInviteListFileSize(t *testing.T) {
	info, err := os.Stat("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to stat invite_list.css: %v", err)
	}

	maxSize := int64(50 * 1024)
	if info.Size() > maxSize {
		t.Errorf("invite_list.css is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestInviteListResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	breakpoints := map[string]string{
		"tablet":  "@media (min-width: 768px)",
		"desktop": "@media (min-width: 1024px)",
	}

	for name, breakpoint := range breakpoints {
		if !strings.Contains(css, breakpoint) {
			t.Errorf("Missing %s breakpoint: %s", name, breakpoint)
		}
	}
}

func TestInviteListAccessibilityFeatures(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	accessibilityFeatures := []string{
		":focus",
		"outline:",
		"cursor:",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(css, feature) {
			t.Errorf("Missing accessibility feature: %s", feature)
		}
	}
}

func TestInviteListAnimations(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "@keyframes") {
		t.Error("invite_list.css should include keyframe animations")
	}

	if !strings.Contains(css, "animation:") {
		t.Error("invite_list.css should use animations")
	}

	if !strings.Contains(css, "transition:") {
		t.Error("invite_list.css should use transitions for smooth interactions")
	}
}

func TestInviteListIntegrationWithButtons(t *testing.T) {
	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	inviteListContent, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	buttons := string(buttonsContent)
	inviteList := string(inviteListContent)

	if !strings.Contains(buttons, ".btn") {
		t.Error("buttons.css should define .btn class")
	}

	if strings.Contains(inviteList, ".btn") {
		t.Log("invite_list.css references button styles from buttons.css")
	}
}

func TestInviteListIntegrationWithForms(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	inviteListContent, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	forms := string(formsContent)
	inviteList := string(inviteListContent)

	if !strings.Contains(forms, "input") || !strings.Contains(forms, "select") {
		t.Error("forms.css should define input and select styles")
	}

	formElements := []string{
		"input",
		"select",
	}

	for _, element := range formElements {
		if strings.Contains(inviteList, element) {
			t.Logf("invite_list.css uses form element: %s", element)
		}
	}
}

func TestInviteListMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	mobileFirstIndicators := []string{
		".invite-cards",
		"@media (min-width:",
	}

	for _, indicator := range mobileFirstIndicators {
		if !strings.Contains(css, indicator) {
			t.Errorf("Missing mobile-first indicator: %s", indicator)
		}
	}

	desktopOnlyIndex := strings.Index(css, ".invite-table-container")
	mediaQueryIndex := strings.Index(css, "@media (min-width: 1024px)")

	if desktopOnlyIndex > 0 && mediaQueryIndex > 0 {
		if desktopOnlyIndex < mediaQueryIndex {
			t.Log("Mobile-first approach: base styles before media queries")
		}
	}
}

func TestInviteListTableHiddenOnMobile(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".invite-table-container") {
		t.Error("CSS should define .invite-table-container")
	}

	if !strings.Contains(css, "display: none") {
		t.Error("CSS should hide table on mobile using display: none")
	}
}

func TestInviteListCardsVisibleOnMobile(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".invite-cards") {
		t.Error("CSS should define .invite-cards for mobile view")
	}

	if !strings.Contains(css, ".invite-card") {
		t.Error("CSS should define .invite-card for mobile view")
	}
}
