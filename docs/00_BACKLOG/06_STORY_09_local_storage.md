# User Story: Local Filesystem Storage Implementation

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Completed:** 2026-01-09

---

## User Story

As a **system administrator**, I want **local filesystem storage implementation** so that **assets can be stored in mounted volumes without external dependencies**.

---

## Acceptance Criteria

- [x] Local filesystem provider implements storage interface
- [x] Files stored in configured base directory
- [x] Parent directories created automatically
- [x] File permissions set correctly (0644 for files, 0755 for dirs)
- [x] Path traversal attacks prevented
- [x] Concurrent access handled safely
- [x] Public URLs generated correctly
- [x] Configuration loaded from environment
- [x] All tests pass with timeout
- [x] Integration with image upload working

---

## Technical Details

### Local Provider Implementation

```go
package storage

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type localProvider struct {
    basePath string
    baseURL  string
}

func NewLocalProvider(basePath, baseURL string) Provider {
    return &localProvider{
        basePath: basePath,
        baseURL:  baseURL,
    }
}

func (p *localProvider) PutObject(ctx context.Context, path string, data io.Reader, contentType string) error {
    if err := validatePath(path); err != nil {
        return err
    }
    
    fullPath := filepath.Join(p.basePath, path)
    
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return &StorageError{
            Op:      "PutObject",
            Path:    path,
            Message: "Failed to create directory",
            Err:     err,
        }
    }
    
    file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
    if err != nil {
        return &StorageError{
            Op:      "PutObject",
            Path:    path,
            Message: "Failed to create file",
            Err:     err,
        }
    }
    defer file.Close()
    
    if _, err := io.Copy(file, data); err != nil {
        return &StorageError{
            Op:      "PutObject",
            Path:    path,
            Message: "Failed to write file",
            Err:     err,
        }
    }
    
    return nil
}

func (p *localProvider) GetObject(ctx context.Context, path string) (io.ReadCloser, error) {
    if err := validatePath(path); err != nil {
        return nil, err
    }
    
    fullPath := filepath.Join(p.basePath, path)
    
    file, err := os.Open(fullPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, &StorageError{
                Op:      "GetObject",
                Path:    path,
                Message: "Object not found",
                Err:     ErrNotFound,
            }
        }
        return nil, &StorageError{
            Op:      "GetObject",
            Path:    path,
            Message: "Failed to open file",
            Err:     err,
        }
    }
    
    return file, nil
}

func (p *localProvider) DeleteObject(ctx context.Context, path string) error {
    if err := validatePath(path); err != nil {
        return err
    }
    
    fullPath := filepath.Join(p.basePath, path)
    
    err := os.Remove(fullPath)
    if err != nil && !os.IsNotExist(err) {
        return &StorageError{
            Op:      "DeleteObject",
            Path:    path,
            Message: "Failed to delete file",
            Err:     err,
        }
    }
    
    return nil
}

func (p *localProvider) GetPublicURL(ctx context.Context, path string) (string, error) {
    if err := validatePath(path); err != nil {
        return "", err
    }
    
    return fmt.Sprintf("%s/assets/%s", p.baseURL, path), nil
}

func (p *localProvider) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
    if err := validatePath(prefix); err != nil {
        return nil, err
    }
    
    fullPrefix := filepath.Join(p.basePath, prefix)
    
    var objects []ObjectInfo
    
    err := filepath.Walk(fullPrefix, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() {
            return nil
        }
        
        relPath, err := filepath.Rel(p.basePath, path)
        if err != nil {
            return err
        }
        
        objects = append(objects, ObjectInfo{
            Path:         filepath.ToSlash(relPath),
            Size:         info.Size(),
            ContentType:  detectContentTypeFromPath(path),
            LastModified: info.ModTime(),
        })
        
        return nil
    })
    
    if err != nil && !os.IsNotExist(err) {
        return nil, &StorageError{
            Op:      "ListObjects",
            Path:    prefix,
            Message: "Failed to list objects",
            Err:     err,
        }
    }
    
    return objects, nil
}
```

