package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid simple path",
			path:    "images/test.png",
			wantErr: false,
		},
		{
			name:    "valid nested path",
			path:    "images/123/logo.png",
			wantErr: false,
		},
		{
			name:    "valid path with multiple levels",
			path:    "templates/789/custom/file.html",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path traversal with double dots",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path traversal in middle",
			path:    "images/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path with single dot",
			path:    "images/./test.png",
			wantErr: false,
		},
		{
			name:    "absolute path",
			path:    "/etc/passwd",
			wantErr: false,
		},
		{
			name:    "path with trailing slash",
			path:    "images/123/",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectContentTypeFromPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "jpeg extension",
			path: "test.jpg",
			want: "image/jpeg",
		},
		{
			name: "jpeg uppercase",
			path: "test.JPEG",
			want: "image/jpeg",
		},
		{
			name: "png extension",
			path: "test.png",
			want: "image/png",
		},
		{
			name: "gif extension",
			path: "test.gif",
			want: "image/gif",
		},
		{
			name: "webp extension",
			path: "test.webp",
			want: "image/webp",
		},
		{
			name: "html extension",
			path: "test.html",
			want: "text/html",
		},
		{
			name: "txt extension",
			path: "test.txt",
			want: "text/plain",
		},
		{
			name: "css extension",
			path: "test.css",
			want: "text/css",
		},
		{
			name: "unknown extension",
			path: "test.xyz",
			want: "application/octet-stream",
		},
		{
			name: "no extension",
			path: "test",
			want: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectContentTypeFromPath(tt.path)
			if got != tt.want {
				t.Errorf("detectContentTypeFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLocalProvider(t *testing.T) {
	basePath := "/data/uploads"
	baseURL := "http://localhost:8080"

	provider := NewLocalProvider(basePath, baseURL)
	if provider == nil {
		t.Fatal("NewLocalProvider() returned nil")
	}

	localProv, ok := provider.(*localProvider)
	if !ok {
		t.Fatal("NewLocalProvider() did not return *localProvider")
	}

	if localProv.basePath != basePath {
		t.Errorf("basePath = %v, want %v", localProv.basePath, basePath)
	}

	if localProv.baseURL != baseURL {
		t.Errorf("baseURL = %v, want %v", localProv.baseURL, baseURL)
	}
}

func TestLocalProvider_PutObject(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	tests := []struct {
		name        string
		path        string
		data        []byte
		contentType string
		wantErr     bool
		checkFile   bool
	}{
		{
			name:        "valid put simple path",
			path:        "test.txt",
			data:        []byte("test data"),
			contentType: "text/plain",
			wantErr:     false,
			checkFile:   true,
		},
		{
			name:        "valid put nested path",
			path:        "images/123/logo.png",
			data:        []byte("image data"),
			contentType: "image/png",
			wantErr:     false,
			checkFile:   true,
		},
		{
			name:        "overwrite existing file",
			path:        "overwrite.txt",
			data:        []byte("new data"),
			contentType: "text/plain",
			wantErr:     false,
			checkFile:   true,
		},
		{
			name:        "path traversal attempt",
			path:        "../../../etc/passwd",
			data:        []byte("malicious"),
			contentType: "text/plain",
			wantErr:     true,
			checkFile:   false,
		},
		{
			name:        "empty path",
			path:        "",
			data:        []byte("data"),
			contentType: "text/plain",
			wantErr:     true,
			checkFile:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.PutObject(ctx, tt.path, bytes.NewReader(tt.data), tt.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("PutObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFile && !tt.wantErr {
				fullPath := filepath.Join(tempDir, tt.path)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Error("File was not created")
				}

				content, err := os.ReadFile(fullPath)
				if err != nil {
					t.Fatalf("Failed to read file: %v", err)
				}

				if !bytes.Equal(content, tt.data) {
					t.Errorf("File content = %v, want %v", content, tt.data)
				}

				info, err := os.Stat(fullPath)
				if err != nil {
					t.Fatalf("Failed to stat file: %v", err)
				}

				if info.Mode().Perm() != 0644 {
					t.Errorf("File permissions = %o, want 0644", info.Mode().Perm())
				}
			}
		})
	}

	t.Run("overwrite test setup", func(t *testing.T) {
		path := "overwrite.txt"
		originalData := []byte("original data")
		err := provider.PutObject(ctx, path, bytes.NewReader(originalData), "text/plain")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		newData := []byte("new data")
		err = provider.PutObject(ctx, path, bytes.NewReader(newData), "text/plain")
		if err != nil {
			t.Fatalf("Overwrite failed: %v", err)
		}

		fullPath := filepath.Join(tempDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if !bytes.Equal(content, newData) {
			t.Errorf("File was not overwritten correctly")
		}
	})
}

func TestLocalProvider_GetObject(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	testPath := "test/file.txt"
	testData := []byte("test data")

	err := provider.PutObject(ctx, testPath, bytes.NewReader(testData), "text/plain")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		wantNil bool
	}{
		{
			name:    "get existing file",
			path:    testPath,
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "get nonexistent file",
			path:    "nonexistent.txt",
			wantErr: true,
			wantNil: true,
		},
		{
			name:    "path traversal attempt",
			path:    "../../../etc/passwd",
			wantErr: true,
			wantNil: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := provider.GetObject(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantNil && reader != nil {
				reader.Close()
				t.Error("GetObject() returned non-nil reader for error case")
				return
			}

			if !tt.wantErr && reader != nil {
				defer reader.Close()
				retrieved, err := io.ReadAll(reader)
				if err != nil {
					t.Fatalf("Failed to read data: %v", err)
				}

				if !bytes.Equal(retrieved, testData) {
					t.Errorf("Retrieved data = %v, want %v", retrieved, testData)
				}
			}
		})
	}
}

func TestLocalProvider_DeleteObject(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	testPath := "test/file.txt"
	err := provider.PutObject(ctx, testPath, bytes.NewReader([]byte("data")), "text/plain")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "delete existing file",
			path:    testPath,
			wantErr: false,
		},
		{
			name:    "delete nonexistent file (idempotent)",
			path:    "nonexistent.txt",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.DeleteObject(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.path == testPath {
				fullPath := filepath.Join(tempDir, tt.path)
				if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
					t.Error("File was not deleted")
				}
			}
		})
	}
}

