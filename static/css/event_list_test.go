package css

import (
	"os"
	"strings"
	"testing"
)

func TestEventListCSSExists(t *testing.T) {
	if _, err := os.Stat("event_list.css"); os.IsNotExist(err) {
		t.Fatal("event_list.css does not exist")
	}
}

func TestEventListCSSContent(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".event-list",
		".event-card",
		".event-filters",
		".event-search",
		".event-sort",
		".event-pagination",
		".event-empty",
		".event-actions",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing required class: %s", class)
		}
	}
}

func TestEventListUsesVariables(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
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

func TestEventListResponsive(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
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

func TestEventListCardComponents(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	css := string(content)

	cardComponents := []string{
		".event-card-header",
		".event-card-body",
		".event-card-footer",
		".event-card-title",
		".event-card-meta",
		".event-card-status",
	}

	for _, component := range cardComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing event card component: %s", component)
		}
	}
}

func TestEventListFilterComponents(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	css := string(content)

	filterComponents := []string{
		".filter-group",
		".filter-select",
		".search-input",
		".sort-select",
	}

	for _, component := range filterComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing filter component: %s", component)
		}
	}
}

func TestEventListStates(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	css := string(content)

	states := []string{
		".event-list-loading",
		".event-empty",
		".event-card:hover",
	}

	for _, state := range states {
		if !strings.Contains(css, state) {
			t.Errorf("Missing state: %s", state)
		}
	}
}

func TestEventListPagination(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	css := string(content)

	paginationComponents := []string{
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

func TestEventListAccessibility(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ":focus") {
		t.Error("CSS should include focus styles for accessibility")
	}
}

func TestEventListNoHardcodedValues(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
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
