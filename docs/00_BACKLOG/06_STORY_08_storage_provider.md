# User Story: Storage Provider Interface

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-09

---

## User Story

As a **developer**, I want **a pluggable storage provider interface** so that **asset storage can be implemented with local filesystem or S3-compatible backends**.

---

## Acceptance Criteria

- [x] Storage Provider interface defined
- [x] Interface supports all required operations
- [x] PutObject method defined
- [x] GetObject method defined
- [x] DeleteObject method defined
- [x] GetPublicURL method defined
- [x] ListObjects method defined
- [x] Context support for cancellation
- [x] Error handling standardized
- [x] All tests pass with timeout
- [x] Mock implementation provided for testing

---

## Technical Details

### Storage Provider Interface

```go
package storage

import (
    "context"
    "io"
)

type Provider interface {
    PutObject(ctx context.Context, path string, data io.Reader, contentType string) error
    GetObject(ctx context.Context, path string) (io.ReadCloser, error)
    DeleteObject(ctx context.Context, path string) error
    GetPublicURL(ctx context.Context, path string) (string, error)
    ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error)
}

type ObjectInfo struct {
    Path         string
    Size         int64
    ContentType  string
    LastModified time.Time
}
```

### Provider Configuration

```go
type Config struct {
    Type     string
    BasePath string
    BaseURL  string
    
    S3Endpoint  string
    S3Region    string
    S3Bucket    string
    S3AccessKey string
    S3SecretKey string
}

func NewProvider(config *Config) (Provider, error) {
    switch config.Type {
    case "local":
        return NewLocalProvider(config.BasePath, config.BaseURL), nil
    case "s3":
        return NewS3Provider(config)
    default:
        return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
    }
}
```

### Error Types

```go
type StorageError struct {
    Op      string
    Path    string
    Message string
    Err     error
}

func (e *StorageError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s %s: %s: %v", e.Op, e.Path, e.Message, e.Err)
    }
    return fmt.Sprintf("%s %s: %s", e.Op, e.Path, e.Message)
}

func (e *StorageError) Unwrap() error {
    return e.Err
}

var (
    ErrNotFound      = &StorageError{Message: "object not found"}
    ErrAlreadyExists = &StorageError{Message: "object already exists"}
    ErrAccessDenied  = &StorageError{Message: "access denied"}
)
```

---

## Interface Methods

### PutObject

**Purpose:** Store an object at the specified path

**Parameters:**
- `ctx` - Context for cancellation
- `path` - Storage path (e.g., "images/123/logo.png")
- `data` - Object data as io.Reader
- `contentType` - MIME type (e.g., "image/png")

**Returns:** Error if operation fails

**Behavior:**
- Create parent directories if needed
- Overwrite existing object
- Set content type metadata
- Return error if write fails

### GetObject

**Purpose:** Retrieve an object from storage

**Parameters:**
- `ctx` - Context for cancellation
- `path` - Storage path

**Returns:** io.ReadCloser with object data, error if not found

