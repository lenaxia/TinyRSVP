# User Story 06.11: Component-Based Rendering

**Epic:** 06 (Templates)  
**Priority:** CRITICAL  
**Status:** ⚠️ BROKEN  
**Estimated Effort:** 2-3 days  
**Created:** 2026-02-04  

---

## User Story

As an **event manager**,  
I want **component-based RSVP page customization**,  
So that **I can create Evite-style themed invitations**.

---

## Context

Component system was implemented but broke during refactoring. Template field access doesn't match struct design. This is blocking Epic 11 (RSVP Themes) and preventing the core customization feature from working.

**Root Cause:** Struct refactoring changed the ComponentContent structure, but template files were not updated to match the new field paths.

**Impact:** All theme rendering is broken, 18 tests failing, cannot customize RSVP pages.

---

## Current Status

- **Designed:** ✅ Complete
- **Implemented:** ⚠️ Partial
- **Working:** ❌ NO
- **Test Pass Rate:** 15%

**Problem:**
```
Template field access: .Content.textAlign
Actual struct path:   .Content.TextBox.TextAlign

ERROR: can't evaluate field textAlign in type *models.ComponentContent
```

---

## Acceptance Criteria

- [ ] Component templates match struct design
- [ ] All 6 component types render correctly:
  - [ ] TextBox
  - [ ] Image
  - [ ] Background
  - [ ] Overlay
  - [ ] Container
  - [ ] Divider
- [ ] Component tests passing (18 tests)
- [ ] XSS protection working
- [ ] Theme rendering working (all 7 themes)
- [ ] Template validation catches errors
- [ ] CSS sanitization applied

---

## Component Types

### 1. TextBox
**Fields:** Text, Color, FontSize, FontFamily, TextAlign, TextShadow, FontWeight, TextDecoration, LineHeight, LetterSpacing

**Template Path:** `templates/web/partials/component_textbox.html`

### 2. Image
**Fields:** URL, Alt, Width, Height, Filter, Transform, ObjectFit, BorderRadius

**Template Path:** `templates/web/partials/component_image.html`

### 3. Background
**Fields:** Color, GradientStart, GradientEnd, GradientDirection, ImageURL, BackgroundSize, BackgroundPosition, BackgroundRepeat

**Template Path:** `templates/web/partials/component_background.html`

### 4. Overlay
**Fields:** BackgroundColor, BorderColor, BorderWidth, BorderRadius, BoxShadow, Padding

**Template Path:** `templates/web/partials/component_overlay.html`

### 5. Container
**Fields:** Display, FlexDirection, JustifyContent, AlignItems, Gap, GridTemplateColumns, GridTemplateRows

**Template Path:** `templates/web/partials/component_container.html`

### 6. Divider
**Fields:** Style, Color, Width, Height, Margin

**Template Path:** `templates/web/partials/component_divider.html`

---

## Technical Details

### Data Structure

```go
type ComponentContent struct {
    TextBox    *TextBoxContent    `json:"textBox,omitempty"`
    Image      *ImageContent      `json:"image,omitempty"`
    Background *BackgroundContent `json:"background,omitempty"`
    Overlay    *OverlayContent    `json:"overlay,omitempty"`
    Container  *ContainerContent  `json:"container,omitempty"`
    Divider    *DividerContent    `json:"divider,omitempty"`
}
```

### Template Access Pattern

**Correct:**
```html
{{with .Content.TextBox}}
    <div style="color: {{.Color}}; font-size: {{.FontSize}};">
        {{.Text}}
    </div>
{{end}}
```

**Incorrect (Current):**
```html
{{with .Content}}
    <div style="color: {{.textAlign}};">  <!-- BROKEN -->
        {{.text}}
    </div>
{{end}}
```

---

## Fix Plan

See: [`docs/00_BACKLOG/11_RSVP_THEMES/11_FIX_PLAN_component_system.md`](../11_RSVP_THEMES/11_FIX_PLAN_component_system.md)

### High-Level Steps:

1. **Audit all component templates** (6 files)
2. **Fix field access paths** to match ComponentContent struct
3. **Update CSS sanitization** to handle component styles
4. **Fix XSS tests** for component rendering
5. **Verify all 7 themes render** correctly
6. **Run full test suite** and ensure 100% pass rate

---

## Dependencies

**Blocks:**
- Epic 11 (RSVP Themes) - Cannot render themes without components
- Story 11.05 (Theme Rendering Engine) - Depends on component rendering
- Story 11.07 (Theme Integration Testing) - All tests failing due to broken components

**Blocked By:**
- None (can be fixed immediately)

---

## Testing Requirements

### Unit Tests
- [ ] TextBox rendering
- [ ] Image rendering with filters
- [ ] Background rendering with gradients
- [ ] Overlay rendering with shadows
- [ ] Container flex/grid layouts
- [ ] Divider styling
- [ ] XSS protection for all fields
- [ ] CSS sanitization validation

### Integration Tests
- [ ] Full theme rendering (7 themes)
- [ ] Component with overrides
- [ ] Multiple components on page
- [ ] Nested container layouts
- [ ] Animation rendering
- [ ] Responsive behavior

---

## Security Considerations

- All HTML auto-escaped via html/template
- CSS sanitized via whitelist
- No arbitrary JavaScript execution
- Component overrides validated
- Template injection prevention
- XSS tests must pass

---

## Definition of Done

- [ ] All component templates updated
- [ ] All 18 component tests passing
- [ ] All 7 themes render without errors
- [ ] XSS protection validated
- [ ] CSS sanitization working
- [ ] No template execution errors
- [ ] Epic 06 test pass rate > 90%
- [ ] Story marked complete with confidence level
- [ ] Documentation updated

---

## Effort Breakdown

| Task | Estimated Time |
|------|----------------|
| Audit component templates | 2 hours |
| Fix TextBox template | 2 hours |
| Fix Image template | 2 hours |
| Fix Background template | 2 hours |
| Fix Overlay template | 2 hours |
| Fix Container template | 2 hours |
| Fix Divider template | 2 hours |
| Update CSS sanitization | 3 hours |
| Fix XSS tests | 2 hours |
| Test all 7 themes | 2 hours |
| Integration testing | 2 hours |
| Documentation | 1 hour |
| **TOTAL** | **24 hours (3 days)** |

---

## Notes

- This story should have been created during Epic 06 implementation
- The component system is well-designed but implementation broke during refactoring
- Fix is straightforward: update templates to match struct paths
- High priority because it blocks all customization features
- Must be fixed before Epic 11 can be completed

---

## Related Documents

- [HLD Section 11: Templates & Customization](../../02_DESIGN/02_REVISED_HLD.md#11-templates--customization)
- [Template LLD](../../02_DESIGN/lld/06_TEMPLATE_LLD.md)
- [Component System Status Report](./COMPONENT_SYSTEM_STATUS.md)
- [Fix Plan: Component System](../11_RSVP_THEMES/11_FIX_PLAN_component_system.md)
- [Epic 11 README](../11_RSVP_THEMES/README.md)

---

**Status:** ⚠️ BROKEN  
**Next Action:** Begin template audit and fix field access paths  
**Target Completion:** 2026-02-07 (3 days)
