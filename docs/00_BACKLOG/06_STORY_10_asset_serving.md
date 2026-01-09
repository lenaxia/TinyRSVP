# User Story: Asset Serving

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 0.5 days

---

## User Story

As a **guest**, I want **uploaded images to be accessible via public URLs** so that **I can see images in email invitations and RSVP pages**.

---

## Acceptance Criteria

- [x] Asset serving endpoint created
- [x] Images served with correct content type
- [x] Images served with caching headers
- [x] No authentication required for assets
- [x] Path traversal attacks prevented
- [x] 404 for missing assets
- [x] Support for all image formats
- [x] Efficient streaming (no full load into memory)
- [x] All tests pass with timeout
- [x] Performance verified

---

## Technical Details

### Asset Handler

```go
package handlers

import (
    "net/http"
    "path/filepath"
    "strings"
    "github.com/lenaxia/tinyrsvp/internal/storage"
)

type AssetHandler struct {
    provider storage.Provider
}

func NewAssetHandler(provider storage.Provider) *AssetHandler {
    return &AssetHandler{
        provider: provider,
    }
}

func (h *AssetHandler) ServeAsset(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet && r.Method != http.MethodHead {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    path := strings.TrimPrefix(r.URL.Path, "/assets/")
    
    if path == "" || strings.Contains(path, "..") {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }
    
    reader, err := h.provider.GetObject(r.Context(), path)
    if err != nil {
        if errors.Is(err, storage.ErrNotFound) {
            http.NotFound(w, r)
            return
        }
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer reader.Close()
    
    contentType := detectContentType(path)
    w.Header().Set("Content-Type", contentType)
    w.Header().Set("Cache-Control", "public, max-age=86400")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    
    if r.Method == http.MethodHead {
        return
    }
    
    if _, err := io.Copy(w, reader); err != nil {
        return
    }
}

func detectContentType(path string) string {
    ext := strings.ToLower(filepath.Ext(path))
    
    types := map[string]string{
        ".jpg":  "image/jpeg",
        ".jpeg": "image/jpeg",
        ".png":  "image/png",
        ".gif":  "image/gif",
        ".webp": "image/webp",
    }
    
    if ct, ok := types[ext]; ok {
        return ct
    }
    
    return "application/octet-stream"
}
```

---

## HTTP Endpoint

### Serve Asset
```
GET /assets/{path}
Authorization: None (public access)

Example:
GET /assets/images/123/logo_a1b2c3d4.png

Response 200 OK:
Content-Type: image/png
Cache-Control: public, max-age=86400
X-Content-Type-Options: nosniff

(binary image data)

Response 404 Not Found:
Asset not found

Response 400 Bad Request:
Invalid path
```

---

## Caching Strategy

### HTTP Headers

```go
w.Header().Set("Cache-Control", "public, max-age=86400")
w.Header().Set("ETag", generateETag(path))
w.Header().Set("Last-Modified", lastModified.Format(http.TimeFormat))
```

### Cache Duration
- Images: 24 hours (86400 seconds)
- Immutable assets: 1 year (31536000 seconds)
- Use unique filenames for cache busting

### ETag Generation

```go
func generateETag(path string, modTime time.Time) string {
    h := sha256.New()
    h.Write([]byte(path))
    h.Write([]byte(modTime.Format(time.RFC3339)))
    return fmt.Sprintf(`"%x"`, h.Sum(nil)[:16])
}
```

---

## Tasks

### Phase 1: Handler Implementation (TDD)
- [x] Define AssetHandler struct
- [x] Write test for ServeAsset success
- [x] Write test for ServeAsset not found
- [x] Write test for ServeAsset invalid path
- [x] Write test for ServeAsset method not allowed
- [x] Write test for HEAD request
- [x] Implement ServeAsset
- [x] Run tests (should pass)

### Phase 2: Content Type Detection (TDD)
- [x] Write test for detectContentType JPEG
- [x] Write test for detectContentType PNG
- [x] Write test for detectContentType GIF
- [x] Write test for detectContentType WebP
- [x] Write test for detectContentType unknown
- [x] Implement detectContentType
- [x] Run tests (should pass)

### Phase 3: Caching (TDD)
- [x] Write test for Cache-Control header
- [x] Write test for ETag generation
- [x] Write test for Last-Modified header
- [x] Write test for conditional requests
- [x] Implement caching headers
- [x] Run tests (should pass)

