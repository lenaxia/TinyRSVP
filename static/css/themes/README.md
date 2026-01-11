# Theme CSS Files

This directory contains CSS files for RSVP page themes.

## Structure

```
themes/
├── plain-text.css                  # Plain text theme styles
├── wedding-elegance.css            # Wedding theme styles
├── birthday-celebration.css        # Birthday theme styles
├── corporate-professional.css      # Corporate theme styles
├── holiday-festive.css             # Holiday theme styles
├── garden-party.css                # Garden theme styles
├── modern-minimalist.css           # Modern theme styles
└── theme_css_test.go               # CSS validation tests
```

## Theme CSS Structure

Each theme CSS file defines:

### Required Variables

```css
[data-event-theme="theme-name"] {
    --theme-primary: #hexcolor;      /* Primary brand color */
    --theme-secondary: #hexcolor;    /* Secondary brand color */
    --theme-accent: #hexcolor;       /* Accent color */
    --theme-font-heading: font-stack; /* Heading font */
    --theme-font-body: font-stack;    /* Body font */
}
```

### Dark Mode Overrides

```css
[data-theme="dark"][data-event-theme="theme-name"] {
    --theme-primary: #hexcolor;      /* Adjusted for dark mode */
    --theme-secondary: #hexcolor;    /* Adjusted for dark mode */
    --theme-accent: #hexcolor;       /* Adjusted for dark mode */
}
```

### Theme-Specific Styles

```css
[data-event-theme="theme-name"] .rsvp-card {
    /* Custom card styling */
}

[data-event-theme="theme-name"] .event-title {
    /* Custom title styling */
}

/* Responsive adjustments */
@media (max-width: 767px) {
    /* Mobile overrides */
}
```

## Testing

Run CSS validation tests:

```bash
cd static/css/themes
go test -timeout 30s -v
```

Tests verify:
- All CSS files exist
- Required variables defined
- Dark mode support present
- Data attribute selectors correct
- Correct theme count (7 themes)

## Adding New Theme CSS

1. Create new CSS file: `theme-name.css`
2. Define theme variables
3. Add dark mode overrides
4. Add theme-specific styles
5. Add responsive media queries
6. Update `theme_css_test.go`
7. Run tests

## Design Guidelines

See [docs/THEME_DESIGN_SYSTEM.md](../../../docs/THEME_DESIGN_SYSTEM.md) for:
- Color selection guidelines
- Typography recommendations
- Dark mode strategies
- Complete specifications
