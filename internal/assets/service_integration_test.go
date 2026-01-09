package assets

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/storage"
)

func TestImageUpload_Integration_JPEG(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 800, 600)
	eventID := int64(123)
	filename := "photo.jpg"

	metadata, err := service.UploadImage(ctx, eventID, filename, bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	if metadata.ContentType != "image/jpeg" {
		t.Errorf("ContentType = %v, want image/jpeg", metadata.ContentType)
	}

	if metadata.Width != 800 || metadata.Height != 600 {
		t.Errorf("Dimensions = %dx%d, want 800x600", metadata.Width, metadata.Height)
	}

	if !strings.Contains(metadata.Path, "images/123/") {
		t.Errorf("Path = %v, want to contain images/123/", metadata.Path)
	}

	if !strings.HasSuffix(metadata.Filename, ".jpg") {
		t.Errorf("Filename = %v, want to end with .jpg", metadata.Filename)
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
		t.Errorf("Stored format = %v, want jpeg", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 600 {
		t.Errorf("Stored dimensions = %dx%d, want 800x600", bounds.Dx(), bounds.Dy())
	}
}

func TestImageUpload_Integration_PNG(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestPNG(t, 1024, 768)
	eventID := int64(456)
	filename := "logo.png"

	metadata, err := service.UploadImage(ctx, eventID, filename, bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	if metadata.ContentType != "image/png" {
		t.Errorf("ContentType = %v, want image/png", metadata.ContentType)
	}

	if metadata.Width != 1024 || metadata.Height != 768 {
		t.Errorf("Dimensions = %dx%d, want 1024x768", metadata.Width, metadata.Height)
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

	if format != "png" {
		t.Errorf("Stored format = %v, want png", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 1024 || bounds.Dy() != 768 {
		t.Errorf("Stored dimensions = %dx%d, want 1024x768", bounds.Dx(), bounds.Dy())
	}
}

func TestImageUpload_Integration_GIF(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestGIF(t, 640, 480)
	eventID := int64(789)
	filename := "animation.gif"

	metadata, err := service.UploadImage(ctx, eventID, filename, bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	if metadata.ContentType != "image/gif" {
		t.Errorf("ContentType = %v, want image/gif", metadata.ContentType)
	}

	if metadata.Width != 640 || metadata.Height != 480 {
		t.Errorf("Dimensions = %dx%d, want 640x480", metadata.Width, metadata.Height)
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

	if format != "gif" {
		t.Errorf("Stored format = %v, want gif", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 640 || bounds.Dy() != 480 {
		t.Errorf("Stored dimensions = %dx%d, want 640x480", bounds.Dx(), bounds.Dy())
	}
}

func TestImageUpload_Integration_OversizedImage(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	largeImage := createTestJPEG(t, 5000, 5000)

	_, err := service.UploadImage(ctx, 123, "huge.jpg", bytes.NewReader(largeImage))
	if err == nil {
		t.Error("Expected error for oversized image")
	}

	if !strings.Contains(err.Error(), "dimensions exceed") {
		t.Errorf("Error = %v, want error containing 'dimensions exceed'", err)
	}

	objects, _ := provider.ListObjects(ctx, "images/123/")
	if len(objects) != 0 {
		t.Error("Expected no objects stored for failed upload")
	}
}

func TestImageUpload_Integration_InvalidFormat(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	invalidData := []byte("This is not an image file")

	_, err := service.UploadImage(ctx, 123, "file.txt", bytes.NewReader(invalidData))
	if err == nil {
		t.Error("Expected error for invalid format")
	}

	if !strings.Contains(err.Error(), "allowed") {
		t.Errorf("Error = %v, want error containing 'allowed'", err)
	}

	objects, _ := provider.ListObjects(ctx, "images/123/")
	if len(objects) != 0 {
		t.Error("Expected no objects stored for failed upload")
	}
}

func TestImageUpload_Integration_FilenameUniqueness(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 200, 200)
	eventID := int64(123)
	filename := "logo.jpg"

	metadata1, err := service.UploadImage(ctx, eventID, filename, bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("First upload error = %v", err)
	}

	metadata2, err := service.UploadImage(ctx, eventID, filename, bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("Second upload error = %v", err)
	}

	if metadata1.Filename == metadata2.Filename {
		t.Error("Expected unique filenames for multiple uploads")
	}

	if metadata1.Path == metadata2.Path {
		t.Error("Expected unique paths for multiple uploads")
	}

	objects, err := provider.ListObjects(ctx, "images/123/")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(objects) != 2 {
		t.Errorf("Expected 2 stored objects, got %d", len(objects))
	}
}

func TestImageUpload_Integration_EXIFStripping(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	originalImage := createTestJPEGWithPattern(t, 400, 300)
	originalSize := len(originalImage)

	metadata, err := service.UploadImage(ctx, 123, "photo.jpg", bytes.NewReader(originalImage))
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
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Dimensions = %dx%d, want 400x300", bounds.Dx(), bounds.Dy())
	}

	if len(storedBytes) > originalSize {
		t.Error("Expected EXIF-stripped image to not be larger than original")
	}
}

func TestImageUpload_Integration_MultipleFormats(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()
	eventID := int64(123)

	formats := []struct {
		name        string
		data        []byte
		filename    string
		contentType string
		format      string
	}{
		{
			name:        "JPEG",
			data:        createTestJPEG(t, 200, 200),
			filename:    "test.jpg",
			contentType: "image/jpeg",
			format:      "jpeg",
		},
		{
			name:        "PNG",
			data:        createTestPNG(t, 200, 200),
			filename:    "test.png",
			contentType: "image/png",
			format:      "png",
		},
		{
			name:        "GIF",
			data:        createTestGIF(t, 200, 200),
			filename:    "test.gif",
			contentType: "image/gif",
			format:      "gif",
		},
	}

	for _, fmt := range formats {
		t.Run(fmt.name, func(t *testing.T) {
			metadata, err := service.UploadImage(ctx, eventID, fmt.filename, bytes.NewReader(fmt.data))
			if err != nil {
				t.Fatalf("UploadImage() error = %v", err)
			}

			if metadata.ContentType != fmt.contentType {
				t.Errorf("ContentType = %v, want %v", metadata.ContentType, fmt.contentType)
			}

			storedData, err := provider.GetObject(ctx, metadata.Path)
			if err != nil {
				t.Fatalf("GetObject() error = %v", err)
			}
			defer storedData.Close()

			storedBytes, _ := io.ReadAll(storedData)
			_, storedFormat, err := image.Decode(bytes.NewReader(storedBytes))
			if err != nil {
				t.Fatalf("Stored image is not valid: %v", err)
			}

			if storedFormat != fmt.format {
				t.Errorf("Stored format = %v, want %v", storedFormat, fmt.format)
			}
		})
	}
}

func TestImageUpload_Integration_ConcurrentUploads(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	eventID := int64(123)
	numUploads := 10

	results := make(chan *ImageMetadata, numUploads)
	errors := make(chan error, numUploads)

	for i := 0; i < numUploads; i++ {
		go func(index int) {
			imageData := createTestJPEG(t, 100, 100)
			metadata, err := service.UploadImage(ctx, eventID, "concurrent.jpg", bytes.NewReader(imageData))
			if err != nil {
				errors <- err
				return
			}
			results <- metadata
		}(i)
	}

	var metadataList []*ImageMetadata
	for i := 0; i < numUploads; i++ {
		select {
		case metadata := <-results:
			metadataList = append(metadataList, metadata)
		case err := <-errors:
			t.Fatalf("Concurrent upload error: %v", err)
		}
	}

	if len(metadataList) != numUploads {
		t.Errorf("Expected %d successful uploads, got %d", numUploads, len(metadataList))
	}

	filenameMap := make(map[string]bool)
	for _, metadata := range metadataList {
		if filenameMap[metadata.Filename] {
			t.Errorf("Duplicate filename detected: %v", metadata.Filename)
		}
		filenameMap[metadata.Filename] = true
	}

	objects, err := provider.ListObjects(ctx, "images/123/")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(objects) != numUploads {
		t.Errorf("Expected %d stored objects, got %d", numUploads, len(objects))
	}
}

func TestImageUpload_Integration_DeleteAfterUpload(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 300, 300)
	eventID := int64(123)

	metadata, err := service.UploadImage(ctx, eventID, "temp.jpg", bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	_, err = provider.GetObject(ctx, metadata.Path)
	if err != nil {
		t.Fatalf("GetObject() error = %v, expected image to exist", err)
	}

	err = service.DeleteImage(ctx, metadata.Path)
	if err != nil {
		t.Fatalf("DeleteImage() error = %v", err)
	}

	_, err = provider.GetObject(ctx, metadata.Path)
	if err != storage.ErrNotFound {
		t.Errorf("GetObject() error = %v, want ErrNotFound", err)
	}
}

func TestImageUpload_Integration_MultipleEvents(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 200, 200)

	metadata1, err := service.UploadImage(ctx, 100, "image.jpg", bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("Upload to event 100 error = %v", err)
	}

	metadata2, err := service.UploadImage(ctx, 200, "image.jpg", bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("Upload to event 200 error = %v", err)
	}

	if !strings.Contains(metadata1.Path, "images/100/") {
		t.Errorf("Event 100 path = %v, want to contain images/100/", metadata1.Path)
	}

	if !strings.Contains(metadata2.Path, "images/200/") {
		t.Errorf("Event 200 path = %v, want to contain images/200/", metadata2.Path)
	}

	objects100, _ := provider.ListObjects(ctx, "images/100/")
	if len(objects100) != 1 {
		t.Errorf("Event 100 objects = %d, want 1", len(objects100))
	}

	objects200, _ := provider.ListObjects(ctx, "images/200/")
	if len(objects200) != 1 {
		t.Errorf("Event 200 objects = %d, want 1", len(objects200))
	}
}

func TestImageUpload_Integration_SpecialCharactersInFilename(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 100, 100)
	eventID := int64(123)

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "spaces",
			filename: "my photo.jpg",
		},
		{
			name:     "special characters",
			filename: "photo@#$%.jpg",
		},
		{
			name:     "unicode",
			filename: "фото.jpg",
		},
		{
			name:     "very long name",
			filename: "this_is_a_very_long_filename_that_should_be_truncated_to_prevent_filesystem_issues.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := service.UploadImage(ctx, eventID, tt.filename, bytes.NewReader(imageData))
			if err != nil {
				t.Fatalf("UploadImage() error = %v", err)
			}

			if metadata.Filename == "" {
				t.Error("Expected non-empty sanitized filename")
			}

			if strings.Contains(metadata.Filename, " ") {
				t.Errorf("Filename contains spaces: %v", metadata.Filename)
			}

			if strings.Contains(metadata.Filename, "@") || strings.Contains(metadata.Filename, "#") {
				t.Errorf("Filename contains special characters: %v", metadata.Filename)
			}

			storedData, err := provider.GetObject(ctx, metadata.Path)
			if err != nil {
				t.Fatalf("GetObject() error = %v", err)
			}
			defer storedData.Close()
		})
	}
}

func TestImageUpload_Integration_GetImageURL(t *testing.T) {
	provider := storage.NewMockProvider()
	service := NewImageService(provider)
	ctx := context.Background()

	imageData := createTestJPEG(t, 100, 100)
	eventID := int64(123)

	metadata, err := service.UploadImage(ctx, eventID, "test.jpg", bytes.NewReader(imageData))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}

	url, err := service.GetImageURL(ctx, metadata.Path)
	if err != nil {
		t.Fatalf("GetImageURL() error = %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	if !strings.Contains(url, metadata.Path) {
		t.Errorf("URL = %v, want to contain %v", url, metadata.Path)
	}

	if url != metadata.PublicURL {
		t.Errorf("GetImageURL() = %v, want %v", url, metadata.PublicURL)
	}
}

func createTestJPEGWithPattern(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(((x + y) * 255) / (width + height))
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("Failed to create test JPEG: %v", err)
	}
	return buf.Bytes()
}
