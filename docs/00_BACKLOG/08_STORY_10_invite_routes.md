# User Story: Invite CRUD Routes

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 2 days

---

## User Story

As an **event manager**, I want **to create, read, update, and delete invites via HTTP API** so that **I can manage guest invitations through the web interface**.

---

## Acceptance Criteria

- [ ] GET /events/{id}/invites - List invites for event
- [ ] GET /events/{id}/invites/new - New invite form
- [ ] POST /events/{id}/invites - Create individual invite
- [ ] POST /events/{id}/invites/bulk - CSV bulk import
- [ ] GET /invites/{id} - View invite details
- [ ] PUT /invites/{id} - Update invite
- [ ] DELETE /invites/{id} - Delete invite
- [ ] POST /invites/{id}/revoke - Revoke token
- [ ] POST /invites/{id}/regenerate - Regenerate token
- [ ] POST /invites/{id}/send - Send invite email
- [ ] Permission checks on all routes
- [ ] Input validation
- [ ] CSRF protection on mutations

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

- [ ] Implement list invites handler
- [ ] Implement create invite handler
- [ ] Implement bulk import handler
- [ ] Implement get invite handler
- [ ] Implement update invite handler
- [ ] Implement delete invite handler
- [ ] Implement revoke handler
- [ ] Implement regenerate handler
- [ ] Implement send email handler
- [ ] Add permission checks
- [ ] Add CSV parsing
- [ ] Test all routes

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

- [ ] All acceptance criteria met
- [ ] All routes implemented
- [ ] CSV import working
- [ ] Permission checks working
- [ ] Tests passing
- [ ] Documentation complete
