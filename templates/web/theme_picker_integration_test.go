package web

import (
	"html/template"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestThemePickerPartialIntegration(t *testing.T) {
	funcMap := template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	}

	tmpl, err := template.New("theme_picker").Funcs(funcMap).ParseFiles(
		"partials/theme_picker.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse theme_picker template: %v", err)
	}

	data := struct {
		Themes          []*models.Template
		SelectedThemeID int64
	}{
		Themes: []*models.Template{
			{
				ID:           1,
				Name:         "Plain Theme",
				Type:         models.TemplateTypeRSVPPage,
				Category:     models.CategoryPlain,
				Description:  "Simple plain text theme",
				ThumbnailURL: stringPtr("/static/images/themes/plain-thumbnail.svg"),
				Tags:         []string{"simple", "minimal"},
				SortOrder:    1,
			},
			{
				ID:           2,
				Name:         "Card Theme",
				Type:         models.TemplateTypeRSVPPage,
				Category:     models.CategoryCard,
				Description:  "Elegant card design",
				ThumbnailURL: stringPtr("/static/images/themes/card-thumbnail.svg"),
				Tags:         []string{"elegant", "modern"},
				SortOrder:    2,
			},
		},
		SelectedThemeID: 1,
	}

	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, "theme_picker", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	t.Run("renders theme picker container", func(t *testing.T) {
		if !strings.Contains(html, `class="theme-picker"`) {
			t.Error("Should render theme-picker container")
		}
	})

	t.Run("renders filter dropdown", func(t *testing.T) {
		if !strings.Contains(html, `id="theme-category-filter"`) {
			t.Error("Should render theme category filter")
		}

		if !strings.Contains(html, "All Themes") {
			t.Error("Should have 'All Themes' option")
		}

		if !strings.Contains(html, "Plain Text") {
			t.Error("Should have 'Plain Text' option")
		}

		if !strings.Contains(html, "Card Designs") {
			t.Error("Should have 'Card Designs' option")
		}
	})

	t.Run("renders theme gallery", func(t *testing.T) {
		if !strings.Contains(html, `class="theme-gallery"`) {
			t.Error("Should render theme-gallery")
		}

		// Theme gallery uses role="tabpanel" (tabs pattern) or role="region"/"radiogroup" for accessibility
		if !strings.Contains(html, `role="tabpanel"`) && !strings.Contains(html, `role="region"`) && !strings.Contains(html, `role="radiogroup"`) {
			t.Error("Theme gallery should have a landmark role (tabpanel, region, or radiogroup)")
		}

		if !strings.Contains(html, `aria-labelledby=`) && !strings.Contains(html, `aria-label="Theme gallery"`) && !strings.Contains(html, `aria-label="Select theme"`) {
			t.Error("Theme gallery should have aria-labelledby or aria-label")
		}
	})

	t.Run("renders theme cards", func(t *testing.T) {
		if !strings.Contains(html, "Plain Theme") {
			t.Error("Should render Plain Theme")
		}

		if !strings.Contains(html, "Card Theme") {
			t.Error("Should render Card Theme")
		}

		if !strings.Contains(html, `data-theme-id="1"`) {
			t.Error("Should have data-theme-id attribute")
		}

		if !strings.Contains(html, `data-category="plain"`) {
			t.Error("Should have data-category attribute")
		}
	})

	t.Run("marks selected theme", func(t *testing.T) {
		if !strings.Contains(html, `class="theme-card selected"`) {
			t.Error("Selected theme should have 'selected' class")
		}

		if !strings.Contains(html, `aria-checked="true"`) {
			t.Error("Selected theme should have aria-checked='true'")
		}

		if !strings.Contains(html, `tabindex="0"`) {
			t.Error("Selected theme should have tabindex='0'")
		}
	})

	t.Run("renders thumbnails", func(t *testing.T) {
		if !strings.Contains(html, `class="theme-thumbnail"`) {
			t.Error("Should render theme thumbnails")
		}

		if !strings.Contains(html, "/static/images/themes/plain-thumbnail.svg") {
			t.Error("Should render thumbnail image URL")
		}

		if !strings.Contains(html, `loading="lazy"`) {
			t.Error("Thumbnail images should use lazy loading")
		}
	})

	t.Run("renders theme info", func(t *testing.T) {
		if !strings.Contains(html, `class="theme-name"`) {
			t.Error("Should render theme name")
		}

		if !strings.Contains(html, `class="theme-description"`) {
			t.Error("Should render theme description")
		}

		if !strings.Contains(html, "Simple plain text theme") {
			t.Error("Should render theme description text")
		}
	})

	t.Run("renders theme tags", func(t *testing.T) {
		if !strings.Contains(html, `class="theme-tags"`) {
			t.Error("Should render theme tags container")
		}

		if !strings.Contains(html, "simple") {
			t.Error("Should render 'simple' tag")
		}

		if !strings.Contains(html, "minimal") {
			t.Error("Should render 'minimal' tag")
		}
	})

	t.Run("renders action buttons", func(t *testing.T) {
		if !strings.Contains(html, `class="btn-preview"`) {
			t.Error("Should render preview button")
		}

		if !strings.Contains(html, `class="btn-select"`) {
			t.Error("Should render select button")
		}

		if !strings.Contains(html, `aria-label="Preview Plain Theme theme"`) {
			t.Error("Preview button should have descriptive aria-label")
		}

		if !strings.Contains(html, `aria-label="Select Plain Theme theme"`) {
			t.Error("Select button should have descriptive aria-label")
		}
	})

	t.Run("renders hidden input", func(t *testing.T) {
		if !strings.Contains(html, `id="selected-theme-id"`) {
			t.Error("Should render hidden input for form submission")
		}

		if !strings.Contains(html, `name="template_id"`) {
			t.Error("Hidden input should have name='template_id'")
		}

		if !strings.Contains(html, `value="1"`) {
			t.Error("Hidden input should have selected theme ID as value")
		}
	})
}

func TestThemePickerPartialWithoutThumbnail(t *testing.T) {
	funcMap := template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	}

	tmpl, err := template.New("theme_picker").Funcs(funcMap).ParseFiles(
		"partials/theme_picker.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse theme_picker template: %v", err)
	}

	data := struct {
		Themes          []*models.Template
		SelectedThemeID int64
	}{
		Themes: []*models.Template{
			{
				ID:          1,
				Name:        "No Thumbnail Theme",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryPlain,
				Description: "Theme without thumbnail",
				SortOrder:   1,
			},
		},
		SelectedThemeID: 1,
	}

	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, "theme_picker", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "default-thumbnail.svg") {
		t.Error("Should use default thumbnail when ThumbnailURL is nil")
	}
}

func TestThemePickerPartialEmptyThemes(t *testing.T) {
	funcMap := template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	}

	tmpl, err := template.New("theme_picker").Funcs(funcMap).ParseFiles(
		"partials/theme_picker.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse theme_picker template: %v", err)
	}

	data := struct {
		Themes          []*models.Template
		SelectedThemeID int64
	}{
		Themes:          []*models.Template{},
		SelectedThemeID: 0,
	}

	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, "theme_picker", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, `class="theme-gallery"`) {
		t.Error("Should still render theme-gallery even when empty")
	}

	if strings.Contains(html, `class="theme-card"`) {
		t.Error("Should not render any theme cards when themes list is empty")
	}
}

func stringPtr(s string) *string {
	return &s
}
