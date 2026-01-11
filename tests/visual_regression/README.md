# Visual Regression Testing Framework

## Overview

This directory contains the visual regression testing framework for validating that migrated component-based themes render identically to the original HTML themes.

## Status

⚠️ **Not Yet Implemented** - Requires browser automation environment

## Purpose

Visual regression testing ensures that the component-based rendering system produces pixel-perfect (or visually identical) output compared to the original HTML templates.

## Requirements

### Tools Needed
- **Playwright for Go** or **Chromedp** - Browser automation
- **go-imagediff** or similar - Image comparison library
- **Running TinyRSVP server** - To render pages
- **Test database** - With seeded themes

### System Requirements
- Headless Chrome/Chromium
- Sufficient disk space for screenshots (~50MB)
- Network access to localhost

## Planned Implementation

### 1. Screenshot Capture

```go
// Capture screenshots of original HTML themes
func CaptureOriginalThemes(ctx context.Context, baseURL string) error {
    themes := []string{
        "plain-text",
        "wedding-elegance",
        "birthday-celebration",
        "corporate-professional",
        "holiday-festive",
        "garden-party",
        "modern-minimalist",
    }
    
    for _, theme := range themes {
        url := fmt.Sprintf("%s/rsvp/test-token?legacy=true&theme=%s", baseURL, theme)
        screenshot := fmt.Sprintf("screenshots/original/%s.png", theme)
        
        // Capture screenshot using browser automation
        if err := captureScreenshot(ctx, url, screenshot); err != nil {
            return err
        }
    }
    
    return nil
}

// Capture screenshots of component-based themes
func CaptureComponentThemes(ctx context.Context, baseURL string) error {
    themes := []string{
        "plain-text",
        "wedding-elegance",
        "birthday-celebration",
        "corporate-professional",
        "holiday-festive",
        "garden-party",
        "modern-minimalist",
    }
    
    for _, theme := range themes {
        url := fmt.Sprintf("%s/rsvp/test-token?component=true&theme=%s", baseURL, theme)
        screenshot := fmt.Sprintf("screenshots/component/%s.png", theme)
        
        // Capture screenshot using browser automation
        if err := captureScreenshot(ctx, url, screenshot); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 2. Image Comparison

```go
// Compare screenshots and generate diff images
func CompareScreenshots(originalPath, componentPath, diffPath string) (*ComparisonResult, error) {
    original, err := loadImage(originalPath)
    if err != nil {
        return nil, err
    }
    
    component, err := loadImage(componentPath)
    if err != nil {
        return nil, err
    }
    
    // Calculate pixel difference
    diff, similarity := compareImages(original, component)
    
    // Save diff image highlighting differences
    if err := saveDiffImage(diff, diffPath); err != nil {
        return nil, err
    }
    
    return &ComparisonResult{
        Similarity:      similarity,
        PixelDifference: diff.PixelCount,
        DiffImagePath:   diffPath,
    }, nil
}
```

### 3. Test Suite

```go
func TestVisualRegression_AllThemes(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping visual regression tests in short mode")
    }
    
    // Start test server
    server := startTestServer(t)
    defer server.Close()
    
    // Capture original theme screenshots
    if err := CaptureOriginalThemes(context.Background(), server.URL); err != nil {
        t.Fatalf("Failed to capture original themes: %v", err)
    }
    
    // Capture component-based theme screenshots
    if err := CaptureComponentThemes(context.Background(), server.URL); err != nil {
        t.Fatalf("Failed to capture component themes: %v", err)
    }
    
    // Compare screenshots
    themes := []string{
        "plain-text",
        "wedding-elegance",
        "birthday-celebration",
        "corporate-professional",
        "holiday-festive",
        "garden-party",
        "modern-minimalist",
    }
    
    for _, theme := range themes {
        t.Run(theme, func(t *testing.T) {
            original := fmt.Sprintf("screenshots/original/%s.png", theme)
            component := fmt.Sprintf("screenshots/component/%s.png", theme)
            diff := fmt.Sprintf("screenshots/diff/%s.png", theme)
            
            result, err := CompareScreenshots(original, component, diff)
            if err != nil {
                t.Fatalf("Failed to compare %s: %v", theme, err)
            }
            
            // Allow for minor differences due to anti-aliasing, etc.
            minSimilarity := 0.98 // 98% similarity threshold
            
            if result.Similarity < minSimilarity {
                t.Errorf("%s similarity %.2f%% is below threshold %.2f%%",
                    theme, result.Similarity*100, minSimilarity*100)
                t.Logf("Diff image saved to: %s", diff)
            } else {
                t.Logf("%s similarity: %.2f%%", theme, result.Similarity*100)
            }
        })
    }
}
```

## Directory Structure

```
tests/visual_regression/
├── README.md                    # This file
├── visual_regression_test.go    # Test implementation
├── screenshot_capture.go        # Browser automation
├── image_comparison.go          # Image diff logic
└── screenshots/
    ├── original/                # Original HTML theme screenshots
    │   ├── plain-text.png
    │   ├── wedding-elegance.png
    │   └── ...
    ├── component/               # Component-based theme screenshots
    │   ├── plain-text.png
    │   ├── wedding-elegance.png
    │   └── ...
    └── diff/                    # Difference images
        ├── plain-text.png
        ├── wedding-elegance.png
        └── ...
