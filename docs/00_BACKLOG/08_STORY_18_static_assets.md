# User Story: Static Asset Serving

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As a **user**, I want **static assets (CSS, JS, images) served efficiently** so that **the web interface loads quickly and looks correct**.

---

## Acceptance Criteria

- [ ] GET /assets/* - Serve static files
- [ ] CSS files served with correct MIME type
- [ ] JavaScript files served with correct MIME type
- [ ] Images served with correct MIME type
- [ ] Cache headers set appropriately
- [ ] Gzip compression enabled
- [ ] 404 for missing assets
- [ ] Security headers on asset responses
- [ ] No directory listing
- [ ] Asset versioning support

---

## Technical Details

### Route
```go
r.Get("/assets/*", handlers.ServeStatic)
```

### Handler
```go
func (h *Handlers) ServeStatic(w http.ResponseWriter, r *http.Request) {
    fs := http.FileServer(http.Dir("./static"))
    
    // Strip /assets prefix
    http.StripPrefix("/assets", fs).ServeHTTP(w, r)
}
```

### With Caching
```go
func ServeStatic(staticDir string) http.HandlerFunc {
    fs := http.FileServer(http.Dir(staticDir))
    
    return func(w http.ResponseWriter, r *http.Request) {
        // Set cache headers
        w.Header().Set("Cache-Control", "public, max-age=31536000")
        w.Header().Set("Vary", "Accept-Encoding")
        
        // Serve file
        http.StripPrefix("/assets", fs).ServeHTTP(w, r)
    }
}
```

---

## Tasks

- [ ] Implement static file handler
- [ ] Configure MIME types
- [ ] Set cache headers
- [ ] Enable gzip compression
- [ ] Prevent directory listing
- [ ] Add asset versioning
- [ ] Test asset serving
- [ ] Test cache headers

---

## Dependencies

**Depends on:** 08_STORY_00_router_setup.md

**Blocks:** None

---

## Static Asset Structure

```
static/
├── css/
│   ├── main.css
│   ├── dashboard.css
│   └── rsvp.css
├── js/
│   ├── main.js
│   ├── form-validation.js
│   └── loading-states.js
└── images/
    ├── logo.png
    └── icons/
```

---

## Cache Headers

```
Cache-Control: public, max-age=31536000
Vary: Accept-Encoding
ETag: "abc123"
```

---

## MIME Types

| Extension | MIME Type |
|-----------|-----------|
| .css | text/css |
| .js | application/javascript |
| .png | image/png |
| .jpg | image/jpeg |
| .svg | image/svg+xml |
| .woff2 | font/woff2 |

---

## Asset Versioning

```html
<link rel="stylesheet" href="/assets/css/main.css?v=abc123">
<script src="/assets/js/main.js?v=abc123"></script>
```

---

## Gzip Compression

```go
func GzipMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }
        
        w.Header().Set("Content-Encoding", "gzip")
        gz := gzip.NewWriter(w)
        defer gz.Close()
        
        gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
        next.ServeHTTP(gzw, r)
    })
}
```

---

## Security Considerations

- Prevent directory traversal attacks
- No directory listing
- Validate file paths
- Set security headers
- Limit file size
- Rate limit asset requests

---

## Testing Strategy

```go
func TestServeStatic_CSS(t *testing.T)
func TestServeStatic_JavaScript(t *testing.T)
func TestServeStatic_Image(t *testing.T)
func TestServeStatic_NotFound(t *testing.T)
func TestServeStatic_CacheHeaders(t *testing.T)
func TestServeStatic_Gzip(t *testing.T)
```

---

## Performance Targets

- CSS load: <100ms
- JS load: <100ms
- Image load: <200ms
- Gzip compression: 70%+ reduction
- Cache hit rate: >90%

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Static assets serving correctly
- [ ] Cache headers set
- [ ] Gzip compression working
- [ ] Tests passing
- [ ] Performance targets met
- [ ] Documentation complete