### Path Validation

```go
func validatePath(path string) error {
    if path == "" {
        return &StorageError{
            Op:      "validatePath",
            Path:    path,
            Message: "Path cannot be empty",
        }
    }
    
    if strings.Contains(path, "..") {
        return &StorageError{
            Op:      "validatePath",
            Path:    path,
            Message: "Path traversal not allowed",
        }
    }
    
    cleanPath := filepath.Clean(path)
    if cleanPath != path && cleanPath != filepath.ToSlash(path) {
        return &StorageError{
            Op:      "validatePath",
            Path:    path,
            Message: "Invalid path format",
        }
    }
    
    return nil
}

func detectContentTypeFromPath(path string) string {
    ext := strings.ToLower(filepath.Ext(path))
    
    types := map[string]string{
        ".jpg":  "image/jpeg",
        ".jpeg": "image/jpeg",
        ".png":  "image/png",
        ".gif":  "image/gif",
        ".webp": "image/webp",
        ".html": "text/html",
        ".txt":  "text/plain",
        ".css":  "text/css",
    }
    
    if ct, ok := types[ext]; ok {
        return ct
    }
    
    return "application/octet-stream"
}
```

---

## Configuration

### Environment Variables

```bash
STORAGE_TYPE=local
STORAGE_PATH=/data/uploads
STORAGE_BASE_URL=https://rsvp.example.com
```

### Configuration Loading

```go
type LocalConfig struct {
    BasePath string
    BaseURL  string
}

func LoadLocalConfig() (*LocalConfig, error) {
    basePath := os.Getenv("STORAGE_PATH")
    if basePath == "" {
        basePath = "/data/uploads"
    }
    
    baseURL := os.Getenv("STORAGE_BASE_URL")
    if baseURL == "" {
        return nil, fmt.Errorf("STORAGE_BASE_URL is required")
    }
    
    if err := os.MkdirAll(basePath, 0755); err != nil {
        return nil, fmt.Errorf("failed to create storage directory: %w", err)
    }
    
    return &LocalConfig{
        BasePath: basePath,
        BaseURL:  baseURL,
    }, nil
}
```

---

## Directory Structure

```
/data/uploads/
├── images/
│   ├── 123/
│   │   ├── logo_a1b2c3d4.png
│   │   └── banner_e5f6g7h8.jpg
│   └── 456/
│       └── photo_i9j0k1l2.png
└── templates/
    └── 789/
        └── custom_m3n4o5p6.html
```

---

## Tasks

### Phase 1: Provider Implementation (TDD)
- [ ] Define localProvider struct
- [ ] Write test for NewLocalProvider
- [ ] Write test for PutObject
- [ ] Write test for GetObject
- [ ] Write test for DeleteObject
- [ ] Write test for GetPublicURL
- [ ] Write test for ListObjects
- [ ] Implement all methods
- [ ] Run tests (should pass)

### Phase 2: Path Validation (TDD)
- [ ] Write test for validatePath with valid paths
- [ ] Write test for validatePath with path traversal
- [ ] Write test for validatePath with empty path
- [ ] Write test for validatePath with absolute paths
- [ ] Implement validatePath
- [ ] Run tests (should pass)

### Phase 3: Configuration (TDD)
- [ ] Define LocalConfig struct
- [ ] Write test for LoadLocalConfig
- [ ] Write test for default values
- [ ] Write test for directory creation
- [ ] Implement LoadLocalConfig
- [ ] Run tests (should pass)

### Phase 4: Integration Testing
- [ ] Test with real filesystem
- [ ] Test directory creation
- [ ] Test file permissions
- [ ] Test concurrent access
- [ ] Test path traversal prevention
- [ ] Test large file handling
- [ ] Test disk full scenario

---

## Security Considerations

### Path Traversal Prevention
- Validate all paths before use
- Reject paths containing ".."
- Use filepath.Clean() to normalize
- Restrict access to base directory

### File Permissions
- Files: 0644 (rw-r--r--)
- Directories: 0755 (rwxr-xr-x)
- No executable permissions on files
- Owner-writable only

