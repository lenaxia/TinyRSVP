package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
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

	objectData, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	m.objects[path] = objectData
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