```

## Running Visual Regression Tests

```bash
# Install dependencies (when implemented)
go get github.com/chromedp/chromedp
go get github.com/oliamb/cutter

# Run visual regression tests
go test -v ./tests/visual_regression -timeout 5m

# Run with screenshot regeneration
go test -v ./tests/visual_regression -regenerate

# Run with specific theme
go test -v ./tests/visual_regression -run TestVisualRegression_AllThemes/wedding-elegance
```

## Configuration

### Viewport Sizes
- **Desktop:** 1920x1080
- **Tablet:** 768x1024
- **Mobile:** 375x667

### Comparison Thresholds
- **Similarity:** 98% minimum
- **Pixel Difference:** <2% of total pixels
- **Color Tolerance:** ±5 RGB values per channel

### Screenshot Options
- **Format:** PNG (lossless)
- **Full Page:** Yes
- **Wait for Load:** 2 seconds
- **Disable Animations:** Yes (for consistency)

## Current Validation

While visual regression testing is not yet implemented, visual fidelity is ensured through:

1. **CSS Analysis** - Manually extracted all styles from theme CSS files
2. **Component Mapping** - Careful mapping of HTML elements to components
3. **Integration Tests** - Verify HTML structure and content presence
4. **Responsive Tests** - Confirm mobile configurations exist
5. **Manual Review** - Architecture team reviewed component configs

## Future Enhancements

### Phase 1: Basic Implementation
- [ ] Set up Playwright/Chromedp
- [ ] Implement screenshot capture
- [ ] Implement basic image comparison
- [ ] Create test suite

### Phase 2: Advanced Features
- [ ] Multi-viewport testing (desktop, tablet, mobile)
- [ ] Dark mode screenshot comparison
- [ ] Animated diff visualization
- [ ] Baseline management system

### Phase 3: CI/CD Integration
- [ ] Automated screenshot capture in CI
- [ ] Baseline update workflow
- [ ] Visual diff reporting
- [ ] Slack/email notifications for failures

## Alternative Validation Methods

Until visual regression testing is implemented, use these methods:

### 1. Manual Visual Inspection
```bash
# Start server
go run cmd/server/main.go

# Visit each theme in browser
open http://localhost:8080/rsvp/test-token?theme=wedding-elegance
open http://localhost:8080/rsvp/test-token?theme=birthday-celebration
# etc.

# Compare side-by-side with original themes
```

### 2. HTML Structure Validation
```bash
# Run integration tests
go test -v ./internal/templates -run TestMigratedThemes

# Verify component structure
go test -v ./internal/templates -run TestMigratedThemes_ComponentStructure

# Check rendering output
go test -v ./internal/templates -run TestMigratedThemes_RenderCorrectly
```

### 3. CSS Inspection
- Review generated component configs against original CSS
- Verify font families, sizes, colors match
- Confirm responsive breakpoints preserved
- Check z-index layering

## Notes

- Visual regression testing is valuable but not critical for Phase 2 completion
- The comprehensive integration test suite provides high confidence
- Manual visual inspection can be performed when needed
- Future phases can add automated visual testing

---

**Status:** Framework designed, implementation deferred  
**Priority:** Medium (nice-to-have, not blocking)  
**Effort:** ~2-3 days to implement fully
