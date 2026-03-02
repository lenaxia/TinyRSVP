# Epic 11 Component System Fix Plan

**Date:** 2026-02-03  
**Epic:** 11 (RSVP Themes) + Epic 06 (Templates)  
**Priority:** CRITICAL  
**Status:** Planning  
**Estimated Effort:** 2-3 days  

---

## Problem Statement

The component-based customization system is completely broken. All theme rendering tests are failing because of a mismatch between the ComponentContent struct design and how templates access fields.

**Root Cause:**
```go
// Current struct (internal/models/component.go:144-149)
type ComponentContent struct {
    TextBox    *TextBoxContent    `json:"-"`
    Image      *ImageContent      `json:"-"`
    Background *BackgroundContent `json:"-"`
    Overlay    *OverlayContent    `json:"-"`
}

// Template expects (component_textbox.html:19)
{{if .Content.textAlign}}

// Actual required path:
{{if .Content.TextBox.TextAlign}}
```

**Impact:**
- 0% of theme tests passing
- 18 component rendering tests failing
- All 7 themes broken
- Evite-style customization doesn't work
- Major MVP feature missing

---

## Where It Fits in Epic Structure

### Epic 06: Templates & Asset Management
**Status:** ⚠️ BROKEN (was marked Complete, but broken)  
**Confidence:** LOW (30%)  

**Stories:**
- Story 06.03: Default Templates ✅ (seeding works)
- Story 06.08: Storage Provider ✅ (upload works)
- Story 06.11: Component Rendering ❌ (BROKEN)

### Epic 11: RSVP Page Theme System  
**Status:** ⚠️ BROKEN (was marked Complete, but broken)  
**Confidence:** LOW (25%)  

**Stories:**
- Story 11.01: Theme Model Extension ✅ (database works)
- Story 11.02: Theme Asset Creation ✅ (assets exist)
- Story 11.03: Theme Picker UI ✅ (UI exists)
- Story 11.04: Theme Preview Modal ⚠️ (JS issues)
- Story 11.05: Theme Rendering Engine ❌ (BROKEN)
- Story 11.06: Theme Seeding System ✅ (seeding works)
- Story 11.07: Theme Integration Testing ❌ (all failing)

**This fix completes:**
- Story 11.05: Theme Rendering Engine
- Story 11.07: Theme Integration Testing

---

## Fix Options Analysis

### Option 1: Update Templates (RECOMMENDED)
**Effort:** 2 days  
**Risk:** LOW  
**Maintains:** Current struct design  

**Changes Required:**
- Update 6 component templates
- Update component renderer logic
- No data model changes
- No migration needed
- No storage format changes

**Pros:**
- Clean separation of concerns
- Type-safe field access
- Easier to extend
- No breaking changes to database

**Cons:**
- More verbose templates
- Harder to write by hand

### Option 2: Flatten ComponentContent
**Effort:** 3 days  
**Risk:** MEDIUM  
**Changes:** Data model  

**Changes Required:**
- Redesign ComponentContent struct
- Update all component code
- Update JSON marshaling/unmarshaling
- Migrate existing component configs
- Update templates
- Update all tests

**Pros:**
- Simpler template syntax
- Easier manual editing

**Cons:**
- Breaking change
- Requires data migration
- Loses type safety
- More error-prone

### Option 3: Template Helper Functions
**Effort:** 2.5 days  
**Risk:** MEDIUM  
**Hybrid approach**

**Changes Required:**
- Add template helper functions
- Expose flattened fields via helpers
- Update templates to use helpers
- No struct changes

**Pros:**
- Clean template syntax
- Keeps type safety
- No data changes

**Cons:**
- More template logic
- Harder to debug
- Performance overhead

---

## Recommended Solution: Option 1 (Update Templates)

**Rationale:**
1. Fastest implementation (2 days vs 3 days)
2. Lowest risk (no data model changes)
3. Maintains type safety
4. No breaking changes
5. Follows Go best practices
6. Already have test infrastructure

---

## Implementation Plan

### Phase 1: Analysis (4 hours)
- [x] Understand current struct design ✅
- [ ] Document all field access patterns
- [ ] Create field mapping table
- [ ] Identify all affected files
- [ ] Review test expectations

### Phase 2: Template Updates (8 hours)
- [ ] Update `component_textbox.html`
- [ ] Update `component_image.html`
- [ ] Update `component_background.html`
- [ ] Update `component_overlay.html`
- [ ] Update `component_container.html`
- [ ] Update `component_divider.html`
- [ ] Test each template individually

### Phase 3: Renderer Updates (4 hours)
- [ ] Update `component_renderer.go`
- [ ] Update `component_renderer_advanced.go`
- [ ] Fix field access in rendering logic
- [ ] Update template data preparation
- [ ] Add nil checks for safety

