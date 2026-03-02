# User Story: Bulk CSV Import

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1.5 days
**Actual Effort:** 1 day
**Completed:** 2026-01-07

---

## User Story

As an **event manager**, I want **to import multiple invites from a CSV file** so that **I can efficiently invite many guests at once**.

---

## Acceptance Criteria

- [x] Event manager can upload CSV file with guest list
- [x] CSV supports up to 500 rows
- [x] CSV header row required with 'email' column
- [x] Optional columns: 'name', 'max_plus_ones'
- [x] Email validation for each row
- [x] Duplicate email detection within CSV
- [x] Duplicate email detection against existing invites
- [x] Invalid rows reported with line numbers
- [x] Successful rows create invites
- [x] Transaction rollback on critical errors
- [x] Import summary returned (total, created, failed, duplicates)
- [x] All tokens generated securely
- [x] HTTP API endpoint for CSV upload

---

## Technical Details

### Package Location
- `internal/invites/csv.go` - CSV parsing logic
- `internal/invites/csv_test.go` - CSV tests
- `internal/invites/service.go` - Import service method
- `internal/handlers/invites.go` - Upload handler

### CSV Format

```csv
name,email,max_plus_ones
John Doe,john@example.com,2
Jane Smith,jane@example.com,1
Bob Johnson,bob@example.com,0
```

Minimum required format:
```csv
email
john@example.com
jane@example.com
```

### Service Interface

```go
type ImportResult struct {
    Total      int
    Created    int
    Failed     int
    Duplicates int
    Errors     []ImportError
}

type ImportError struct {
    Row     int
    Email   string
    Message string
}

func (s *service) ImportCSV(ctx context.Context, eventID int64, csvData []byte) (*ImportResult, error)
```

### HTTP Endpoint

```
POST /api/events/:eventId/invites/import
Content-Type: multipart/form-data

file: guests.csv

Response 200 OK:
{
    "total": 100,
    "created": 95,
    "failed": 3,
    "duplicates": 2,
    "errors": [
        {
            "row": 5,
            "email": "invalid-email",
            "message": "Invalid email format"
        },
        {
            "row": 12,
            "email": "duplicate@example.com",
            "message": "Email already invited"
        }
    ]
}
```

---

## Subtasks

### CSV Parser Implementation
- [x] Create CSV parser in `csv.go`
- [x] Parse header row and validate columns
- [x] Require 'email' column
- [x] Support optional 'name' and 'max_plus_ones' columns
- [x] Handle various CSV formats (comma delimiter)
- [x] Handle quoted fields
- [x] Handle empty rows
- [x] Trim whitespace from fields
- [x] Validate row count (max 500)

### Service Implementation
- [x] Implement `ImportCSV()` method
- [x] Parse CSV data
- [x] Validate each row
- [x] Check for duplicate emails within CSV
- [x] Check for duplicate emails in database
- [x] Generate tokens for all valid invites
- [x] Batch insert invites (transaction)
- [x] Collect and return errors
- [x] Return import summary

### Handler Implementation
- [x] Create POST `/api/events/:eventId/invites/import` endpoint
- [x] Handle multipart/form-data upload
- [x] Validate file size (max 1MB)
- [x] Validate file extension (.csv)
- [x] Read file contents
- [x] Call import service
- [x] Return import results

### Testing
- [x] Test valid CSV import
- [x] Test CSV with missing email column
- [x] Test CSV with invalid emails
- [x] Test CSV with duplicates
- [x] Test CSV exceeding 500 rows
- [x] Test empty CSV
- [x] Test malformed CSV
- [x] Test CSV injection attempts
- [x] Test transaction rollback
- [x] Test permission checks
- [x] Integration test full flow

### Documentation
- [x] CSV format specification
- [x] API endpoint documentation
- [x] Error handling guide
- [x] Example CSV files

---

## Dependencies

**Depends on:**
- Story 03: Invite Model
- Story 04: Individual Invite (reuses validation logic)

**Blocks:**
- Epic 05: Email (bulk email sending)

---

## Testing Strategy

### Unit Tests

1. **CSV Parser Tests**
   ```go
   func TestCSVParser_Parse_ValidCSV(t *testing.T)
   func TestCSVParser_Parse_MissingEmailColumn(t *testing.T)
   func TestCSVParser_Parse_EmptyCSV(t *testing.T)
   func TestCSVParser_Parse_MalformedCSV(t *testing.T)
   func TestCSVParser_Parse_QuotedFields(t *testing.T)
   func TestCSVParser_Parse_ExtraColumns(t *testing.T)
   ```

