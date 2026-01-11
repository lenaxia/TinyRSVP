package handlers

import (
	"fmt"
	"math"
	"strconv"
)

func calculateRelativeLuminance(hexColor string) float64 {
	if len(hexColor) != 7 || hexColor[0] != '#' {
		return 0
	}

	r, err := strconv.ParseInt(hexColor[1:3], 16, 64)
	if err != nil {
		return 0
	}
	g, err := strconv.ParseInt(hexColor[3:5], 16, 64)
	if err != nil {
		return 0
	}
	b, err := strconv.ParseInt(hexColor[5:7], 16, 64)
	if err != nil {
		return 0
	}

	rNorm := float64(r) / 255.0
	gNorm := float64(g) / 255.0
	bNorm := float64(b) / 255.0

	rLinear := sRGBToLinear(rNorm)
	gLinear := sRGBToLinear(gNorm)
	bLinear := sRGBToLinear(bNorm)

	return 0.2126*rLinear + 0.7152*gLinear + 0.0722*bLinear
}

func sRGBToLinear(channel float64) float64 {
	if channel <= 0.03928 {
		return channel / 12.92
	}
	return math.Pow((channel+0.055)/1.055, 2.4)
}

func calculateContrastRatio(color1, color2 string) float64 {
	l1 := calculateRelativeLuminance(color1)
	l2 := calculateRelativeLuminance(color2)

	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)

	return (lighter + 0.05) / (darker + 0.05)
}

func meetsWCAGAA(foreground, background string) bool {
	ratio := calculateContrastRatio(foreground, background)
	return ratio >= 3.0
}

func validateCustomColorContrast(customColor string) (bool, string) {
	if !isValidHexColor(customColor) {
		return false, "invalid color format"
	}

	lightBg := "#FFFFFF"
	darkBg := "#0F172A"

	lightRatio := calculateContrastRatio(customColor, lightBg)
	darkRatio := calculateContrastRatio(customColor, darkBg)

	if lightRatio >= 3.0 && darkRatio >= 3.0 {
		return true, ""
	}

	return false, "color does not meet WCAG AA contrast requirements"
}

func generateColorOverrideCSS(customColor string) string {
	if customColor == "" {
		return ""
	}

	if !isValidHexColor(customColor) {
		return ""
	}

	valid, _ := validateCustomColorContrast(customColor)
	if !valid {
		return ""
	}

	return fmt.Sprintf(`<style>
[data-event-theme] {
    --theme-primary: %s !important;
}
[data-theme="dark"][data-event-theme] {
    --theme-primary: %s !important;
}
</style>`, customColor, customColor)
}