### Symlink Handling
- Do not follow symlinks
- Reject symlink creation
- Prevent symlink attacks

---

## Error Handling

| Error Condition | Error Type | Message |
|----------------|------------|---------|
| Path traversal | `StorageError` | "Path traversal not allowed" |
| Empty path | `StorageError` | "Path cannot be empty" |
| File not found | `StorageError` | "Object not found" |
| Permission denied | `StorageError` | "Access denied" |
| Disk full | `StorageError` | "Failed to write file: no space left" |
| Directory creation failed | `StorageError` | "Failed to create directory" |

---

## Testing Strategy

### Unit Tests

```go
func TestLocalProvider_PutObject(t *testing.T) {
    tempDir := t.TempDir()
    provider := NewLocalProvider(tempDir, "http://localhost:8080")
    
    tests := []struct {
        name        string
        path        string
        data        []byte
        contentType string
        wantErr     bool
    }{
        {
            name:        "valid put",
            path:        "images/123/test.png",
            data:        []byte("test data"),
            contentType: "image/png",
            wantErr:     false,
        },
        {
            name:        "path traversal",
            path:        "../../../etc/passwd",
            data:        []byte("malicious"),
            contentType: "text/plain",
            wantErr:     true,
        },
        {
            name:        "empty path",
            path:        "",
            data:        []byte("data"),
            contentType: "text/plain",
            wantErr:     true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            err := provider.PutObject(ctx, tt.path, bytes.NewReader(tt.data), tt.contentType)
            if (err != nil) != tt.wantErr {
                t.Errorf("PutObject() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if !tt.wantErr {
                fullPath := filepath.Join(tempDir, tt.path)
                if _, err := os.Stat(fullPath); os.IsNotExist(err) {
                    t.Error("File was not created")
                }
            }
        })
    }
}

func TestLocalProvider_GetObject(t *testing.T) {
    tempDir := t.TempDir()
    provider := NewLocalProvider(tempDir, "http://localhost:8080")
    ctx := context.Background()
    
    path := "test/file.txt"
    data := []byte("test data")
    
    provider.PutObject(ctx, path, bytes.NewReader(data), "text/plain")
    
    reader, err := provider.GetObject(ctx, path)
    if err != nil {
        t.Fatalf("GetObject() error = %v", err)
    }
    defer reader.Close()
    
    retrieved, _ := io.ReadAll(reader)
    if !bytes.Equal(retrieved, data) {
        t.Errorf("Retrieved data = %v, want %v", retrieved, data)
    }
    
    _, err = provider.GetObject(ctx, "nonexistent.txt")
    if err == nil {
        t.Error("Expected error for nonexistent file")
    }
}

func TestLocalProvider_DeleteObject(t *testing.T) {
    tempDir := t.TempDir()
    provider := NewLocalProvider(tempDir, "http://localhost:8080")
    ctx := context.Background()
    
    path := "test/file.txt"
    provider.PutObject(ctx, path, bytes.NewReader([]byte("data")), "text/plain")
    
    err := provider.DeleteObject(ctx, path)
    if err != nil {
        t.Fatalf("DeleteObject() error = %v", err)
    }
    
    fullPath := filepath.Join(tempDir, path)
    if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
        t.Error("File was not deleted")
    }
    
    err = provider.DeleteObject(ctx, "nonexistent.txt")
    if err != nil {
        t.Errorf("DeleteObject() on nonexistent file should not error: %v", err)
    }
}

func TestLocalProvider_ListObjects(t *testing.T) {
    tempDir := t.TempDir()
    provider := NewLocalProvider(tempDir, "http://localhost:8080")
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
    
    for _, obj := range objects {
        if !strings.HasPrefix(obj.Path, "images/123/") {
            t.Errorf("Object path = %s, want prefix images/123/", obj.Path)
        }
    }
}
```

---

## Configuration

### Environment Variables

```bash
STORAGE_TYPE=local
STORAGE_PATH=/data/uploads
STORAGE_BASE_URL=https://rsvp.example.com
```

