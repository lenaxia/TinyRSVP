package handlers

import (
	"testing"
)

func TestCalculateRelativeLuminance(t *testing.T) {
	tests := []struct {
		name     string
		hexColor string
		want     float64
	}{
		{
			name:     "white",
			hexColor: "#FFFFFF",
			want:     1.0,
		},
		{
			name:     "black",
			hexColor: "#000000",
			want:     0.0,
		},
		{
			name:     "red",
			hexColor: "#FF0000",
			want:     0.2126,
		},
		{
			name:     "green",
			hexColor: "#00FF00",
			want:     0.7152,
		},
		{
			name:     "blue",
			hexColor: "#0000FF",
			want:     0.0722,
		},
		{
			name:     "gray",
			hexColor: "#808080",
			want:     0.2159,
		},
		{
			name:     "bootstrap primary",
			hexColor: "#007BFF",
			want:     0.2139,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRelativeLuminance(tt.hexColor)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Errorf("calculateRelativeLuminance(%q) = %v, want %v (diff: %v)", tt.hexColor, got, tt.want, diff)
			}
		})
	}
}

func TestCalculateContrastRatio(t *testing.T) {
	tests := []struct {
		name   string
		color1 string
		color2 string
		want   float64
	}{
		{
			name:   "white on black",
			color1: "#FFFFFF",
			color2: "#000000",
			want:   21.0,
		},
		{
			name:   "black on white",
			color1: "#000000",
			color2: "#FFFFFF",
			want:   21.0,
		},
		{
			name:   "same color",
			color1: "#808080",
			color2: "#808080",
			want:   1.0,
		},
		{
			name:   "blue on white",
			color1: "#0000FF",
			color2: "#FFFFFF",
			want:   8.59,
		},
		{
			name:   "red on white",
			color1: "#FF0000",
			color2: "#FFFFFF",
			want:   3.998,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateContrastRatio(tt.color1, tt.color2)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.1 {
				t.Errorf("calculateContrastRatio(%q, %q) = %v, want %v (diff: %v)", tt.color1, tt.color2, got, tt.want, diff)
			}
		})
	}
}

func TestMeetsWCAGAA(t *testing.T) {
	tests := []struct {
		name       string
		foreground string
		background string
		want       bool
	}{
		{
			name:       "white on black - passes",
			foreground: "#FFFFFF",
			background: "#000000",
			want:       true,
		},
		{
			name:       "black on white - passes",
			foreground: "#000000",
			background: "#FFFFFF",
			want:       true,
		},
		{
			name:       "blue on white - passes",
			foreground: "#0000FF",
			background: "#FFFFFF",
			want:       true,
		},
		{
			name:       "light gray on white - fails",
			foreground: "#CCCCCC",
			background: "#FFFFFF",
			want:       false,
		},
		{
			name:       "yellow on white - fails",
			foreground: "#FFFF00",
			background: "#FFFFFF",
			want:       false,
		},
		{
			name:       "dark blue on white - passes",
			foreground: "#007BFF",
			background: "#FFFFFF",
			want:       true,
		},
		{
			name:       "dark blue on black - passes",
			foreground: "#007BFF",
			background: "#000000",
			want:       true,
		},
		{
			name:       "green on white - passes",
			foreground: "#008000",
			background: "#FFFFFF",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meetsWCAGAA(tt.foreground, tt.background)
			if got != tt.want {
				ratio := calculateContrastRatio(tt.foreground, tt.background)
				t.Errorf("meetsWCAGAA(%q, %q) = %v, want %v (ratio: %.2f)", tt.foreground, tt.background, got, tt.want, ratio)
			}
		})
	}
}

func TestValidateCustomColorContrast(t *testing.T) {
	tests := []struct {
		name        string
		customColor string
		wantValid   bool
		wantReason  string
	}{
		{
			name:        "valid dark color for light mode",
			customColor: "#007BFF",
			wantValid:   true,
			wantReason:  "",
		},
		{
			name:        "invalid very dark color - fails on dark bg",
			customColor: "#000080",
			wantValid:   false,
			wantReason:  "color does not meet WCAG AA contrast requirements",
		},
		{
			name:        "invalid light color - fails on dark bg",
			customColor: "#FFFF00",
			wantValid:   false,
			wantReason:  "color does not meet WCAG AA contrast requirements",
		},
		{
			name:        "invalid very light color - fails on dark bg",
			customColor: "#CCCCCC",
			wantValid:   false,
			wantReason:  "color does not meet WCAG AA contrast requirements",
		},
		{
			name:        "valid medium dark color",
			customColor: "#2563EB",
			wantValid:   true,
			wantReason:  "",
		},
		{
			name:        "valid dark green",
			customColor: "#16A34A",
			wantValid:   true,
			wantReason:  "",
		},
		{
			name:        "invalid light pink - fails on dark bg",
			customColor: "#FFB6C1",
			wantValid:   false,
			wantReason:  "color does not meet WCAG AA contrast requirements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, gotReason := validateCustomColorContrast(tt.customColor)
			if gotValid != tt.wantValid {
				ratio := calculateContrastRatio(tt.customColor, "#FFFFFF")
				t.Errorf("validateCustomColorContrast(%q) valid = %v, want %v (ratio: %.2f)", tt.customColor, gotValid, tt.wantValid, ratio)
			}
			if gotReason != tt.wantReason {
				t.Errorf("validateCustomColorContrast(%q) reason = %q, want %q", tt.customColor, gotReason, tt.wantReason)
			}
		})
	}
}

func TestGenerateColorOverrideCSS(t *testing.T) {
	tests := []struct {
		name        string
		customColor string
		want        string
	}{
		{
			name:        "valid color",
			customColor: "#007BFF",
			want: `<style>
[data-event-theme] {
    --theme-primary: #007BFF !important;
}
[data-theme="dark"][data-event-theme] {
    --theme-primary: #007BFF !important;
}
</style>`,
		},
		{
			name:        "different color",
			customColor: "#16A34A",
			want: `<style>
[data-event-theme] {
    --theme-primary: #16A34A !important;
}
[data-theme="dark"][data-event-theme] {
    --theme-primary: #16A34A !important;
}
</style>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateColorOverrideCSS(tt.customColor)
			if got != tt.want {
				t.Errorf("generateColorOverrideCSS(%q) = %q, want %q", tt.customColor, got, tt.want)
			}
		})
	}
}

func TestGenerateColorOverrideCSS_EmptyColor(t *testing.T) {
	got := generateColorOverrideCSS("")
	if got != "" {
		t.Errorf("generateColorOverrideCSS(\"\") = %q, want empty string", got)
	}
}

func TestGenerateColorOverrideCSS_InvalidColor(t *testing.T) {
	tests := []struct {
		name  string
		color string
	}{
		{"no hash", "007BFF"},
		{"too short", "#FFF"},
		{"too long", "#007BFF00"},
		{"invalid chars", "#GGGGGG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateColorOverrideCSS(tt.color)
			if got != "" {
				t.Errorf("generateColorOverrideCSS(%q) = %q, want empty string for invalid color", tt.color, got)
			}
		})
	}
}
