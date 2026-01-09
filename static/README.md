# Static Assets

This directory contains static assets served by the TinyRSVP application.

## Purpose

Houses CSS, JavaScript, and other static files that are served directly to browsers without server-side processing.

## Structure

```
static/
├── css/              # Stylesheets
│   ├── variables.css # CSS custom properties (design tokens)
│   └── README.md     # CSS documentation
├── js/               # JavaScript files (future)
└── images/           # Static images (future)
```

## Rules

1. **CSS First:** All styling should use CSS custom properties from [`css/variables.css`](css/variables.css)
2. **No Build Step:** Files served directly, no compilation/transpilation
3. **Mobile First:** All CSS should be mobile-first with progressive enhancement
4. **Vanilla JS:** No frameworks, use plain JavaScript (ES6+)
5. **Progressive Enhancement:** Pages should work without JavaScript

## CSS Variables System

The [`css/variables.css`](css/variables.css) file provides a comprehensive design token system:

- **Colors:** Primary, semantic, gray scale, and functional colors
- **Spacing:** 8px-based scale (0-96px)
- **Typography:** Font sizes, weights, line heights, and families
- **Visual Effects:** Border radius, shadows, transitions
- **Layout:** Breakpoints, container widths, z-index scale
- **Dark Mode:** Automatic support via `prefers-color-scheme`

See [`css/README.md`](css/README.md) for complete variable reference.

## Integration with Templates

### HTML Templates

Include CSS variables in your HTML templates:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Page Title</title>
    <link rel="stylesheet" href="/static/css/variables.css">
    <style>
        /* Your component styles using variables */
        .button {
            background-color: var(--color-primary-600);
            padding: var(--spacing-3) var(--spacing-6);
            border-radius: var(--radius-md);
        }
    </style>
</head>
<body>
    <!-- Content -->
</body>
</html>
```

### Go Templates

Reference static assets in Go templates:

```go
// In your handler
tmpl := template.Must(template.ParseFiles("templates/web/page.html"))
tmpl.Execute(w, data)
```

## Serving Static Files

Static files are served by the application at the `/static/` path:

- `/static/css/variables.css` → `static/css/variables.css`
- `/static/js/app.js` → `static/js/app.js` (future)
- `/static/images/logo.png` → `static/images/logo.png` (future)

## Testing

CSS files have Go tests to validate structure and completeness:

```bash
cd static/css && go test -timeout 30s -v
```

Tests validate:
- All required variables are defined
- Syntax is correct
- Dark mode support is present
- Integration with existing templates
- WCAG AA color contrast compliance

## Browser Support

All CSS features used are supported in:
- Chrome/Edge 49+
- Firefox 31+
- Safari 9.1+
- iOS Safari 9.3+
- Chrome for Android 49+

## Performance

- **CSS Variables:** ~5KB uncompressed
- **Total CSS:** Target <25KB for all styles
- **JavaScript:** Target <10KB for all scripts
- **Images:** Optimized, lazy-loaded where appropriate

## Future Additions

- `js/` - Vanilla JavaScript for progressive enhancement
- `images/` - Static images (logos, icons)
- `fonts/` - Custom fonts (if needed)

## Related Documentation

- **Epic:** [docs/00_BACKLOG/07_EPIC_frontend.md](../docs/00_BACKLOG/07_EPIC_frontend.md)
- **Story:** [docs/00_BACKLOG/07_STORY_00_css_variables.md](../docs/00_BACKLOG/07_STORY_00_css_variables.md)
- **HLD:** [docs/02_REVISED_HLD.md](../docs/02_REVISED_HLD.md) Section 22 (UI/UX)