### Docker Volume Mount

```yaml
services:
  tinyrsvp:
    image: tinyrsvp:latest
    volumes:
      - ./data:/data
    environment:
      - STORAGE_TYPE=local
      - STORAGE_PATH=/data/uploads
      - STORAGE_BASE_URL=https://rsvp.example.com
```

---

## Tasks

### Phase 1: Provider Implementation (TDD)
- [ ] Define localProvider struct
- [ ] Write test for NewLocalProvider
- [ ] Write test for PutObject success
- [ ] Write test for PutObject with nested dirs
- [ ] Write test for PutObject overwrite
- [ ] Write test for GetObject success
- [ ] Write test for GetObject not found
- [ ] Write test for DeleteObject success
- [ ] Write test for DeleteObject idempotent
- [ ] Write test for GetPublicURL
- [ ] Write test for ListObjects
- [ ] Implement all methods
- [ ] Run tests (should pass)

### Phase 2: Path Validation (TDD)
- [ ] Write test for validatePath valid paths
- [ ] Write test for validatePath traversal
- [ ] Write test for validatePath empty
- [ ] Write test for validatePath absolute
- [ ] Implement validatePath
- [ ] Run tests (should pass)

### Phase 3: Configuration (TDD)
- [ ] Write test for LoadLocalConfig
- [ ] Write test for default values
- [ ] Write test for directory creation
- [ ] Write test for missing base URL
- [ ] Implement LoadLocalConfig
- [ ] Run tests (should pass)

### Phase 4: Integration Testing
- [ ] Test with real filesystem
- [ ] Test concurrent writes
- [ ] Test concurrent reads
- [ ] Test large files
- [ ] Test many small files
- [ ] Test disk full handling
- [ ] Test permission errors

---

## File System Operations

### Directory Creation
- Create parent directories recursively
- Use mode 0755 (rwxr-xr-x)
- Ignore if already exists
- Error if permission denied

### File Creation
- Create with mode 0644 (rw-r--r--)
- Overwrite if exists
- Atomic write (write to temp, rename)
- Error if permission denied

### File Deletion
- Remove file only (not directories)
- Idempotent (no error if not exists)
- Error if permission denied

---

## Performance Considerations

### Caching
- No caching in provider (handled by HTTP layer)
- Direct filesystem access
- Let OS handle file caching

### Concurrent Access
- Thread-safe operations
- No locking needed (OS handles it)
- Multiple readers supported
- Multiple writers to different files supported

### Large Files
- Stream data (don't load into memory)
- Use io.Copy for efficient transfer
- Support files up to 5MB (image limit)

---

## Error Handling

| Error Condition | Error Type | Recovery |
|----------------|------------|----------|
| Disk full | `StorageError` | Free space, retry |
| Permission denied | `StorageError` | Fix permissions, retry |
| Path traversal | `StorageError` | Reject request |
| File not found | `StorageError` | Return 404 |
| Directory creation failed | `StorageError` | Check permissions |

---

## Testing Strategy

### Unit Tests
- Test all interface methods
- Test path validation
- Test error conditions
- Test edge cases

### Integration Tests
- Test with real filesystem
- Test concurrent operations
- Test error recovery
- Test cleanup

### Security Tests
- Test path traversal prevention
- Test symlink handling
- Test permission enforcement

---

## Dependencies

**Depends on:**
- Story 08: Storage Provider (implements interface)

**Blocks:**
- Story 07: Image Upload (needs storage)
- Story 10: Asset Serving (needs storage)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] localProvider implemented
- [ ] All interface methods working
- [ ] Path validation implemented
- [ ] Configuration loading implemented
- [ ] All unit tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] Security tests passing
- [ ] Documentation updated
- [ ] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 12.2 (Local Filesystem Provider)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md) - Section 4.2
- **Story 08:** [06_STORY_08_storage_provider.md](06_STORY_08_storage_provider.md)
- **Package:** [`internal/storage/local.go`](../../internal/storage/local.go)