2. **Import Service Tests**
   ```go
   func TestImportCSV_Success(t *testing.T)
   func TestImportCSV_DuplicatesWithinCSV(t *testing.T)
   func TestImportCSV_DuplicatesInDatabase(t *testing.T)
   func TestImportCSV_InvalidEmails(t *testing.T)
   func TestImportCSV_ExceedsRowLimit(t *testing.T)
   func TestImportCSV_TransactionRollback(t *testing.T)
   func TestImportCSV_PartialSuccess(t *testing.T)
   ```

3. **Handler Tests**
   ```go
   func TestImportHandler_Success(t *testing.T)
   func TestImportHandler_InvalidFile(t *testing.T)
   func TestImportHandler_FileTooLarge(t *testing.T)
   func TestImportHandler_Unauthorized(t *testing.T)
   ```

---

## Validation Rules

### CSV File Validation
- Max file size: 1MB
- File extension: .csv
- Max rows: 500 (excluding header)
- Header row required
- Email column required

### Row Validation
- Email: Required, valid format, max 255 chars
- Name: Optional, max 100 chars
- Max plus ones: Optional, 0-10, defaults to event's max_plus_ones

### Duplicate Detection
1. Check for duplicates within CSV (case-insensitive)
2. Check for duplicates against existing invites
3. Report all duplicates in error list

---

## CSV Injection Prevention

Prevent CSV injection attacks:
- Sanitize all fields before storage
- Escape special characters (=, +, -, @)
- Validate field content
- Never execute formulas
- Log suspicious content

Example malicious CSV:
```csv
email,name
user@example.com,=1+1
user2@example.com,@SUM(A1:A10)
```

Sanitization:
```go
func sanitizeCSVField(field string) string {
    if len(field) > 0 && (field[0] == '=' || field[0] == '+' || 
        field[0] == '-' || field[0] == '@') {
        return "'" + field
    }
    return field
}
```

---

## Error Handling

| Error Condition | Error Type | HTTP Status | User Message |
|----------------|------------|-------------|--------------|
| Missing email column | `ValidationError` | 400 | "CSV must have 'email' column" |
| Too many rows | `ValidationError` | 400 | "CSV exceeds 500 row limit" |
| File too large | `ValidationError` | 400 | "File size exceeds 1MB limit" |
| Invalid file type | `ValidationError` | 400 | "File must be CSV format" |
| Malformed CSV | `ValidationError` | 400 | "Invalid CSV format" |
| Permission denied | `PermissionDeniedError` | 403 | "Not authorized to import invites" |
| Database error | `InternalError` | 500 | "Import failed" |

Row-level errors returned in `errors` array, not as HTTP errors.

---

## Performance Considerations

1. **Batch Processing**
   - Parse CSV in memory (max 1MB)
   - Validate all rows before database operations
   - Use batch insert for valid invites
   - Single transaction for atomicity

2. **Memory Management**
   - Stream large CSVs if needed (future)
   - Limit concurrent imports per user
   - Clean up temp files

3. **Database Optimization**
   - Prepare statement for batch insert
   - Use transaction for rollback
   - Index on (event_id, email) for duplicate check

---

## Import Flow

```
1. Upload CSV file
2. Validate file (size, type)
3. Parse CSV header
4. Validate header has 'email' column
5. Parse all rows
6. Validate each row
7. Check for duplicates within CSV
8. Check for duplicates in database
9. Generate tokens for valid invites
10. Begin transaction
11. Batch insert valid invites
12. Commit transaction
13. Return import summary
```

---

## Example CSV Files

### Minimal CSV
```csv
email
john@example.com
jane@example.com
bob@example.com
```

### Full CSV
```csv
name,email,max_plus_ones
John Doe,john@example.com,2
Jane Smith,jane@example.com,1
Bob Johnson,bob@example.com,0
Alice Williams,alice@example.com,3
```

### CSV with Errors
```csv
name,email,max_plus_ones
John Doe,john@example.com,2
Invalid User,not-an-email,1
Jane Smith,jane@example.com,1
Duplicate,john@example.com,0
```

Expected result:
- Total: 4
- Created: 2 (John, Jane)
- Failed: 1 (Invalid User)
- Duplicates: 1 (Duplicate)

---

## References

- **HLD:** Section 6.3 (Bulk Import)
- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md) Section 4.3
- **Story 04:** [03_STORY_04_individual_invite.md](03_STORY_04_individual_invite.md)
- **RFC 4180:** CSV format specification
- **OWASP:** CSV Injection Prevention

---

## Definition of Done

- [x] All acceptance criteria met
- [x] CSV parser implemented and tested
- [x] Import service implemented and tested
- [x] HTTP handler implemented and tested
- [x] Unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] CSV injection prevention verified
- [x] Performance acceptable (500 rows < 5 seconds)
- [x] Documentation complete
- [x] Example CSV files provided
- [x] Code reviewed (gaps fixed)
- [x] No linter warnings
