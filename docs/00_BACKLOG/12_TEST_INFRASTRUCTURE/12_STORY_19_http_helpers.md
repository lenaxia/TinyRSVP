# User Story: HTTP Test Helpers

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 1-2 hours
**Phase:** 5 - Advanced Features

---

## User Story

As a **developer**, I want **fluent HTTP test helpers** so that **I can construct test requests and assert responses with clean, readable code**.

---

## Acceptance Criteria

- [ ] `internal/testutil/http.go` created
- [ ] HTTPRequestBuilder implemented
- [ ] HTTPResponseChecker implemented
- [ ] All helpers have tests
- [ ] Documentation with examples

---

## Implementation

### HTTPRequestBuilder

```go
package testutil

type HTTPRequestBuilder struct {
    t      *testing.T
    method string
    path   string
    body   io.Reader
    headers map[string]string
}

func NewHTTPRequest(t *testing.T, method, path string) *HTTPRequestBuilder
func (b *HTTPRequestBuilder) WithJSONBody(v interface{}) *HTTPRequestBuilder
func (b *HTTPRequestBuilder) WithAuth(token string) *HTTPRequestBuilder
func (b *HTTPRequestBuilder) WithHeader(key, value string) *HTTPRequestBuilder
func (b *HTTPRequestBuilder) Build() *http.Request
```

### HTTPResponseChecker

```go
type HTTPResponseChecker struct {
    t    *testing.T
    resp *httptest.ResponseRecorder
}

func NewResponseChecker(t *testing.T, resp *httptest.ResponseRecorder) *HTTPResponseChecker
func (c *HTTPResponseChecker) ExpectStatus(expected int) *HTTPResponseChecker
func (c *HTTPResponseChecker) ExpectJSON(target interface{}) *HTTPResponseChecker
func (c *HTTPResponseChecker) ExpectHeader(key, expected string) *HTTPResponseChecker
```

---

## Usage Example

**Before:**
```go
body, _ := json.Marshal(createEventReq)
req := httptest.NewRequest("POST", "/api/events", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)

resp := httptest.NewRecorder()
handler.ServeHTTP(resp, req)

if resp.Code != 200 {
    t.Errorf("Expected 200, got %d", resp.Code)
}

var event models.Event
json.NewDecoder(resp.Body).Decode(&event)
```

**After:**
```go
req := testutil.NewHTTPRequest(t, "POST", "/api/events").
    WithJSONBody(createEventReq).
    WithAuth(token).
    Build()

resp := httptest.NewRecorder()
handler.ServeHTTP(resp, req)

var event models.Event
testutil.NewResponseChecker(t, resp).
    ExpectStatus(200).
    ExpectJSON(&event)
```

---

## Tasks

- [ ] Implement HTTPRequestBuilder with tests
- [ ] Implement HTTPResponseChecker with tests
- [ ] Add documentation and examples
- [ ] Update README.md

---

## Dependencies

**Depends on:** Story 17
