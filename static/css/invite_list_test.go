package css

import (
	"os"
	"strings"
	"testing"
)

func TestInviteListCSSExists(t *testing.T) {
	if _, err := os.Stat("invite_list.css"); os.IsNotExist(err) {
		t.Fatal("invite_list.css does not exist")
	}
}

func TestInviteListCSSContent(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".invite-list",
		".invite-list-header",
		".invite-list-title",
		".invite-stats",
		".invite-filters",
		".invite-search",
		".invite-table",
		".invite-card",
		".invite-actions",
		".invite-empty",
		".invite-pagination",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing required class: %s", class)
		}
	}
}

func TestInviteListUsesVariables(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	requiredVariables := []string{
		"var(--spacing-",
		"var(--color-",
		"var(--font-",
		"var(--radius-",
	}

	for _, variable := range requiredVariables {
		if !strings.Contains(css, variable) {
			t.Errorf("CSS should use variable pattern: %s", variable)
		}
	}
}

func TestInviteListResponsive(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	breakpoints := []string{
		"@media (min-width: 768px)",
		"@media (min-width: 1024px)",
	}

	for _, breakpoint := range breakpoints {
		if !strings.Contains(css, breakpoint) {
			t.Errorf("Missing responsive breakpoint: %s", breakpoint)
		}
	}
}

func TestInviteListTableComponents(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	tableComponents := []string{
		".invite-table",
		".invite-table-row",
		".invite-table-checkbox",
		".invite-table-name",
		".invite-table-email",
		".invite-table-status",
		".invite-table-actions",
	}

	for _, component := range tableComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing table component: %s", component)
		}
	}
}

func TestInviteListCardComponents(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	cardComponents := []string{
		".invite-card",
		".invite-card-header",
		".invite-card-body",
		".invite-card-footer",
	}

	for _, component := range cardComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing card component: %s", component)
		}
	}
}

func TestInviteListStatusBadges(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	statusClasses := []string{
		".invite-status-badge",
		".invite-status-draft",
		".invite-status-sent",
		".invite-status-viewed",
		".invite-status-responded",
		".invite-status-revoked",
	}

	for _, class := range statusClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing status badge class: %s", class)
		}
	}
}

func TestInviteListFilterComponents(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	filterComponents := []string{
		".invite-filters",
		".filter-group",
		".filter-select",
		".search-input",
	}

	for _, component := range filterComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing filter component: %s", component)
		}
	}
}

func TestInviteListStates(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	states := []string{
		".invite-list-loading",
		".invite-empty",
		".loading-spinner",
	}

	for _, state := range states {
		if !strings.Contains(css, state) {
			t.Errorf("Missing state: %s", state)
		}
	}
}

func TestInviteListPagination(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	paginationComponents := []string{
		".invite-pagination",
		".pagination",
		".pagination-item",
		".pagination-item.active",
		".pagination-prev",
		".pagination-next",
	}

	for _, component := range paginationComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing pagination component: %s", component)
		}
	}
}

func TestInviteListAccessibility(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ":focus") {
		t.Error("CSS should include focus styles for accessibility")
	}
}

func TestInviteListNoHardcodedValues(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	hardcodedPatterns := []string{
		"#fff",
		"#000",
		"16px",
		"14px",
	}

	for _, pattern := range hardcodedPatterns {
		if strings.Contains(css, pattern) {
			t.Errorf("CSS should not contain hardcoded value: %s (use CSS variables instead)", pattern)
		}
	}
}

func TestInviteListBulkActions(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	bulkActionComponents := []string{
		".bulk-actions",
		".bulk-checkbox",
		".invite-actions-bar",
	}

	for _, component := range bulkActionComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing bulk action component: %s", component)
		}
	}
}

func TestInviteListMobileFirst(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".invite-table") {
		t.Error("CSS should include table styles for desktop")
	}

	if !strings.Contains(css, ".invite-card") {
		t.Error("CSS should include card styles for mobile")
	}
}

func TestInviteListNoSyntaxErrors(t *testing.T) {
	content, err := os.ReadFile("invite_list.css")
	if err != nil {
		t.Fatalf("Failed to read invite_list.css: %v", err)
	}

	css := string(content)

	openBraces := strings.Count(css, "{")
	closeBraces := strings.Count(css, "}")

	if openBraces != closeBraces {
		t.Errorf("CSS has unbalanced braces: %d open, %d close", openBraces, closeBraces)
	}
}
