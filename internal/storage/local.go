package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	if prefix != "" {
		if err := validatePath(prefix); err != nil {
			return nil, err
		}
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
