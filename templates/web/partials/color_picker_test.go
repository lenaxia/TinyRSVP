package web

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestColorPickerPartial(t *testing.T) {
	tests := []struct {
		name           string
		customColor    string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:        "renders color picker with no custom color",
			customColor: "",
			wantContains: []string{
				`class="color-picker-section"`,
				`for="custom-theme-color"`,
				`Custom Theme Color`,
				`type="color"`,
				`id="custom-theme-color"`,
				`name="custom_theme_color"`,
				`class="color-input"`,
				`type="text"`,
				`id="custom-theme-color-hex"`,
				`class="color-hex-input"`,
				`pattern="^#[0-9A-Fa-f]{6}$"`,
				`maxlength="7"`,
				`type="button"`,
				`btn-reset-color`,
				`Reset to Theme Default`,
				`aria-label="Reset color to theme default"`,
			},
		},
		{
			name:        "renders color picker with custom color",
			customColor: "#FF5733",
			wantContains: []string{
				`value="#FF5733"`,
				`class="color-preview"`,
				`aria-label="Current color preview"`,
			},
		},
		{
			name:        "renders with lowercase hex color",
			customColor: "#ff5733",
			wantContains: []string{
				`value="#ff5733"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("color_picker").ParseFiles("color_picker.html")
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			data := struct {
				Event struct {
					CustomThemeColor *string
				}
			}{}

			if tt.customColor != "" {
				data.Event.CustomThemeColor = &tt.customColor
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", want, output)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(output, notWant) {
					t.Errorf("Expected output NOT to contain %q, but it did.\nOutput:\n%s", notWant, output)
				}
			}
		})
	}
}

func TestColorPickerAccessibility(t *testing.T) {
	tests := []struct {
		name         string
		customColor  string
		wantContains []string
	}{
		{
			name:        "has proper ARIA labels",
			customColor: "#FF5733",
			wantContains: []string{
				`aria-label="Select custom theme color"`,
				`aria-describedby="color-help"`,
				`aria-label="Enter hex color code"`,
				`aria-label="Current color preview"`,
				`aria-label="Reset color to theme default"`,
			},
		},
		{
			name:        "has help text with proper ID",
			customColor: "",
			wantContains: []string{
				`id="color-help"`,
				`class="form-help-text"`,
			},
		},
		{
			name:        "color input has proper attributes",
			customColor: "#123456",
			wantContains: []string{
				`type="color"`,
				`required`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("color_picker").ParseFiles("color_picker.html")
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			data := struct {
				Event struct {
					CustomThemeColor *string
				}
			}{}

			if tt.customColor != "" {
				data.Event.CustomThemeColor = &tt.customColor
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q for accessibility, but it didn't", want)
				}
			}
		})
	}
}

func TestColorPickerHexValidation(t *testing.T) {
	tmpl, err := template.New("color_picker").ParseFiles("color_picker.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := struct {
		Event struct {
			CustomThemeColor *string
		}
	}{}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	requiredPatterns := []string{
		`pattern="^#[0-9A-Fa-f]{6}$"`,
		`maxlength="7"`,
		`placeholder="#007bff"`,
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected hex input to have validation pattern %q", pattern)
		}
	}
}

func TestColorPickerHiddenInput(t *testing.T) {
	tmpl, err := template.New("color_picker").ParseFiles("color_picker.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	customColor := "#FF5733"
	data := struct {
		Event struct {
			CustomThemeColor *string
		}
	}{}
	data.Event.CustomThemeColor = &customColor

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, `type="hidden"`) {
		t.Error("Expected hidden input for form submission")
	}

	if !strings.Contains(output, `name="custom_theme_color"`) {
		t.Error("Expected hidden input to have name='custom_theme_color'")
	}

	if !strings.Contains(output, `value="#FF5733"`) {
		t.Error("Expected hidden input to have the custom color value")
	}
}