### Phase 4: Integration Testing
- [x] Test serving real images
- [x] Test serving from local provider
- [x] Test 404 handling
- [x] Test path traversal prevention
- [x] Test concurrent requests
- [x] Test caching behavior

---

## Security Considerations

### Path Traversal Prevention
- Strip "/assets/" prefix
- Reject paths with ".."
- Validate path before storage access
- Use filepath.Clean()

### Content Type Security
- Set X-Content-Type-Options: nosniff
- Prevent MIME type sniffing
- Serve with correct content type
- Prevent script execution

### Access Control
- Assets are public (no auth required)
- Used in emails and public pages
- No sensitive data in assets
- Event-scoped paths

---

## Performance Optimization

### Streaming
- Use io.Copy for efficient transfer
- Don't load entire file into memory
- Support range requests (future)

### Caching
- Long cache duration (24 hours)
- Unique filenames prevent stale cache
- ETag support for validation
- Conditional requests (304 Not Modified)

### Connection Handling
- Set appropriate timeouts
- Handle slow clients
- Limit concurrent connections

---

## Error Handling

| Error Condition | HTTP Status | Response |
|----------------|-------------|----------|
| Asset not found | 404 | "Asset not found" |
| Invalid path | 400 | "Invalid path" |
| Path traversal | 400 | "Invalid path" |
| Method not allowed | 405 | "Method not allowed" |
| Storage error | 500 | "Internal server error" |

---

## Testing Strategy

### Unit Tests

```go
func TestAssetHandler_ServeAsset(t *testing.T) {
    mockProvider := storage.NewMockProvider()
    handler := NewAssetHandler(mockProvider)
    
    testData := []byte("test image data")
    mockProvider.PutObject(context.Background(), "images/123/test.png", bytes.NewReader(testData), "image/png")
    
    tests := []struct {
        name       string
        path       string
        wantStatus int
        wantBody   bool
    }{
        {
            name:       "valid asset",
            path:       "/assets/images/123/test.png",
            wantStatus: http.StatusOK,
            wantBody:   true,
        },
        {
            name:       "not found",
            path:       "/assets/images/123/missing.png",
            wantStatus: http.StatusNotFound,
            wantBody:   false,
        },
        {
            name:       "path traversal",
            path:       "/assets/../../../etc/passwd",
            wantStatus: http.StatusBadRequest,
            wantBody:   false,
        },
        {
            name:       "empty path",
            path:       "/assets/",
            wantStatus: http.StatusBadRequest,
            wantBody:   false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, tt.path, nil)
            w := httptest.NewRecorder()
            
            handler.ServeAsset(w, req)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
            }
            
            if tt.wantBody && w.Body.Len() == 0 {
                t.Error("Expected response body")
            }
            
            if tt.wantStatus == http.StatusOK {
                ct := w.Header().Get("Content-Type")
                if ct == "" {
                    t.Error("Expected Content-Type header")
                }
                
                cc := w.Header().Get("Cache-Control")
                if !strings.Contains(cc, "max-age") {
                    t.Error("Expected Cache-Control header with max-age")
                }
            }
        })
    }
}

func TestAssetHandler_HeadRequest(t *testing.T) {
    mockProvider := storage.NewMockProvider()
    handler := NewAssetHandler(mockProvider)
    
    mockProvider.PutObject(context.Background(), "images/123/test.png", bytes.NewReader([]byte("data")), "image/png")
    
    req := httptest.NewRequest(http.MethodHead, "/assets/images/123/test.png", nil)
    w := httptest.NewRecorder()
    
    handler.ServeAsset(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
    }
    
    if w.Body.Len() > 0 {
        t.Error("HEAD request should not return body")
    }
    
    if w.Header().Get("Content-Type") == "" {
        t.Error("Expected Content-Type header")
    }
}
```

---

## Dependencies

**Depends on:**
- Story 08: Storage Provider (for provider interface)
- Story 09: Local Storage (for storage implementation)

**Blocks:**
- Story 03: Default Templates (can reference assets)
- Story 07: Image Upload (generates asset URLs)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] AssetHandler implemented
- [x] Content type detection implemented
- [x] Caching headers implemented
- [x] All unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] Security tests passing
- [x] Performance verified
- [x] Documentation updated
- [x] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 12 (Asset Storage)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **Story 08:** [06_STORY_08_storage_provider.md](06_STORY_08_storage_provider.md)
- **Story 09:** [06_STORY_09_local_storage.md](06_STORY_09_local_storage.md)
