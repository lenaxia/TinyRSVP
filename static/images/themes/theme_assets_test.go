package themes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type ThemeAsset struct {
	Name         string
	HeaderImage  string
	ThumbImage   string
	ExpectedSize int64
}

var expectedThemes = []ThemeAsset{
	{
		Name:         "plain-text",
		HeaderImage:  "",
		ThumbImage:   "plain-text-thumb.svg",
		ExpectedSize: 50000,
	},
	{
		Name:         "wedding-elegance",
		HeaderImage:  "wedding-elegance-header.svg",
		ThumbImage:   "wedding-elegance-thumb.svg",
		ExpectedSize: 50000,
	},
	{
		Name:         "birthday-celebration",
		HeaderImage:  "birthday-celebration-header.svg",
		ThumbImage:   "birthday-celebration-thumb.svg",
		ExpectedSize: 50000,
	},
	{
		Name:         "corporate-professional",
		HeaderImage:  "corporate-professional-header.svg",
		ThumbImage:   "corporate-professional-thumb.svg",
		ExpectedSize: 50000,
	},
	{
		Name:         "holiday-festive",
		HeaderImage:  "holiday-festive-header.svg",
		ThumbImage:   "holiday-festive-thumb.svg",
		ExpectedSize: 50000,
	},
	{
		Name:         "garden-party",
		HeaderImage:  "garden-party-header.svg",
		ThumbImage:   "garden-party-thumb.svg",
		ExpectedSize: 50000,
	},
	{
		Name:         "modern-minimalist",
		HeaderImage:  "modern-minimalist-header.svg",
		ThumbImage:   "modern-minimalist-thumb.svg",
		ExpectedSize: 50000,
	},
}

func TestThemeHeaderImagesExist(t *testing.T) {
	for _, theme := range expectedThemes {
		if theme.HeaderImage == "" {
			continue
		}

		t.Run(theme.Name+"_header", func(t *testing.T) {
			path := filepath.Join(".", theme.HeaderImage)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Header image not found: %s, error: %v", path, err)
			}

			if info.Size() == 0 {
				t.Errorf("Header image is empty: %s", path)
			}

			if info.Size() > theme.ExpectedSize {
				t.Errorf("Header image too large: %s, size: %d bytes, expected max: %d bytes",
					path, info.Size(), theme.ExpectedSize)
			}
		})
	}
}

func TestThemeThumbnailImagesExist(t *testing.T) {
	for _, theme := range expectedThemes {
		t.Run(theme.Name+"_thumb", func(t *testing.T) {
			path := filepath.Join(".", theme.ThumbImage)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Thumbnail image not found: %s, error: %v", path, err)
			}

			if info.Size() == 0 {
				t.Errorf("Thumbnail image is empty: %s", path)
			}

			maxThumbSize := int64(30000)
			if info.Size() > maxThumbSize {
				t.Errorf("Thumbnail image too large: %s, size: %d bytes, expected max: %d bytes",
					path, info.Size(), maxThumbSize)
			}
		})
	}
}

func TestThemeImagesAreSVG(t *testing.T) {
	for _, theme := range expectedThemes {
		if theme.HeaderImage != "" {
			t.Run(theme.Name+"_header_svg", func(t *testing.T) {
				if !strings.HasSuffix(theme.HeaderImage, ".svg") {
					t.Errorf("Header image should be SVG: %s", theme.HeaderImage)
				}

				content, err := os.ReadFile(theme.HeaderImage)
				if err != nil {
					t.Fatalf("Failed to read header image: %v", err)
				}

				if !strings.Contains(string(content), "<svg") {
					t.Errorf("Header image does not contain valid SVG content: %s", theme.HeaderImage)
				}
			})
		}

		t.Run(theme.Name+"_thumb_svg", func(t *testing.T) {
			if !strings.HasSuffix(theme.ThumbImage, ".svg") {
				t.Errorf("Thumbnail image should be SVG: %s", theme.ThumbImage)
			}

			content, err := os.ReadFile(theme.ThumbImage)
			if err != nil {
				t.Fatalf("Failed to read thumbnail image: %v", err)
			}

			if !strings.Contains(string(content), "<svg") {
				t.Errorf("Thumbnail image does not contain valid SVG content: %s", theme.ThumbImage)
			}
		})
	}
}

func TestThemeImagesHaveViewBox(t *testing.T) {
	for _, theme := range expectedThemes {
		if theme.HeaderImage != "" {
			t.Run(theme.Name+"_header_viewbox", func(t *testing.T) {
				content, err := os.ReadFile(theme.HeaderImage)
				if err != nil {
					t.Fatalf("Failed to read header image: %v", err)
				}

				if !strings.Contains(string(content), "viewBox") {
					t.Errorf("Header image should have viewBox attribute for responsiveness: %s", theme.HeaderImage)
				}
			})
		}

		t.Run(theme.Name+"_thumb_viewbox", func(t *testing.T) {
			content, err := os.ReadFile(theme.ThumbImage)
			if err != nil {
				t.Fatalf("Failed to read thumbnail image: %v", err)
			}

			if !strings.Contains(string(content), "viewBox") {
				t.Errorf("Thumbnail image should have viewBox attribute for responsiveness: %s", theme.ThumbImage)
			}
		})
	}
}

func TestThemeCount(t *testing.T) {
	expectedCount := 7
	actualCount := len(expectedThemes)

	if actualCount != expectedCount {
		t.Errorf("Expected %d themes, got %d", expectedCount, actualCount)
	}

	plainTextCount := 0
	cardBasedCount := 0

	for _, theme := range expectedThemes {
		if theme.HeaderImage == "" {
			plainTextCount++
		} else {
			cardBasedCount++
		}
	}

	if plainTextCount != 1 {
		t.Errorf("Expected 1 plain text theme, got %d", plainTextCount)
	}

	if cardBasedCount != 6 {
		t.Errorf("Expected 6 card-based themes, got %d", cardBasedCount)
	}
}
