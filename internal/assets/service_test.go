package assets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/storage"
)

func TestImageService_UploadImage_Success(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 800, 600)
	eventID := int64(123)
	filename := "logo.jpg"

	metadata, err := service.UploadImage(ctx, eventID, filename, bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	if metadata.Path == "" {
		t.Error("Expected path to be set")
	}

	if !strings.Contains(metadata.Path, fmt.Sprintf("images/%d/", eventID)) {
		t.Errorf("Path = %v, want to contain images/%d/", metadata.Path, eventID)
	}

	if metadata.PublicURL == "" {
		t.Error("Expected public URL to be set")
	}

	if metadata.Filename == "" {
		t.Error("Expected filename to be set")
	}

	if metadata.ContentType != "image/jpeg" {
		t.Errorf("ContentType = %v, want image/jpeg", metadata.ContentType)
	}

	if metadata.Width != 800 {
		t.Errorf("Width = %v, want 800", metadata.Width)
	}

	if metadata.Height != 600 {
		t.Errorf("Height = %v, want 600", metadata.Height)
	}

	if metadata.Size == 0 {
		t.Error("Expected size to be set")
	}

	storedData, err := provider.GetObject(ctx, metadata.Path)
	if err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}
	defer storedData.Close()

	storedBytes, _ := io.ReadAll(storedData)
	if len(storedBytes) == 0 {
		t.Error("Expected stored image data")
	}
}

func TestImageService_UploadImage_ValidationError_Size(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	largeData := make([]byte, 6*1024*1024)
	copy(largeData, []byte{0xFF, 0xD8, 0xFF})

	_, err := service.UploadImage(ctx, 123, "large.jpg", bytes.NewReader(largeData))
	if err == nil {
		t.Error("UploadImage() expected error for oversized image")
	}

	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("Error = %v, want error containing 'exceeds'", err)
	}
}

func TestImageService_UploadImage_ValidationError_InvalidFormat(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	invalidData := []byte("This is not an image file")

	_, err := service.UploadImage(ctx, 123, "file.txt", bytes.NewReader(invalidData))
	if err == nil {
		t.Error("UploadImage() expected error for invalid format")
	}

	if !strings.Contains(err.Error(), "allowed") {
		t.Errorf("Error = %v, want error containing 'allowed'", err)
	}
}

func TestImageService_UploadImage_ValidationError_Dimensions(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	oversizedImage := createTestJPEG(t, 5000, 5000)

	_, err := service.UploadImage(ctx, 123, "huge.jpg", bytes.NewReader(oversizedImage))
	if err == nil {
		t.Error("UploadImage() expected error for oversized dimensions")
	}

	if !strings.Contains(err.Error(), "dimensions exceed") {
		t.Errorf("Error = %v, want error containing 'dimensions exceed'", err)
	}
}

func TestImageService_UploadImage_StorageError(t *testing.T) {
	provider := storage.NewMockProvider()
	provider.PutObjectFunc = func(ctx context.Context, path string, data io.Reader, contentType string) error {
		return errors.New("storage failure")
	}

	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 100, 100)

	_, err := service.UploadImage(ctx, 123, "logo.jpg", bytes.NewReader(imageData))
	if err == nil {
		t.Error("UploadImage() expected error for storage failure")
	}

	if !strings.Contains(err.Error(), "storage") {
		t.Errorf("Error = %v, want error containing 'storage'", err)
	}
}

func TestImageService_UploadImage_ReadError(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	errorReader := &errorReader{err: errors.New("read failure")}

	_, err := service.UploadImage(ctx, 123, "logo.jpg", errorReader)
	if err == nil {
		t.Error("UploadImage() expected error for read failure")
	}

	if !strings.Contains(err.Error(), "read") {
		t.Errorf("Error = %v, want error containing 'read'", err)
	}
}

func TestImageService_UploadImage_UniqueFilenames(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 100, 100)
	filename := "logo.jpg"
	eventID := int64(123)

	metadata1, err := service.UploadImage(ctx, eventID, filename, bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	metadata2, err := service.UploadImage(ctx, eventID, filename, bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	if metadata1.Filename == metadata2.Filename {
		t.Error("Expected unique filenames for multiple uploads")
	}

	if metadata1.Path == metadata2.Path {
		t.Error("Expected unique paths for multiple uploads")
	}
}

func TestImageService_UploadImage_EXIFStripped(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 200, 150)
	originalSize := len(imageData)

	metadata, err := service.UploadImage(ctx, 123, "photo.jpg", bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	storedData, err := provider.GetObject(ctx, metadata.Path)
	if err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}
	defer storedData.Close()

	storedBytes, _ := io.ReadAll(storedData)

	img, format, err := image.Decode(bytes.NewReader(storedBytes))
	if err != nil {
		t.Fatalf("Stored image is not valid: %v", err)
	}

	if format != "jpeg" {
		t.Errorf("Format = %v, want jpeg", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("Dimensions = %dx%d, want 200x150", bounds.Dx(), bounds.Dy())
	}

	if len(storedBytes) > originalSize {
		t.Error("Expected EXIF-stripped image to not be larger than original")
	}
}

func TestImageService_DeleteImage(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	path := "images/123/logo.jpg"
	provider.PutObject(ctx, path, bytes.NewReader([]byte("data")), "image/jpeg")

	err := service.DeleteImage(ctx, path)
	if err != nil {
		t.Fatalf("DeleteImage() error = %v", err)
	}

	_, err = provider.GetObject(ctx, path)
	if err != storage.ErrNotFound {
		t.Errorf("GetObject() error = %v, want ErrNotFound", err)
	}
}

func TestImageService_GetImageURL(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	path := "images/123/logo.jpg"

	url, err := service.GetImageURL(ctx, path)
	if err != nil {
		t.Fatalf("GetImageURL() error = %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	if !strings.Contains(url, path) {
		t.Errorf("URL = %v, want to contain %v", url, path)
	}
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
