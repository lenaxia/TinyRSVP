package templates

import (
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestGetPreset_HeroSection(t *testing.T) {
	preset, err := GetPreset("hero-section")
	if err != nil {
		t.Fatalf("GetPreset() error = %v", err)
	}

	if preset == nil {
		t.Fatal("Expected preset, got nil")
	}

	if len(preset.Components) == 0 {
		t.Error("Expected components in preset")
	}

	hasBackground := false
	hasTitle := false
	for _, comp := range preset.Components {
		if comp.Type == models.ComponentTypeBackground {
			hasBackground = true
		}
		if comp.Type == models.ComponentTypeTextBox && comp.ID == "hero-title" {
			hasTitle = true
		}
	}

	if !hasBackground {
		t.Error("Hero section should have background component")
	}
	if !hasTitle {
		t.Error("Hero section should have title component")
	}
}

func TestGetPreset_CallToAction(t *testing.T) {
	preset, err := GetPreset("call-to-action")
	if err != nil {
		t.Fatalf("GetPreset() error = %v", err)
	}

	if preset == nil {
		t.Fatal("Expected preset, got nil")
	}

	if len(preset.Components) == 0 {
		t.Error("Expected components in preset")
	}

	hasContainer := false
	hasButton := false
	for _, comp := range preset.Components {
		if comp.Type == models.ComponentTypeContainer {
			hasContainer = true
		}
		if comp.Type == models.ComponentTypeTextBox && comp.ID == "cta-button" {
			hasButton = true
		}
	}

	if !hasContainer {
		t.Error("CTA should have container component")
	}
	if !hasButton {
		t.Error("CTA should have button component")
	}
}

func TestGetPreset_ImageGallery(t *testing.T) {
	preset, err := GetPreset("image-gallery")
	if err != nil {
		t.Fatalf("GetPreset() error = %v", err)
	}

	if preset == nil {
		t.Fatal("Expected preset, got nil")
	}

	if len(preset.Components) == 0 {
		t.Error("Expected components in preset")
	}

	hasContainer := false
	imageCount := 0
	for _, comp := range preset.Components {
		if comp.Type == models.ComponentTypeContainer {
			hasContainer = true
		}
		if comp.Type == models.ComponentTypeImage {
			imageCount++
		}
	}

	if !hasContainer {
		t.Error("Gallery should have container component")
	}
	if imageCount < 3 {
		t.Errorf("Gallery should have at least 3 images, got %d", imageCount)
	}
}

func TestGetPreset_Testimonial(t *testing.T) {
	preset, err := GetPreset("testimonial")
	if err != nil {
		t.Fatalf("GetPreset() error = %v", err)
	}

	if preset == nil {
		t.Fatal("Expected preset, got nil")
	}

	if len(preset.Components) == 0 {
		t.Error("Expected components in preset")
	}

	hasQuote := false
	hasAuthor := false
	for _, comp := range preset.Components {
		if comp.Type == models.ComponentTypeTextBox && comp.ID == "testimonial-quote" {
			hasQuote = true
		}
		if comp.Type == models.ComponentTypeTextBox && comp.ID == "testimonial-author" {
			hasAuthor = true
		}
	}

	if !hasQuote {
		t.Error("Testimonial should have quote component")
	}
	if !hasAuthor {
		t.Error("Testimonial should have author component")
	}
}

func TestGetPreset_InvalidName(t *testing.T) {
	preset, err := GetPreset("invalid-preset-name")
	if err == nil {
		t.Error("Expected error for invalid preset name")
	}
	if preset != nil {
		t.Error("Expected nil preset for invalid name")
	}
}

func TestGetPreset_EmptyName(t *testing.T) {
	preset, err := GetPreset("")
	if err == nil {
		t.Error("Expected error for empty preset name")
	}
	if preset != nil {
		t.Error("Expected nil preset for empty name")
	}
}

func TestListPresets(t *testing.T) {
	presets := ListPresets()

	if len(presets) == 0 {
		t.Error("Expected at least one preset")
	}

	expectedPresets := map[string]bool{
		"hero-section":   false,
		"call-to-action": false,
		"image-gallery":  false,
		"testimonial":    false,
	}

	for _, preset := range presets {
		if preset.Name == "" {
			t.Error("Preset name should not be empty")
		}
		if preset.Description == "" {
			t.Error("Preset description should not be empty")
		}
		if preset.Category == "" {
			t.Error("Preset category should not be empty")
		}

		if _, exists := expectedPresets[preset.Name]; exists {
			expectedPresets[preset.Name] = true
		}
	}

	for name, found := range expectedPresets {
		if !found {
			t.Errorf("Expected preset %s not found in list", name)
		}
	}
}

func TestPresetMetadata(t *testing.T) {
	presets := ListPresets()

	for _, preset := range presets {
		t.Run(preset.Name, func(t *testing.T) {
			if preset.Name == "" {
				t.Error("Name should not be empty")
			}
			if preset.Description == "" {
				t.Error("Description should not be empty")
			}
			if preset.Category == "" {
				t.Error("Category should not be empty")
			}
			if len(preset.Tags) == 0 {
				t.Error("Tags should not be empty")
			}
			if preset.ThumbnailURL == "" {
				t.Error("ThumbnailURL should not be empty")
			}
		})
	}
}

func TestGetPreset_ComponentValidity(t *testing.T) {
	presets := ListPresets()

	for _, presetMeta := range presets {
		t.Run(presetMeta.Name, func(t *testing.T) {
			preset, err := GetPreset(presetMeta.Name)
			if err != nil {
				t.Fatalf("GetPreset() error = %v", err)
			}

			if len(preset.Components) == 0 {
				t.Error("Preset should have at least one component")
			}

			for i, comp := range preset.Components {
				if comp.ID == "" {
					t.Errorf("Component %d: ID should not be empty", i)
				}
				if !comp.Type.IsValid() {
					t.Errorf("Component %d: Invalid type %s", i, comp.Type)
				}
				if !comp.Position.Mode.IsValid() {
					t.Errorf("Component %d: Invalid position mode %s", i, comp.Position.Mode)
				}
			}
		})
	}
}

func TestGetPreset_UniqueIDs(t *testing.T) {
	presets := ListPresets()

	for _, presetMeta := range presets {
		t.Run(presetMeta.Name, func(t *testing.T) {
			preset, err := GetPreset(presetMeta.Name)
			if err != nil {
				t.Fatalf("GetPreset() error = %v", err)
			}

			ids := make(map[string]bool)
			for _, comp := range preset.Components {
				if ids[comp.ID] {
					t.Errorf("Duplicate component ID: %s", comp.ID)
				}
				ids[comp.ID] = true
			}
		})
	}
}
