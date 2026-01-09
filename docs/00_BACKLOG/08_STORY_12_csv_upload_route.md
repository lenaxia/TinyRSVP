# User Story: CSV Upload Route

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As an **event manager**, I want **to upload CSV files for bulk invite import** so that **I can quickly invite many guests at once**.

---

## Acceptance Criteria

- [ ] POST /events/{id}/invites/bulk - CSV upload endpoint
- [ ] Multipart form data handling
- [ ] CSV parsing and validation
- [ ] Batch invite creation
- [ ] Error reporting per row
- [ ] Success/failure summary
- [ ] File size limits
- [ ] CSRF protection

---

## Technical Details

### Handler
```go
func (h *Handlers) BulkImportInvites(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20) // 10MB limit
    
    file, _, err := r.FormFile("csv")
    if err != nil {
        HandleError(w, r, NewValidationError("CSV file required"))
        return
    }
    defer file.Close()
    
    eventID := GetEventID(r)
    results, err := h.invites.BulkImport(r.Context(), eventID, file)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    json.NewEncoder(w).Encode(results)
}
```

---

## Tasks

- [ ] Implement CSV upload handler
- [ ] Add file size validation
- [ ] Add CSV format validation
- [ ] Handle parsing errors
- [ ] Test bulk import

---

## Dependencies

**Depends on:** 
- 08_STORY_10_invite_routes.md
- Epic 03 (CSV import service)

**Blocks:** None

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **CSV Import:** [03_STORY_05_bulk_csv_import.md](03_STORY_05_bulk_csv_import.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] CSV upload working
- [ ] Error handling complete
- [ ] Tests passing
- [ ] Documentation complete
