# Storage Package

## Purpose
Provides pluggable storage provider interface for asset management.

## Rules
- All storage operations must use context for cancellation
- Providers must be thread-safe
- GetObject returns ReadCloser that caller must close
- DeleteObject is idempotent (no error if object doesn't exist)

## Structure
- `provider.go` - Provider interface and error types
- `provider_test.go` - Interface compliance tests
- `mock.go` - MockProvider for testing

## Key Components

### Provider Interface
```go
type Provider interface {
    PutObject(ctx context.Context, path string, data io.Reader, contentType string) error
    GetObject(ctx context.Context, path string) (io.ReadCloser, error)
    DeleteObject(ctx context.Context, path string) error
    GetPublicURL(ctx context.Context, path string) (string, error)
    ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error)
}
```

### ObjectInfo
Metadata about stored objects:
- Path
- Size
- ContentType
- LastModified

### StorageError
Structured error with operation context:
- Op (operation name)
- Path (object path)
- Message (error description)
- Err (wrapped error)

## Error Types
- `ErrNotFound` - Object does not exist
- `ErrAlreadyExists` - Object already exists
- `ErrAccessDenied` - Permission denied

## MockProvider
In-memory provider for testing:
- Thread-safe with mutex
- Supports all Provider operations
- Allows function injection for custom behavior
- Automatically sorts ListObjects results

## Future Implementations
- Local filesystem provider (Story 09)
- S3-compatible provider (v1+)

## Usage Example
```go
provider := storage.NewMockProvider()

// Store object
err := provider.PutObject(ctx, "images/123/logo.jpg", reader, "image/jpeg")

// Retrieve object
data, err := provider.GetObject(ctx, "images/123/logo.jpg")
defer data.Close()

// Get public URL
url, err := provider.GetPublicURL(ctx, "images/123/logo.jpg")

// List objects
objects, err := provider.ListObjects(ctx, "images/123/")

// Delete object
err = provider.DeleteObject(ctx, "images/123/logo.jpg")
```
