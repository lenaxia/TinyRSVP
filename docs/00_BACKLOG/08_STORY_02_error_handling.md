# User Story: Error Handling and Response Formatting

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 1 day

---

## User Story

As a **developer**, I want **consistent error handling and response formatting** so that **clients receive clear, actionable error messages in a predictable format**.

---

## Acceptance Criteria

- [ ] Consistent error response format (JSON and HTML)
- [ ] User-friendly error messages
- [ ] Field-specific validation errors
- [ ] HTTP status codes correctly set
- [ ] Error logging with context
- [ ] Stack traces in development only
- [ ] Error types properly categorized
- [ ] Localization support for error messages
- [ ] Error response middleware
- [ ] Error recovery from panics

---

## Technical Details

### Package Location
- `internal/handlers/errors.go` - Error handling
- `internal/handlers/errors_test.go` - Error tests
- `internal/models/errors.go` - Error types
- `templates/web/error.html` - Error page template

### Error Response Format

```go
type ErrorResponse struct {
    Error   string            `json:"error"`
    Message string            `json:"message"`
    Code    string            `json:"code,omitempty"`
    Fields  map[string]string `json:"fields,omitempty"`
}

type APIError struct {
    StatusCode int
    Code       string
    Message    string
    Fields     map[string]string
    Err        error
}
```

---

## Tasks

### Error Types
- [ ] Define error type hierarchy
- [ ] Create validation error type
- [ ] Create not found error type
- [ ] Create permission denied error type
- [ ] Create conflict error type
- [ ] Create internal error type

### Response Formatting
- [ ] Implement JSON error formatter
- [ ] Implement HTML error formatter
- [ ] Create error response middleware
- [ ] Add content negotiation
- [ ] Format field validation errors

### Error Handling
- [ ] Create error handler function
- [ ] Map errors to HTTP status codes
- [ ] Log errors with context
- [ ] Sanitize error messages for clients
- [ ] Handle panic recovery

### Testing
- [ ] Test error type creation
- [ ] Test JSON formatting
- [ ] Test HTML formatting
- [ ] Test status code mapping
- [ ] Test field validation errors
- [ ] Integration test error flows

---

## Dependencies

**Depends on:** 08_STORY_01_middleware_chain.md

**Blocks:** All route handler stories

---

## Testing Strategy

### Unit Tests

```go
func TestErrorResponse_JSON(t *testing.T)
func TestErrorResponse_HTML(t *testing.T)
func TestErrorHandler_StatusCodes(t *testing.T)
func TestErrorHandler_ValidationErrors(t *testing.T)
func TestErrorHandler_PanicRecovery(t *testing.T)
```

---

## Error Type Definitions

```go
var (
    ErrValidation       = errors.New("validation error")
    ErrNotFound         = errors.New("not found")
    ErrPermissionDenied = errors.New("permission denied")
    ErrConflict         = errors.New("conflict")
    ErrInternal         = errors.New("internal error")
    ErrUnauthorized     = errors.New("unauthorized")
    ErrBadRequest       = errors.New("bad request")
)

func NewValidationError(fields map[string]string) *APIError {
    return &APIError{
        StatusCode: http.StatusBadRequest,
        Code:       "VALIDATION_ERROR",
        Message:    "Validation failed",
        Fields:     fields,
        Err:        ErrValidation,
    }
}

func NewNotFoundError(message string) *APIError {
    return &APIError{
        StatusCode: http.StatusNotFound,
        Code:       "NOT_FOUND",
        Message:    message,
        Err:        ErrNotFound,
    }
}
```

---

## Error Handler Implementation

```go
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
    apiErr := toAPIError(err)
    
    // Log error with context
    logError(r.Context(), apiErr)
    
    // Set status code
    w.WriteHeader(apiErr.StatusCode)
    
    // Format response based on Accept header
    if wantsJSON(r) {
        writeJSONError(w, apiErr)
    } else {
        writeHTMLError(w, r, apiErr)
    }
}

func toAPIError(err error) *APIError {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr
    }
    
    // Map standard errors
    switch {
    case errors.Is(err, ErrNotFound):
        return NewNotFoundError(err.Error())
    case errors.Is(err, ErrPermissionDenied):
        return NewPermissionDeniedError(err.Error())
    default:
        return NewInternalError()
    }
}
```

---

## JSON Error Response

```go
func writeJSONError(w http.ResponseWriter, err *APIError) {
    w.Header().Set("Content-Type", "application/json")
    
    resp := ErrorResponse{
        Error:   http.StatusText(err.StatusCode),
        Message: err.Message,
        Code:    err.Code,
        Fields:  err.Fields,
    }
    
    json.NewEncoder(w).Encode(resp)
}
```

---

## HTML Error Response

```go
func writeHTMLError(w http.ResponseWriter, r *http.Request, err *APIError) {
    w.Header().Set("Content-Type", "text/html")
    
    data := struct {
        StatusCode int
        Error      string
        Message    string
        RequestID  string
    }{
        StatusCode: err.StatusCode,
        Error:      http.StatusText(err.StatusCode),
        Message:    err.Message,
        RequestID:  GetRequestID(r.Context()),
    }
    
    renderTemplate(w, "error.html", data)
}
```

---

## Error Logging

```go
func logError(ctx context.Context, err *APIError) {
    requestID := GetRequestID(ctx)
    userID := GetUserID(ctx)
    
    log.Printf("[%s] Error %d: %s (user=%s, code=%s)",
        requestID,
        err.StatusCode,
        err.Message,
        userID,
        err.Code,
    )
    
    // Log stack trace for 500 errors
    if err.StatusCode >= 500 && err.Err != nil {
        log.Printf("[%s] Stack: %+v", requestID, err.Err)
    }
}
```

---

## Validation Error Example

```go
func validateEventRequest(req *CreateEventRequest) error {
    fields := make(map[string]string)
    
    if req.Title == "" {
        fields["title"] = "Title is required"
    }
    if req.StartTime.IsZero() {
        fields["start_time"] = "Start time is required"
    }
    if req.StartTime.Before(time.Now()) {
        fields["start_time"] = "Start time must be in the future"
    }
    
    if len(fields) > 0 {
        return NewValidationError(fields)
    }
    
    return nil
}
```

---

## Status Code Mapping

| Error Type | HTTP Status | Code |
|------------|-------------|------|
| Validation | 400 | VALIDATION_ERROR |
| Unauthorized | 401 | UNAUTHORIZED |
| Permission Denied | 403 | PERMISSION_DENIED |
| Not Found | 404 | NOT_FOUND |
| Conflict | 409 | CONFLICT |
| Internal | 500 | INTERNAL_ERROR |
| Timeout | 504 | TIMEOUT |

---

## Security Considerations

- Never expose internal error details to clients
- Sanitize error messages for XSS
- Don't leak database schema in errors
- Log full errors server-side only
- Rate limit error responses

---

## References

- **HLD:** Section 19 (Request Flow)
- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **Story 01:** [08_STORY_01_middleware_chain.md](08_STORY_01_middleware_chain.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Error types defined
- [ ] JSON formatting working
- [ ] HTML formatting working
- [ ] Status codes correct
- [ ] Error logging functional
- [ ] Unit tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] Documentation complete
- [ ] Code reviewed
- [ ] No linter warnings