### Phase 4: Testing (4 hours)
- [ ] Fix component renderer tests
- [ ] Fix XSS protection tests
- [ ] Fix theme rendering tests
- [ ] Test all 7 themes
- [ ] Verify with/without overrides
- [ ] Integration testing

### Phase 5: Validation (2 hours)
- [ ] Manual testing of each theme
- [ ] Test custom overrides
- [ ] Test edge cases
- [ ] Performance validation
- [ ] Documentation update

**Total Estimated Time:** 22 hours (2.75 days)

---

## Field Mapping Reference

### TextBoxContent
```go
// Struct (component.go:66-80)
type TextBoxContent struct {
    Text            string
    FontFamily      string
    FontSize        string
    FontWeight      string
    Color           string
    TextAlign       string
    LineHeight      string
    LetterSpacing   string
    TextTransform   string
    Padding         string
    TextShadow      string
    EvaluatedText   string
    IsEvaluated     bool
}

// Template Access Pattern:
// Old: .Content.textAlign
// New: .Content.TextBox.TextAlign
```

### ImageContent
```go
// Struct (component.go:82-89)
type ImageContent struct {
    Src            string
    Alt            string
    ObjectFit      string
    ObjectPosition string
    Opacity        *float64
    Filter         string
}

// Template Access Pattern:
// Old: .Content.src
// New: .Content.Image.Src
```

### BackgroundContent
```go
// Struct (component.go:91-98)
type BackgroundContent struct {
    Type               string
    Color              string
    Gradient           string
    Image              string
    BackgroundSize     string
    BackgroundPosition string
}

// Template Access Pattern:
// Old: .Content.type
// New: .Content.Background.Type
```

### OverlayContent
```go
// Struct (component.go:100-109)
type OverlayContent struct {
    BackgroundColor string
    BackgroundImage string
    BackgroundSize  string
    BorderRadius    string
    Border          string
    BoxShadow       string
    ClipPath        string
    Placeholder     bool
}

// Template Access Pattern:
// Old: .Content.backgroundColor
// New: .Content.Overlay.BackgroundColor
```

---

## File Changes Required

### Templates (6 files)
1. `templates/web/partials/component_textbox.html`
   - Lines: 19-36 (field access)
   - Lines: 48 (text/evaluatedText)

2. `templates/web/partials/component_image.html`
   - All .Content.* references

3. `templates/web/partials/component_background.html`
   - Lines: 12-20 (type, color, gradient, image)

4. `templates/web/partials/component_overlay.html`
   - All .Content.* references

5. `templates/web/partials/component_container.html`
   - Layout field access

6. `templates/web/partials/component_divider.html`
   - Style field access

### Go Files (2 files)
1. `internal/templates/component_renderer.go`
   - Template data preparation
   - Field access in rendering logic

2. `internal/templates/component_renderer_advanced.go`
   - Advanced rendering logic
   - Override application

### Test Files (5+ files)
1. `internal/templates/component_renderer_test.go`
2. `internal/templates/component_xss_test.go`
3. `internal/templates/component_integration_test.go`
4. `internal/templates/migrated_themes_integration_test.go`
5. `internal/templates/service_integration_test.go`

---

## Example Template Fix

### Before (BROKEN)
```html
{{define "component_textbox"}}
<div id="{{.ID}}" style="
    {{if .Content.textAlign}}text-align: {{.Content.textAlign}};{{end}}
    {{if .Content.fontFamily}}font-family: {{.Content.fontFamily}};{{end}}
    {{if .Content.fontSize}}font-size: {{.Content.fontSize}};{{end}}
">
    {{if .Content.isEvaluated}}{{.Content.evaluatedText}}{{else}}{{.Content.text}}{{end}}
</div>
{{end}}
```

### After (WORKING)
```html
{{define "component_textbox"}}
{{with .Content.TextBox}}
<div id="{{$.ID}}" style="
    {{if .TextAlign}}text-align: {{.TextAlign}};{{end}}
    {{if .FontFamily}}font-family: {{.FontFamily}};{{end}}
    {{if .FontSize}}font-size: {{.FontSize}};{{end}}
">
    {{if .IsEvaluated}}{{.EvaluatedText}}{{else}}{{.Text}}{{end}}
</div>
{{end}}
{{end}}
```

**Key Changes:**
1. Use `{{with .Content.TextBox}}` to scope
2. Access fields directly: `.TextAlign` not `.Content.textAlign`
3. Use `$.ID` for parent context (component ID)
4. Capitalize field names (Go exported fields)

---

## Test Strategy

