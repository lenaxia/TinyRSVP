package storage

import (
	"context"
	"fmt"
	"io"
	"time"
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
