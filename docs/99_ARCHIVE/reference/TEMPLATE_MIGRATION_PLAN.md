# Template Migration Plan

## Status: IN PROGRESS

## Completed Migrations ✅
1. **dashboard.html** - Migrated to base template
2. **event_list.html** - Migrated to base template  
3. **admin_dashboard.html** - Migrated to base template

## Remaining Pages to Migrate

### High Priority (Breaking Pages)
These pages are loaded with already-migrated pages and will fail:

1. **event_form.html** - Loaded with eventWebTemplates
   - Complex page with datetime picker HTML embedded
   - Needs datetime picker partial created first
   - Has theme picker integration
   - ~300 lines

2. **event_detail.html** - Loaded with eventWebTemplates
   - Event display page
   - Needs migration to base template

### Medium Priority (Standalone Pages)
These pages have their own template loaders:

3. **invite_list.html** - Has own template loader (inviteListTemplates)
4. **user_management.html** - Has own template loader (userManagementTemplates)
5. **rsvp_summary.html** - Has own template loader (rsvpSummaryTemplates)

### Low Priority (Guest-Facing Pages)
These don't use navigation and may have different requirements:

6. **rsvp_page.html** - Guest RSVP page (no navigation)
7. **confirmation.html** - RSVP confirmation page
8. **unsubscribe.html** - Unsubscribe page

## Required Components to Create

### 1. DateTime Picker Partial
The datetime picker HTML (150+ lines) is currently embedded in event_form.html.
Should be extracted to: `templates/web/partials/datetime_picker.html`

**Usage:**
```html
{{template "datetime-picker" .}}
```

### 2. Timezone Picker Partial  
The timezone picker HTML is also embedded in event_form.html.
Should be extracted to: `templates/web/partials/timezone_picker.html`

**Usage:**
```html
{{template "timezone-picker" .}}
```

## Migration Strategy

### Phase 1: Extract Reusable Components
1. Create datetime_picker.html partial
2. Create timezone_picker.html partial
3. Update event_form.html to use partials

### Phase 2: Migrate Event Pages
1. Migrate event_form.html to base template
2. Migrate event_detail.html to base template
3. Test event creation/editing workflow

### Phase 3: Migrate Admin Pages
1. Migrate invite_list.html to base template
2. Migrate user_management.html to base template
3. Update template loading in main.go

### Phase 4: Migrate Summary Pages
1. Migrate rsvp_summary.html to base template
2. Update template loading in main.go

### Phase 5: Guest Pages (Optional)
1. Decide if guest pages should use base template
2. May need separate base template without navigation
3. Migrate if appropriate

## Template Loading Updates Needed

After each migration, update `cmd/server/main.go`:

```go
// Add base.html to ParseFiles calls
template.New("name").ParseFiles(
    "templates/web/partials/base.html",  // ADD THIS
    "templates/web/partials/navigation.html",
    "templates/web/page.html",
)
```

## Testing Checklist

After each migration:
- [ ] Page renders without errors
- [ ] All functionality works (forms, buttons, links)
- [ ] Styling is consistent with other pages
- [ ] Mobile responsive behavior works
- [ ] No console errors
- [ ] Navigation works correctly

## Rollback Plan

If migrations cause issues:
1. Revert template files: `git checkout templates/web/[page].html`
2. Revert main.go changes: `git checkout cmd/server/main.go`
3. Restart server

## Current Blocker

Event pages (event_form, event_detail) are failing because they're loaded with event_list.html but haven't been migrated yet. Need to complete Phase 1 and Phase 2 to unblock.
