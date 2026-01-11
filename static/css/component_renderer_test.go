package css

import (
	"os"
	"strings"
	"testing"
)

func TestComponentRendererCSS_Exists(t *testing.T) {
	if _, err := os.Stat("component_renderer.css"); os.IsNotExist(err) {
		t.Fatal("component_renderer.css does not exist")
	}
}

func TestComponentRendererCSS_HasAnimationKeyframes(t *testing.T) {
	content, err := os.ReadFile("component_renderer.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	requiredKeyframes := []string{
		"@keyframes fade",
		"@keyframes slide",
		"@keyframes scale",
		"@keyframes rotate",
		"@keyframes bounce",
	}

	for _, keyframe := range requiredKeyframes {
		if !strings.Contains(cssContent, keyframe) {
			t.Errorf("CSS should contain %s", keyframe)
		}
	}
}

func TestComponentRendererCSS_HasGridUtilities(t *testing.T) {
	cssPath := "component_renderer.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".layout-grid",
		".grid-cols-1",
		".grid-cols-2",
		".grid-cols-3",
		".grid-gap-sm",
		".grid-gap-md",
		".grid-flow-row",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("CSS should contain %s", class)
		}
	}
}

func TestComponentRendererCSS_HasFlexUtilities(t *testing.T) {
	cssPath := "component_renderer.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".layout-flex",
		".flex-row",
		".flex-col",
		".flex-wrap",
		".justify-center",
		".items-center",
		".flex-gap-md",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("CSS should contain %s", class)
		}
	}
}

func TestComponentRendererCSS_HasImageEffects(t *testing.T) {
	cssPath := "component_renderer.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".img-filter-blur",
		".img-filter-grayscale",
		".img-blend-multiply",
		".img-clip-circle",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("CSS should contain %s", class)
		}
	}
}

func TestComponentRendererCSS_HasTextEffects(t *testing.T) {
	cssPath := "component_renderer.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".text-gradient",
		".text-stroke",
		".text-shadow-sm",
		".text-shadow-md",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("CSS should contain %s", class)
		}
	}
}

func TestComponentRendererCSS_HasVisibilityUtilities(t *testing.T) {
	cssPath := "component_renderer.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	requiredClasses := []string{
		".hide-mobile",
		".hide-tablet",
		".hide-desktop",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("CSS should contain %s", class)
		}
	}
}

func TestComponentRendererCSS_HasPerformanceOptimizations(t *testing.T) {
	cssPath := "component_renderer.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "will-change") {
		t.Error("CSS should contain will-change for performance")
	}

	if !strings.Contains(cssContent, "backface-visibility") {
		t.Error("CSS should contain backface-visibility for GPU acceleration")
	}
}

func TestComponentRendererCSS_HasAccessibilitySupport(t *testing.T) {
	cssPath := "component_renderer.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "prefers-reduced-motion") {
		t.Error("CSS should contain prefers-reduced-motion media query")
	}
}

func TestComponentRendererCSS_HasResponsiveBreakpoints(t *testing.T) {
	cssPath := "component_renderer.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	cssContent := string(content)

	breakpoints := []string{
		"@media (max-width: 767px)",
		"@media (min-width: 768px) and (max-width: 1023px)",
		"@media (min-width: 1024px)",
	}

	for _, breakpoint := range breakpoints {
		if !strings.Contains(cssContent, breakpoint) {
			t.Errorf("CSS should contain breakpoint: %s", breakpoint)
		}
	}
}
