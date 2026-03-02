# User Story: Invite CRUD Routes

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)
**Priority:** High
**Status:** ✅ Complete
**Estimated Effort:** 2 days
**Actual Effort:** 1 day
**Completed:** 2026-01-10

---

## User Story

As an **event manager**, I want **to create, read, update, and delete invites via HTTP API** so that **I can manage guest invitations through the web interface**.

---

## Acceptance Criteria

- [x] GET /events/{id}/invites - List invites for event
- [x] GET /events/{id}/invites/new - New invite form (deferred to Story 11 - UI)
- [x] POST /events/{id}/invites - Create individual invite
- [x] POST /events/{id}/invites/bulk - CSV bulk import
- [x] GET /invites/{id} - View invite details
- [x] PUT /invites/{id} - Update invite
- [x] DELETE /invites/{id} - Delete invite
- [x] POST /invites/{id}/revoke - Revoke token
- [x] POST /invites/{id}/regenerate - Regenerate token
- [x] POST /invites/{id}/send - Send invite email
- [x] Permission checks on all routes
- [x] Input validation
- [x] CSRF protection on mutations

---

## Technical Details

### Routes
```go
r.Route("/events/{eventId}/invites", func(r chi.Router) {
    r.Use(RequireAuth)
    r.Use(RequireEventAccess)
    
    r.Get("/", handlers.ListInvites)
    r.Get("/new", handlers.NewInviteForm)
    r.Post("/", handlers.CreateInvite)
    r.Post("/bulk", handlers.BulkImportInvites)
})

r.Route("/invites/{id}", func(r chi.Router) {
    r.Use(RequireAuth)
    r.Use(RequireInviteAccess)
    
    r.Get("/", handlers.GetInvite)
    r.Put("/", handlers.UpdateInvite)
    r.Delete("/", handlers.DeleteInvite)
    r.Post("/revoke", handlers.RevokeInvite)
    r.Post("/regenerate", handlers.RegenerateInvite)
    r.Post("/send", handlers.SendInvite)
})
```

---

## Tasks

- [x] Implement list invites handler (already existed)
- [x] Implement create invite handler (already existed)
- [x] Implement bulk import handler (already existed)
- [x] Implement get invite handler
- [x] Implement update invite handler
- [x] Implement delete invite handler
- [x] Implement revoke handler (already existed)
- [x] Implement regenerate handler (already existed)
- [x] Implement send email handler
- [x] Add permission checks
- [x] Add CSV parsing (already existed)
- [x] Test all routes

---

## Dependencies

**Depends on:** 
- 08_STORY_08_event_routes.md
- Epic 03 (Invites service)
- Epic 05 (Email service)

**Blocks:** 08_STORY_11_invite_ui.md

---

## Handler Examples

### Create Invite
```go
func (h *Handlers) CreateInvite(w http.ResponseWriter, r *http.Request) {
    eventID := GetEventID(r)
    user := GetUser(r.Context())
    
    var req CreateInviteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        HandleError(w, r, NewValidationError("Invalid request"))
        return
    }
    
    req.EventID = eventID
    
    invite, token, err := h.invites.Create(r.Context(), &req)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "invite": invite,
        "token":  token,
        "rsvp_url": fmt.Sprintf("%s/rsvp/%s", h.baseURL, token),
    })
}
```

### Bulk Import
```go
func (h *Handlers) BulkImportInvites(w http.ResponseWriter, r *http.Request) {
    eventID := GetEventID(r)
    
    file, _, err := r.FormFile("csv")
    if err != nil {
        HandleError(w, r, NewValidationError("CSV file required"))
        return
    }
    defer file.Close()
    
    results, err := h.invites.BulkImport(r.Context(), eventID, file)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    json.NewEncoder(w).Encode(results)
}
```

---

## Testing Strategy

```go
func TestListInvites(t *testing.T)
func TestCreateInvite_Success(t *testing.T)
func TestBulkImport_Success(t *testing.T)
func TestBulkImport_InvalidCSV(t *testing.T)
func TestRevokeInvite_Success(t *testing.T)
func TestRegenerateInvite_Success(t *testing.T)
func TestSendInvite_Success(t *testing.T)
```

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **Invites Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All routes implemented
- [x] CSV import working
- [x] Permission checks working
- [x] Tests passing
- [x] Documentation complete

---

## Implementation Notes

See [2026-01-10_42_invite_routes.md](../01_WORKLOG/2026-01-10_42_invite_routes.md) for detailed implementation notes.

**Key Decisions:**
1. Form handler (GET /events/{id}/invites/new) deferred to Story 11 as it's UI-focused
2. Send operation generates new token for security
3. Cannot update/delete responded invites to maintain data integrity
4. Email sending is asynchronous via queue system
