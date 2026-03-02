# User Story: Dashboard Route

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As an **event manager**, I want **a dashboard showing my events and activity** so that **I can quickly see the status of my events and RSVPs**.

---

## Acceptance Criteria

- [x] GET / - Dashboard page (requires auth)
- [x] Shows event statistics
- [x] Shows recent activity
- [x] Shows quick actions
- [x] Responsive layout
- [x] Loading states
- [x] Error handling

---

## Technical Details

### Route
```go
r.Get("/", handlers.Dashboard)
```

### Handler
```go
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
    user := GetUser(r.Context())
    
    stats, err := h.events.GetStatistics(r.Context(), user.ID)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    activity, err := h.events.GetRecentActivity(r.Context(), user.ID)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    data := struct {
        User     *models.User
        Stats    *Statistics
        Activity []*Activity
    }{
        User:     user,
        Stats:    stats,
        Activity: activity,
    }
    
    h.templates.Render(w, "dashboard.html", data)
}
```

---

## Tasks

- [x] Implement dashboard handler
- [x] Fetch event statistics
- [x] Fetch recent activity
- [x] Render dashboard template
- [x] Test dashboard route
- [x] Test authentication requirement

---

## Dependencies

**Depends on:** 
- 08_STORY_06_login_routes.md
- Epic 02 (Events)

**Blocks:** None

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **UI:** [07_STORY_08_dashboard_ui.md](07_STORY_08_dashboard_ui.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Dashboard route implemented
- [x] Statistics displayed
- [x] Tests passing
- [x] Documentation complete

---

## Status

- Status: Complete
- Started: 2026-01-10
- Completed: 2026-01-10
