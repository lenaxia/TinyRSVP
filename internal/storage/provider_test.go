package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestProvider_Interface(t *testing.T) {
	var _ Provider = (*MockProvider)(nil)
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

func TestMockProvider_GetObject_NotFound(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	_, err := provider.GetObject(ctx, "nonexistent/file.txt")
	if err == nil {
		t.Error("GetObject() expected error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("GetObject() error = %v, want error containing 'not found'", err)
	}
	if err != ErrNotFound {
		t.Errorf("GetObject() error = %v, want ErrNotFound", err)
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

func TestMockProvider_Delete_Idempotent(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	path := "test/file.txt"

	err := provider.DeleteObject(ctx, path)
	if err != nil {
		t.Errorf("DeleteObject() on nonexistent file should be idempotent, got error = %v", err)
	}
}

func TestMockProvider_GetPublicURL(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	path := "images/123/logo.png"

	url, err := provider.GetPublicURL(ctx, path)
	if err != nil {
		t.Fatalf("GetPublicURL() error = %v", err)
	}

	if url == "" {
		t.Error("GetPublicURL() returned empty URL")
	}

	if !strings.Contains(url, path) {
		t.Errorf("GetPublicURL() = %v, want URL containing %v", url, path)
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

	for _, obj := range objects {
		if !strings.HasPrefix(obj.Path, "images/123/") {
			t.Errorf("Object path = %v, want prefix images/123/", obj.Path)
		}
		if obj.Size == 0 {
			t.Error("Object size should not be 0")
		}
	}
}

func TestMockProvider_ListObjects_Empty(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	objects, err := provider.ListObjects(ctx, "nonexistent/")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(objects) != 0 {
		t.Errorf("ListObjects() returned %d objects, want 0", len(objects))
	}
}

func TestMockProvider_ListObjects_Sorted(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	paths := []string{
		"files/zebra.txt",
		"files/alpha.txt",
		"files/beta.txt",
	}

	for _, path := range paths {
		provider.PutObject(ctx, path, bytes.NewReader([]byte("data")), "text/plain")
	}

	objects, err := provider.ListObjects(ctx, "files/")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(objects) != 3 {
		t.Fatalf("ListObjects() returned %d objects, want 3", len(objects))
	}

	if objects[0].Path != "files/alpha.txt" {
		t.Errorf("First object = %v, want files/alpha.txt", objects[0].Path)
	}
	if objects[1].Path != "files/beta.txt" {
		t.Errorf("Second object = %v, want files/beta.txt", objects[1].Path)
	}
	if objects[2].Path != "files/zebra.txt" {
		t.Errorf("Third object = %v, want files/zebra.txt", objects[2].Path)
	}
}

func TestMockProvider_Overwrite(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	path := "test/file.txt"
	data1 := []byte("original data")
	data2 := []byte("updated data")

	provider.PutObject(ctx, path, bytes.NewReader(data1), "text/plain")
	provider.PutObject(ctx, path, bytes.NewReader(data2), "text/plain")

	reader, err := provider.GetObject(ctx, path)
	if err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}
	defer reader.Close()

	retrieved, _ := io.ReadAll(reader)
	if !bytes.Equal(retrieved, data2) {
		t.Errorf("Retrieved data = %v, want %v", retrieved, data2)
	}
}

func TestStorageError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *StorageError
		want string
	}{
		{
			name: "with wrapped error",
			err: &StorageError{
				Op:      "GetObject",
				Path:    "test/file.txt",
				Message: "failed to read",
				Err:     io.EOF,
			},
			want: "GetObject test/file.txt: failed to read: EOF",
		},
		{
			name: "without wrapped error",
			err: &StorageError{
				Op:      "PutObject",
				Path:    "test/file.txt",
				Message: "access denied",
			},
			want: "PutObject test/file.txt: access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorageError_Unwrap(t *testing.T) {
	innerErr := io.EOF
	err := &StorageError{
		Op:      "GetObject",
		Path:    "test/file.txt",
		Message: "failed",
		Err:     innerErr,
	}

	if err.Unwrap() != innerErr {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), innerErr)
	}
}

func TestNewProvider_MockType(t *testing.T) {
	config := &Config{
		Type: "mock",
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	if provider == nil {
		t.Error("NewProvider() returned nil provider")
	}

	var _ Provider = provider
}

func TestNewProvider_InvalidType(t *testing.T) {
	config := &Config{
		Type: "invalid",
	}

	provider, err := NewProvider(config)
	if err == nil {
		t.Error("NewProvider() expected error for invalid type")
	}

	if provider != nil {
		t.Error("NewProvider() should return nil provider on error")
	}

	if !strings.Contains(err.Error(), "unsupported storage type") {
		t.Errorf("Error = %v, want error containing 'unsupported storage type'", err)
	}
}

func TestNewProvider_LocalType(t *testing.T) {
	config := &Config{
		Type:     "local",
		BasePath: "/tmp/storage",
		BaseURL:  "http://localhost:8080",
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	if provider == nil {
		t.Error("NewProvider() returned nil provider")
	}

	var _ Provider = provider
}

func TestNewProvider_LocalType_MissingBasePath(t *testing.T) {
	config := &Config{
		Type:    "local",
		BaseURL: "http://localhost:8080",
	}

	provider, err := NewProvider(config)
	if err == nil {
		t.Error("NewProvider() expected error for missing BasePath")
	}

	if provider != nil {
		t.Error("NewProvider() should return nil provider on error")
	}

	if !strings.Contains(err.Error(), "BasePath is required") {
		t.Errorf("Error = %v, want error containing 'BasePath is required'", err)
	}
}

func TestNewProvider_LocalType_MissingBaseURL(t *testing.T) {
	config := &Config{
		Type:     "local",
		BasePath: "/tmp/storage",
	}

	provider, err := NewProvider(config)
	if err == nil {
		t.Error("NewProvider() expected error for missing BaseURL")
	}

	if provider != nil {
		t.Error("NewProvider() should return nil provider on error")
	}

	if !strings.Contains(err.Error(), "BaseURL is required") {
		t.Errorf("Error = %v, want error containing 'BaseURL is required'", err)
	}
}

func TestNewProvider_S3Type(t *testing.T) {
	config := &Config{
		Type:        "s3",
		S3Endpoint:  "s3.amazonaws.com",
		S3Region:    "us-east-1",
		S3Bucket:    "test-bucket",
		S3AccessKey: "access",
		S3SecretKey: "secret",
	}

	provider, err := NewProvider(config)
	if err == nil {
		t.Error("NewProvider() expected error for unimplemented s3 type")
	}

	if provider != nil {
		t.Error("NewProvider() should return nil provider on error")
	}

	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("Error = %v, want error containing 'not yet implemented'", err)
	}
}

func TestNewProvider_NilConfig(t *testing.T) {
	provider, err := NewProvider(nil)
	if err == nil {
		t.Error("NewProvider() expected error for nil config")
	}

	if provider != nil {
		t.Error("NewProvider() should return nil provider on error")
	}

	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Error = %v, want error containing 'cannot be nil'", err)
	}
}
