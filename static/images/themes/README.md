# Theme Assets

This directory contains visual assets for RSVP page themes.

## Structure

```
themes/
├── plain-text-thumb.svg                    # Plain text theme thumbnail
├── wedding-elegance-header.svg             # Wedding theme header (1200x400px)
├── wedding-elegance-thumb.svg              # Wedding theme thumbnail (300x200px)
├── birthday-celebration-header.svg         # Birthday theme header
├── birthday-celebration-thumb.svg          # Birthday theme thumbnail
├── corporate-professional-header.svg       # Corporate theme header
├── corporate-professional-thumb.svg        # Corporate theme thumbnail
├── holiday-festive-header.svg              # Holiday theme header
├── holiday-festive-thumb.svg               # Holiday theme thumbnail
├── garden-party-header.svg                 # Garden theme header
├── garden-party-thumb.svg                  # Garden theme thumbnail
├── modern-minimalist-header.svg            # Modern theme header
├── modern-minimalist-thumb.svg             # Modern theme thumbnail
└── theme_assets_test.go                    # Asset validation tests
```

## Image Specifications

### Header Images

- **Dimensions:** 1200x400px (3:1 aspect ratio)
- **Format:** SVG (scalable vector graphics)
- **Max Size:** 50KB per file
- **Usage:** Displayed at top of RSVP card

**Requirements:**
- Must include `viewBox="0 0 1200 400"`
- Should work in light and dark modes
- No text in images (accessibility)
- Theme-appropriate design

### Thumbnail Images

- **Dimensions:** 300x200px (3:2 aspect ratio)
- **Format:** SVG (scalable vector graphics)
- **Max Size:** 30KB per file
- **Usage:** Theme picker gallery

**Requirements:**
- Must include `viewBox="0 0 300 200"`
- Should represent theme visually
- Include theme name in design

## Testing

Run asset validation tests:

```bash
cd static/images/themes
go test -timeout 30s -v
```

Tests verify:
- All required images exist
- Images are valid SVG files
- Images have viewBox attribute
- File sizes within limits
- Correct theme count (7 themes)

## Adding New Theme Images

1. Create header SVG (1200x400px viewBox)
2. Create thumbnail SVG (300x200px viewBox)
3. Use theme colors in design
4. Keep file sizes small (<50KB header, <30KB thumb)
5. Update `theme_assets_test.go` with new theme
6. Run tests to verify

## Design Guidelines

See [docs/THEME_DESIGN_SYSTEM.md](../../../docs/THEME_DESIGN_SYSTEM.md) for:
- Color selection guidelines
- Typography recommendations
- Dark mode strategies
- Accessibility requirements
- Complete theme specifications
