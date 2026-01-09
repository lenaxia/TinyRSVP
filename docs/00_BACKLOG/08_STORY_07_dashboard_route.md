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

- [ ] GET / - Dashboard page (requires auth)
- [ ] Shows event statistics
- [ ] Shows recent activity
- [ ] Shows quick actions
- [ ] Responsive layout
- [ ] Loading states
- [ ] Error handling

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

- [ ] Implement dashboard handler
- [ ] Fetch event statistics
- [ ] Fetch recent activity
- [ ] Render dashboard template
- [ ] Test dashboard route
- [ ] Test authentication requirement

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

- [ ] All acceptance criteria met
- [ ] Dashboard route implemented
- [ ] Statistics displayed
- [ ] Tests passing
- [ ] Documentation complete
