# User Story: Typography System

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 0.5 days

---

## User Story

As a **developer**, I want **a consistent typography system** so that **text is readable, accessible, and maintains visual hierarchy across all devices**.

---

## Acceptance Criteria

- [ ] Font family stack defined
- [ ] Type scale implemented (6 heading levels + body)
- [ ] Line height system
- [ ] Font weight system
- [ ] Letter spacing values
- [ ] Responsive typography (fluid scaling)
- [ ] Text color utilities
- [ ] Text alignment utilities
- [ ] Readability optimized (45-75 characters per line)
- [ ] All typography tested on mobile and desktop

---

## Technical Details

### Typography CSS

```css
/* static/css/typography.css */

/* Base Typography */
body {
    font-family: var(--font-family-sans);
    font-size: var(--font-size-base);
    line-height: var(--line-height-normal);
    color: var(--color-text-primary);
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
}

/* Headings */
h1, .h1 {
    font-size: var(--font-size-4xl);
    font-weight: var(--font-weight-bold);
    line-height: var(--line-height-tight);
    margin-bottom: var(--spacing-4);
}

h2, .h2 {
    font-size: var(--font-size-3xl);
    font-weight: var(--font-weight-bold);
    line-height: var(--line-height-tight);
    margin-bottom: var(--spacing-3);
}

h3, .h3 {
    font-size: var(--font-size-2xl);
    font-weight: var(--font-weight-semibold);
    line-height: var(--line-height-tight);
    margin-bottom: var(--spacing-3);
}

h4, .h4 {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-semibold);
    line-height: var(--line-height-normal);
    margin-bottom: var(--spacing-2);
}

h5, .h5 {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-medium);
    line-height: var(--line-height-normal);
    margin-bottom: var(--spacing-2);
}

h6, .h6 {
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-medium);
    line-height: var(--line-height-normal);
    margin-bottom: var(--spacing-2);
}

/* Responsive Headings */
@media (min-width: 768px) {
    h1, .h1 { font-size: var(--font-size-5xl); }
    h2, .h2 { font-size: var(--font-size-4xl); }
}

/* Body Text */
p {
    margin-bottom: var(--spacing-4);
    max-width: 65ch; /* Optimal reading width */
}

.text-large {
    font-size: var(--font-size-lg);
}

.text-small {
    font-size: var(--font-size-sm);
}

.text-xs {
    font-size: var(--font-size-xs);
}

/* Text Utilities */
.text-bold { font-weight: var(--font-weight-bold); }
.text-semibold { font-weight: var(--font-weight-semibold); }
.text-medium { font-weight: var(--font-weight-medium); }
.text-normal { font-weight: var(--font-weight-normal); }

.text-primary { color: var(--color-text-primary); }
.text-secondary { color: var(--color-text-secondary); }
.text-disabled { color: var(--color-text-disabled); }
.text-success { color: var(--color-success); }
.text-error { color: var(--color-error); }
.text-warning { color: var(--color-warning); }

.text-left { text-align: left; }
.text-center { text-align: center; }
.text-right { text-align: right; }

/* Links */
a {
    color: var(--color-primary-600);
    text-decoration: underline;
    transition: color var(--transition-fast);
}

a:hover {
    color: var(--color-primary-700);
}

a:focus {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
}

/* Lists */
ul, ol {
    margin-bottom: var(--spacing-4);
    padding-left: var(--spacing-6);
}

li {
    margin-bottom: var(--spacing-2);
}

/* Code */
code {
    font-family: var(--font-family-mono);
    font-size: 0.875em;
    background-color: var(--color-gray-100);
    padding: 0.125rem 0.25rem;
    border-radius: var(--radius-sm);
}

pre {
    font-family: var(--font-family-mono);
    font-size: var(--font-size-sm);
    background-color: var(--color-gray-100);
    padding: var(--spacing-4);
    border-radius: var(--radius-md);
    overflow-x: auto;
    margin-bottom: var(--spacing-4);
}
```

---

## Tasks

### Phase 1: Base Typography
- [ ] Define font family stacks
- [ ] Set base font size (16px)
- [ ] Configure font smoothing
- [ ] Test font rendering across browsers

### Phase 2: Heading System
- [ ] Implement h1-h6 styles
- [ ] Add responsive heading sizes
- [ ] Test heading hierarchy
- [ ] Verify heading contrast

### Phase 3: Body Text
- [ ] Style paragraphs with optimal line length
- [ ] Add text size utilities
- [ ] Test readability on mobile

### Phase 4: Text Utilities
- [ ] Create font weight utilities
- [ ] Create color utilities
- [ ] Create alignment utilities
- [ ] Document utility classes

### Phase 5: Links & Interactive Text
- [ ] Style links with hover/focus states
- [ ] Ensure focus indicators are visible
- [ ] Test keyboard navigation

### Phase 6: Lists & Code
- [ ] Style ordered and unordered lists
- [ ] Style inline code
- [ ] Style code blocks
- [ ] Test code readability

---

## Dependencies

**Depends on:**
- 07_STORY_00_css_variables.md

**Blocks:**
- All UI component stories

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Typography system implemented
- [ ] Responsive scaling working
- [ ] Readability optimized
- [ ] Accessibility verified (contrast, focus)
- [ ] Documentation complete
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
- **WCAG:** Text contrast requirements (4.5:1 for normal text)
- **Readability:** 45-75 characters per line optimal