### Unit Tests
```go
func TestComponentTextBox_FieldAccess(t *testing.T) {
    component := &models.Component{
        ID:   "text1",
        Type: models.ComponentTypeTextBox,
        Content: &models.ComponentContent{
            TextBox: &models.TextBoxContent{
                Text:      "Hello World",
                TextAlign: "center",
                FontSize:  "24px",
            },
        },
    }
    
    html, err := renderer.Render(component, nil)
    require.NoError(t, err)
    
    assert.Contains(t, html, "text-align: center")
    assert.Contains(t, html, "font-size: 24px")
    assert.Contains(t, html, "Hello World")
}
```

### Integration Tests
```go
func TestThemeRendering_AllThemes(t *testing.T) {
    themes := []string{
        "plain_text",
        "wedding_elegance",
        "birthday_celebration",
        "corporate_professional",
        "holiday_festive",
        "garden_party",
        "modern_minimalist",
    }
    
    for _, themeName := range themes {
        t.Run(themeName, func(t *testing.T) {
            theme := loadTheme(t, themeName)
            html, err := service.RenderRSVPPage(ctx, event, theme)
            require.NoError(t, err)
            assert.NotEmpty(t, html)
        })
    }
}
```

---

## Success Criteria

### Code Quality
- [ ] All component tests passing
- [ ] All theme tests passing
- [ ] All XSS tests passing
- [ ] No regressions in other tests
- [ ] Code review approved

### Functionality
- [ ] All 7 themes render correctly
- [ ] Component customization works
- [ ] Overrides apply correctly
- [ ] No visual regressions
- [ ] Performance acceptable (<2s page load)

### Testing
- [ ] Unit tests pass (component rendering)
- [ ] Integration tests pass (theme rendering)
- [ ] XSS tests pass (security)
- [ ] Manual testing complete (all themes)
- [ ] Edge cases tested (nil values, missing fields)

---

## Rollout Plan

### Step 1: Fix & Test (Day 1-2)
1. Fix all templates
2. Fix renderer
3. Get tests passing
4. Manual validation

### Step 2: Integration (Day 2-3)
1. Test with actual events
2. Test with overrides
3. Test all themes
4. Performance testing

### Step 3: Deployment (Day 3)
1. Commit changes
2. Update documentation
3. Mark stories complete
4. Update project status

---

## Risk Mitigation

### Risk: More Issues Discovered
**Mitigation:** Fix as discovered, buffer time allocated

### Risk: Template Syntax Complexity
**Mitigation:** Use `{{with}}` blocks, add comments, provide examples

### Risk: Performance Impact
**Mitigation:** Profile rendering, optimize if needed, cache templates

### Risk: Regression in Other Features
**Mitigation:** Run full test suite, manual regression testing

---

## Documentation Updates Required

1. `README-LLM.md`
   - Update Epic 06 status
   - Update Epic 11 status
   - Add component template examples

2. `docs/00_BACKLOG/11_STORY_05_theme_rendering_engine.md`
   - Mark as complete
   - Document final implementation

3. `docs/00_BACKLOG/11_STORY_07_theme_integration_testing.md`
   - Mark as complete

4. `docs/PROJECT_STATUS_ASSESSMENT.md`
   - Update epic status
   - Update confidence levels
   - Update test pass rates

5. `internal/templates/README.md` (create if missing)
   - Component template guide
   - Field access patterns
   - Examples for each component type

---

## Post-Fix Validation

### Manual Testing Checklist
- [ ] Load plain_text theme
- [ ] Load wedding_elegance theme
- [ ] Load birthday_celebration theme
- [ ] Load corporate_professional theme
- [ ] Load holiday_festive theme
- [ ] Load garden_party theme
- [ ] Load modern_minimalist theme
- [ ] Test custom image override
- [ ] Test custom color override
- [ ] Test component overrides
- [ ] Test mobile responsive
- [ ] Test light/dark mode
- [ ] Test with real event data
- [ ] Test with preference questions
- [ ] Test RSVP submission

### Performance Testing
- [ ] Measure page load time (<2s target)
- [ ] Check rendering time (<100ms target)
- [ ] Profile template execution
- [ ] Verify no memory leaks
- [ ] Check database query count

---

## Next Steps After Fix

Once component system is fixed:

1. **Fix Frontend Integration** (Epic 07)
   - Add missing meta tags
   - Fix CSS references
   - Fix loading states
   - Fix accessibility

2. **Fix Build Failures**
   - Fix invites package
   - Fix auth tests
   - Get everything building

3. **Start Security Testing** (Epic 09)
   - Cannot delay any longer
   - Critical for production

---

## Conclusion

**Estimated Fix Time:** 2-3 days  
**Confidence:** HIGH (95%)  
**Risk:** LOW  
**Recommended Approach:** Update templates (Option 1)  

The component system is well-designed but templates need updating. This is a straightforward fix with clear scope and low risk. Once complete, the Evite-style customization will work as intended.

**Priority:** CRITICAL - This blocks a major MVP feature