func TestLocalProvider_GetPublicURL(t *testing.T) {
	baseURL := "http://localhost:8080"
	provider := NewLocalProvider("/data/uploads", baseURL)
	ctx := context.Background()

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "simple path",
			path:    "test.png",
			want:    "http://localhost:8080/assets/test.png",
			wantErr: false,
		},
		{
			name:    "nested path",
			path:    "images/123/logo.png",
			want:    "http://localhost:8080/assets/images/123/logo.png",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    "../../../etc/passwd",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.GetPublicURL(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("GetPublicURL() = %v, want %v", got, tt.want)
			}
		})
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
		"templates/789/custom.html",
	}

	for _, path := range paths {
		err := provider.PutObject(ctx, path, bytes.NewReader([]byte("data")), "application/octet-stream")
		if err != nil {
			t.Fatalf("Setup failed for %s: %v", path, err)
		}
	}

	tests := []struct {
		name      string
		prefix    string
		wantCount int
		wantErr   bool
		checkPaths []string
	}{
		{
			name:      "list with prefix images/123/",
			prefix:    "images/123/",
			wantCount: 2,
			wantErr:   false,
			checkPaths: []string{"images/123/logo.png", "images/123/banner.jpg"},
		},
		{
			name:      "list with prefix images/",
			prefix:    "images/",
			wantCount: 3,
			wantErr:   false,
			checkPaths: []string{"images/123/logo.png", "images/123/banner.jpg", "images/456/photo.png"},
		},
		{
			name:      "list with prefix templates/",
			prefix:    "templates/",
			wantCount: 1,
			wantErr:   false,
			checkPaths: []string{"templates/789/custom.html"},
		},
		{
			name:      "list with empty prefix",
			prefix:    "",
			wantCount: 4,
			wantErr:   false,
		},
		{
			name:      "list with nonexistent prefix",
			prefix:    "nonexistent/",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "path traversal attempt",
			prefix:    "../../../etc/",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects, err := provider.ListObjects(ctx, tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListObjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(objects) != tt.wantCount {
					t.Errorf("ListObjects() returned %d objects, want %d", len(objects), tt.wantCount)
				}

				if tt.checkPaths != nil {
					foundPaths := make(map[string]bool)
					for _, obj := range objects {
						foundPaths[obj.Path] = true
					}

					for _, expectedPath := range tt.checkPaths {
						if !foundPaths[expectedPath] {
							t.Errorf("Expected path %s not found in results", expectedPath)
						}
					}
				}

				for i := 1; i < len(objects); i++ {
					if objects[i-1].Path >= objects[i].Path {
						t.Errorf("Objects not sorted: %s >= %s", objects[i-1].Path, objects[i].Path)
					}
				}

				for _, obj := range objects {
					if !strings.HasPrefix(obj.Path, tt.prefix) {
						t.Errorf("Object path %s does not have prefix %s", obj.Path, tt.prefix)
					}

					if obj.Size <= 0 {
						t.Errorf("Object %s has invalid size: %d", obj.Path, obj.Size)
					}

					if obj.ContentType == "" {
						t.Errorf("Object %s has empty content type", obj.Path)
					}

					if obj.LastModified.IsZero() {
						t.Errorf("Object %s has zero LastModified time", obj.Path)
					}
				}
			}
		})
	}
}
