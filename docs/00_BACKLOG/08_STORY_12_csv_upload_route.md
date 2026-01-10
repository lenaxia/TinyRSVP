# User Story: CSV Upload Route

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)
**Priority:** High
**Status:** ✅ Complete (Implemented in Story 10)
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-10

---

## User Story

As an **event manager**, I want **to upload CSV files for bulk invite import** so that **I can quickly invite many guests at once**.

---

## Acceptance Criteria

- [x] POST /events/{id}/invites/bulk - CSV upload endpoint (implemented as `/api/events/{eventId}/invites/import`)
- [x] Multipart form data handling
- [x] CSV parsing and validation
- [x] Batch invite creation
- [x] Error reporting per row
- [x] Success/failure summary
- [x] File size limits (1MB)
- [x] CSRF protection (via global middleware)

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

- [x] Implement CSV upload handler
- [x] Add file size validation
- [x] Add CSV format validation
- [x] Handle parsing errors
- [x] Test bulk import

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

- [x] All acceptance criteria met
- [x] CSV upload working
- [x] Error handling complete
- [x] Tests passing
- [x] Documentation complete

---

## Implementation Summary

**Story 12 was fully implemented as part of Story 10 (Invite Routes).**

### Actual Implementation Details

**Endpoint:** `POST /api/events/{eventId}/invites/import`

**Location:** [`internal/handlers/invites.go`](../../internal/handlers/invites.go) (lines 148-224)

**Key Features Implemented:**

1. **Multipart Form Data Handling** (line 188)
   - Uses `r.ParseMultipartForm(1 << 20)` to parse form with 1MB limit
   - Extracts file using `r.FormFile("file")`

2. **File Size Limits** (line 200-203)
   - Hard limit of 1MB enforced at form parsing level
   - Additional check on header.Size to reject oversized files
   - Returns 400 Bad Request with clear error message

3. **CSV Format Validation** (line 205-208)
   - Validates file extension is `.csv`
   - Case-insensitive check using `strings.ToLower()`

4. **Permission Checks** (line 183-186)
   - Requires authentication
   - Only admins or event creators can import
   - Returns 403 Forbidden for unauthorized users

5. **Event Status Validation** (line 173-181)
   - Rejects imports for cancelled events
   - Rejects imports for archived events
   - Returns 400 Bad Request with descriptive messages

6. **CSV Parsing & Batch Creation** (line 217)
   - Delegates to `h.service.ImportCSV()` for parsing and creation
   - Service handles row-by-row validation and creation

7. **Error Reporting** (via `invites.ImportResult`)
   - Returns structured result with:
     - `Total`: Total rows processed
     - `Created`: Successfully created invites
     - `Failed`: Failed rows
     - `Duplicates`: Duplicate email count
     - `Errors`: Array of per-row errors with row number, email, and message

8. **CSRF Protection**
   - Applied via global middleware in router (line 211 of router.go)
   - All POST requests protected automatically

### Test Coverage

**Unit Tests:** [`internal/handlers/invites_import_test.go`](../../internal/handlers/invites_import_test.go)
- Success scenarios
- Partial success with errors
- Missing file
- Invalid file extension
- File too large
- Unauthorized access
- Invalid event ID
- Service errors

**Permission Tests:** [`internal/handlers/invites_import_permission_test.go`](../../internal/handlers/invites_import_permission_test.go)
- Permission denied for non-admin/non-creator
- Permission granted for admin
- Permission granted for creator
- Event not found
- Cancelled event rejection
- Archived event rejection
- Correct expiration time calculation
- Correct default max_plus_ones

**Integration Tests:** [`internal/handlers/invites_import_integration_test.go`](../../internal/handlers/invites_import_integration_test.go)
- End-to-end CSV import
- Duplicate handling

**All 18 tests passing** ✅

### Differences from Story 12 Specification

| Story 12 Spec | Actual Implementation | Notes |
|---------------|----------------------|-------|
| `/events/{id}/invites/bulk` | `/api/events/{eventId}/invites/import` | More RESTful naming; under `/api` prefix |
| 10MB limit | 1MB limit | More conservative limit for homelab deployments |
| Form field: `csv` | Form field: `file` | Generic field name for flexibility |

### Related Documentation

- **Implementation Notes:** [2026-01-10_42_invite_routes.md](../01_WORKLOG/2026-01-10_42_invite_routes.md)
- **Story 10:** [08_STORY_10_invite_routes.md](08_STORY_10_invite_routes.md)
- **CSV Service:** [03_STORY_05_bulk_csv_import.md](03_STORY_05_bulk_csv_import.md)
