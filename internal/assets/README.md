# Assets Package

## Purpose
Handles image upload, validation, and management for event templates.

## Rules
- All images must pass validation (size, dimensions, format)
- EXIF data is automatically stripped for privacy
- Unique filenames are generated using UUID suffix
- Only JPEG, PNG, GIF, and WebP formats are supported

## Structure
- `validator.go` - Image validation (magic bytes, size, dimensions)
- `validator_test.go` - Validator unit tests
- `service.go` - ImageService for upload/delete/URL operations
- `service_test.go` - Service unit tests
- `service_integration_test.go` - End-to-end integration tests

## Key Components

### ImageValidator
Validates images before upload:
- **Size**: Max 5MB
- **Dimensions**: Max 4096x4096 pixels
- **Format**: JPEG, PNG, GIF, WebP (validated by magic bytes)

### ImageService
Manages image lifecycle:
- `UploadImage()` - Validates, strips EXIF, generates unique filename, stores via provider
- `DeleteImage()` - Removes image from storage
- `GetImageURL()` - Returns public URL for image

## Dependencies
- `internal/storage` - Storage provider interface
- `internal/models` - Error types

## Usage Example
```go
provider := storage.NewMockProvider()
service := assets.NewImageService(provider)

metadata, err := service.UploadImage(ctx, eventID, "logo.jpg", fileReader)
if err != nil {
    // Handle validation or storage error
}

// Use metadata.PublicURL in templates
```

## Security Features
- Magic byte validation (not just file extension)
- EXIF data stripping (removes location/camera metadata)
- Filename sanitization (prevents path traversal)
- Size and dimension limits (prevents DoS)
