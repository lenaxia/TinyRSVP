# Event Form Layout Fix Plan

## Current Problems

### Problem 1: Desktop Layout is Backwards
**Current State:**
- Theme column: `order: 2` (right side) + 75% width
- Details column: `order: 1` (left side) + 25% width

**Expected:**
- Theme column should be LEFT (75% width)
- Details column should be RIGHT (25% width)

**Root Cause:**
The `order` values are swapped relative to the width percentages.

### Problem 2: Mobile Theme Grid Shows 1 Column
**Current State:**
- Base: `grid-template-columns: 1fr` (1 column)
- Media query 480-767px: `grid-template-columns: repeat(2, 1fr)` (2 columns)

**Expected:**
- Mobile <480px: 1 column
- Mobile 480-1023px: 2 columns  
- Desktop ≥1024px: responsive grid

**Possible Root Causes:**
1. Media query not applying due to parent container constraints
2. More specific selector overriding the rule
3. Container width is less than 480px even on larger screens

## The Fix

### Fix 1: event_form.css Desktop Section (lines 131-166)

**Change:**
```css
@media (min-width: 1024px) {
    .event-form-theme-column {
        order: 1;  /* LEFT side - CHANGED from 2 */
        flex: 0 0 calc(75% - var(--spacing-3));  /* 75% width */
    }

    .event-form-details-column {
        order: 2;  /* RIGHT side - CHANGED from 1 */
        flex: 0 0 calc(25% - var(--spacing-3));  /* 25% width */
    }
}
```

### Fix 2: theme_picker.css Mobile Grid

**Option A - If container width is the issue:**
Ensure parent `.event-form-theme-column` doesn't constrain width on mobile.

**Option B - If specificity is the issue:**
Check for more specific selectors overriding `.theme-gallery`.

**Option C - Consolidate media queries:**
```css
/* Mobile: 1 column by default */
.theme-gallery {
    grid-template-columns: 1fr;
}

/* Mobile ≥480px: 2 columns */
@media (min-width: 480px) {
    .theme-gallery {
        grid-template-columns: repeat(2, 1fr);
    }
}

/* Desktop ≥1024px: responsive multi-column */
@media (min-width: 1024px) {
    .theme-gallery {
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    }
}
```

## Testing Plan

1. Test desktop ≥1024px:
   - Theme should be on LEFT taking 75% width
   - Details should be on RIGHT taking 25% width
   - Theme grid should show multiple columns

2. Test mobile 480-1023px:
   - Details should be on TOP
   - Theme should be BELOW
   - Theme grid should show 2 columns per row

3. Test mobile <480px:
   - Details should be on TOP
   - Theme should be BELOW
   - Theme grid should show 1 column per row

## Confidence Level

**Desktop Fix**: 100% confident - the order values are clearly swapped
**Mobile Fix**: 80% confident - need to verify no parent constraints or specificity issues