**Behavior:**
- Return ErrNotFound if object doesn't exist
- Caller must close returned ReadCloser
- Support streaming (don't load entire file)

### DeleteObject

**Purpose:** Delete an object from storage

**Parameters:**
- `ctx` - Context for cancellation
- `path` - Storage path

**Returns:** Error if operation fails

**Behavior:**
- Return nil if object doesn't exist (idempotent)
- Don't delete parent directories
- Return error if delete fails

### GetPublicURL

**Purpose:** Get publicly accessible URL for object

**Parameters:**
- `ctx` - Context for cancellation
- `path` - Storage path

**Returns:** Public URL string, error if operation fails

**Behavior:**
- Local FS: Return app-served URL
- S3: Return pre-signed URL or CloudFront URL
- URL should be accessible without authentication

### ListObjects

**Purpose:** List objects with given prefix

**Parameters:**
- `ctx` - Context for cancellation
- `prefix` - Path prefix (e.g., "images/123/")

**Returns:** Slice of ObjectInfo, error if operation fails

**Behavior:**
- Return empty slice if no objects match
- Include size and content type metadata
- Sort by path alphabetically

---

## Tasks

### Phase 1: Interface Definition (TDD)
- [ ] Define Provider interface
- [ ] Define ObjectInfo struct
- [ ] Define Config struct
- [ ] Define error types
- [ ] Write test for interface compliance
- [ ] Run tests (should pass)

### Phase 2: Mock Implementation (TDD)
- [ ] Create MockProvider struct
- [ ] Write test for MockProvider.PutObject
- [ ] Write test for MockProvider.GetObject
- [ ] Write test for MockProvider.DeleteObject
- [ ] Write test for MockProvider.GetPublicURL
- [ ] Write test for MockProvider.ListObjects
- [ ] Implement MockProvider
- [ ] Run tests (should pass)

### Phase 3: Provider Factory (TDD)
- [ ] Write test for NewProvider with local type
- [ ] Write test for NewProvider with s3 type
- [ ] Write test for NewProvider with invalid type
- [ ] Implement NewProvider
- [ ] Run tests (should pass)

### Phase 4: Documentation
- [ ] Document interface contract
- [ ] Document error handling
- [ ] Document implementation requirements
- [ ] Create implementation guide

---

## Mock Implementation

```go
package storage

import (
    "bytes"
    "context"
    "io"
    "sync"
)

type MockProvider struct {
    mu      sync.RWMutex
    objects map[string][]byte
    
    PutObjectFunc    func(ctx context.Context, path string, data io.Reader, contentType string) error
    GetObjectFunc    func(ctx context.Context, path string) (io.ReadCloser, error)
    DeleteObjectFunc func(ctx context.Context, path string) error
    GetPublicURLFunc func(ctx context.Context, path string) (string, error)
    ListObjectsFunc  func(ctx context.Context, prefix string) ([]ObjectInfo, error)
}

func NewMockProvider() *MockProvider {
    return &MockProvider{
        objects: make(map[string][]byte),
    }
}

func (m *MockProvider) PutObject(ctx context.Context, path string, data io.Reader, contentType string) error {
    if m.PutObjectFunc != nil {
        return m.PutObjectFunc(ctx, path, data, contentType)
    }
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    bytes, err := io.ReadAll(data)
    if err != nil {
        return err
    }
    
    m.objects[path] = bytes
    return nil
}

func (m *MockProvider) GetObject(ctx context.Context, path string) (io.ReadCloser, error) {
    if m.GetObjectFunc != nil {
        return m.GetObjectFunc(ctx, path)
    }
    
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    data, exists := m.objects[path]
    if !exists {
        return nil, ErrNotFound
    }
    
    return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MockProvider) DeleteObject(ctx context.Context, path string) error {
    if m.DeleteObjectFunc != nil {
        return m.DeleteObjectFunc(ctx, path)
    }
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    delete(m.objects, path)
    return nil
}

func (m *MockProvider) GetPublicURL(ctx context.Context, path string) (string, error) {
    if m.GetPublicURLFunc != nil {
        return m.GetPublicURLFunc(ctx, path)
    }
    
    return fmt.Sprintf("http://localhost:8080/assets/%s", path), nil
}

func (m *MockProvider) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
    if m.ListObjectsFunc != nil {
        return m.ListObjectsFunc(ctx, prefix)
    }
    
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    var objects []ObjectInfo
    for path, data := range m.objects {
        if strings.HasPrefix(path, prefix) {
            objects = append(objects, ObjectInfo{
                Path:         path,
                Size:         int64(len(data)),
                ContentType:  "application/octet-stream",
                LastModified: time.Now(),
            })
        }
    }
    
    sort.Slice(objects, func(i, j int) bool {
        return objects[i].Path < objects[j].Path
    })
    
    return objects, nil
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestProvider_Interface(t *testing.T) {
    var _ Provider = (*MockProvider)(nil)
    var _ Provider = (*localProvider)(nil)
}

func TestMockProvider_PutAndGet(t *testing.T) {
    provider := NewMockProvider()
    ctx := context.Background()
    
    data := []byte("test data")
    path := "test/file.txt"
    
    err := provider.PutObject(ctx, path, bytes.NewReader(data), "text/plain")
    if err != nil {
        t.Fatalf("PutObject() error = %v", err)
    }
    
    reader, err := provider.GetObject(ctx, path)
    if err != nil {
        t.Fatalf("GetObject() error = %v", err)
    }
    defer reader.Close()
    
    retrieved, _ := io.ReadAll(reader)
    if !bytes.Equal(retrieved, data) {
        t.Errorf("Retrieved data = %v, want %v", retrieved, data)
    }
}

func TestMockProvider_Delete(t *testing.T) {
    provider := NewMockProvider()
    ctx := context.Background()
    
    path := "test/file.txt"
    provider.PutObject(ctx, path, bytes.NewReader([]byte("data")), "text/plain")
    
    err := provider.DeleteObject(ctx, path)
    if err != nil {
        t.Fatalf("DeleteObject() error = %v", err)
    }
    
    _, err = provider.GetObject(ctx, path)
    if err != ErrNotFound {
        t.Errorf("GetObject() error = %v, want ErrNotFound", err)
    }
}

func TestMockProvider_ListObjects(t *testing.T) {
    provider := NewMockProvider()
    ctx := context.Background()
    
    paths := []string{
        "images/123/logo.png",
        "images/123/banner.jpg",
        "images/456/photo.png",
    }
    
    for _, path := range paths {
        provider.PutObject(ctx, path, bytes.NewReader([]byte("data")), "image/png")
    }
    
    objects, err := provider.ListObjects(ctx, "images/123/")
    if err != nil {
        t.Fatalf("ListObjects() error = %v", err)
    }
    
    if len(objects) != 2 {
        t.Errorf("ListObjects() returned %d objects, want 2", len(objects))
    }
}
```

---

## Dependencies

**Depends on:**
- None (foundational interface)

**Blocks:**
- Story 07: Image Upload (needs provider)
- Story 09: Local Storage (implements provider)
- Story 10: Asset Serving (uses provider)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Provider interface defined
- [ ] ObjectInfo struct defined
- [ ] Config struct defined
- [ ] Error types defined
- [ ] MockProvider implemented
- [ ] All unit tests passing (>90% coverage)
- [ ] Interface documentation complete
- [ ] Implementation guide created
- [ ] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 12.1 (Storage Provider Interface)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md) - Section 3.3
- **Package:** [`internal/storage/provider.go`](../../internal/storage/provider.go)
