# User Story: Image Upload with Validation

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As an **event manager**, I want **to upload images for use in templates** so that **I can customize invitations with logos, photos, and graphics**.

---

## Acceptance Criteria

- [ ] Image upload endpoint created
- [ ] File type validation (JPEG, PNG, GIF, WebP, SVG)
- [ ] Magic byte validation (not just extension)
- [ ] File size validation (max 5MB)
- [ ] Image dimension validation (max 4096x4096)
- [ ] EXIF data stripped on upload
- [ ] Unique filename generation
- [ ] Image stored via storage provider
- [ ] Public URL returned
- [ ] All tests pass with timeout
- [ ] Security validation enforced

---

## Technical Details

### Image Upload Service

```go
package storage

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "image"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
    "io"
    "path/filepath"
    "strings"
)

type ImageService interface {
    UploadImage(ctx context.Context, eventID int64, filename string, data io.Reader) (*ImageMetadata, error)
    DeleteImage(ctx context.Context, path string) error
    GetImageURL(ctx context.Context, path string) (string, error)
}

type ImageMetadata struct {
    Path        string `json:"path"`
    PublicURL   string `json:"public_url"`
    Filename    string `json:"filename"`
    ContentType string `json:"content_type"`
    Size        int64  `json:"size"`
    Width       int    `json:"width"`
    Height      int    `json:"height"`
}

type imageService struct {
    provider  Provider
    validator *ImageValidator
}

func NewImageService(provider Provider) ImageService {
    return &imageService{
        provider:  provider,
        validator: NewImageValidator(),
    }
}
```

### Image Validation

```go
type ImageValidator struct {
    maxSize       int64
    maxWidth      int
    maxHeight     int
    allowedTypes  map[string]bool
}

func NewImageValidator() *ImageValidator {
    return &ImageValidator{
        maxSize:   5 * 1024 * 1024, // 5MB
        maxWidth:  4096,
        maxHeight: 4096,
        allowedTypes: map[string]bool{
            "image/jpeg": true,
            "image/png":  true,
            "image/gif":  true,
            "image/webp": true,
        },
    }
}

func (v *ImageValidator) Validate(data []byte) (*ValidationResult, error) {
    if int64(len(data)) > v.maxSize {
        return nil, &ValidationError{
            Field:   "file",
            Message: fmt.Sprintf("Image size exceeds %d bytes", v.maxSize),
        }
    }
    
    contentType := detectContentType(data)
    if !v.allowedTypes[contentType] {
        return nil, &ValidationError{
            Field:   "file",
            Message: "Only JPEG, PNG, GIF, and WebP images are allowed",
        }
    }
    
    config, format, err := image.DecodeConfig(bytes.NewReader(data))
    if err != nil {
        return nil, &ValidationError{
            Field:   "file",
            Message: "File is not a valid image",
        }
    }
    
    if config.Width > v.maxWidth || config.Height > v.maxHeight {
        return nil, &ValidationError{
            Field:   "file",
            Message: fmt.Sprintf("Image dimensions exceed %dx%d pixels", v.maxWidth, v.maxHeight),
        }
    }
    
    return &ValidationResult{
        ContentType: contentType,
        Format:      format,
        Width:       config.Width,
        Height:      config.Height,
        Size:        int64(len(data)),
    }, nil
}

type ValidationResult struct {
    ContentType string
    Format      string
    Width       int
    Height      int
    Size        int64
}
```

### Magic Byte Detection

```go
func detectContentType(data []byte) string {
    if len(data) < 12 {
        return ""
    }
    
    if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
        return "image/jpeg"
    }
    
    if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
        return "image/png"
    }
    
    if bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")) {
        return "image/gif"
    }
    
    if bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
        return "image/webp"
    }
    
    return ""
}
```

### EXIF Stripping

```go
func stripEXIF(data []byte, format string) ([]byte, error) {
    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    
    var buf bytes.Buffer
    
    switch format {
    case "jpeg":
        err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
    case "png":
        err = png.Encode(&buf, img)
    case "gif":
        err = gif.Encode(&buf, img, nil)
    default:
        return data, nil
    }
    
    if err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

### Upload Implementation

```go
func (s *imageService) UploadImage(ctx context.Context, eventID int64, filename string, data io.Reader) (*ImageMetadata, error) {
    imageData, err := io.ReadAll(data)
    if err != nil {
        return nil, fmt.Errorf("failed to read image data: %w", err)
    }
    
    result, err := s.validator.Validate(imageData)
    if err != nil {
        return nil, err
    }
    
    cleanData, err := stripEXIF(imageData, result.Format)
    if err != nil {
        return nil, fmt.Errorf("failed to process image: %w", err)
    }
    
    uniqueFilename := generateUniqueFilename(filename)
    path := fmt.Sprintf("images/%d/%s", eventID, uniqueFilename)
    
    if err := s.provider.PutObject(ctx, path, bytes.NewReader(cleanData), result.ContentType); err != nil {
        return nil, fmt.Errorf("failed to store image: %w", err)
    }
    
    publicURL, err := s.provider.GetPublicURL(ctx, path)
    if err != nil {
        return nil, fmt.Errorf("failed to get public URL: %w", err)
    }
    
    return &ImageMetadata{
        Path:        path,
        PublicURL:   publicURL,
        Filename:    uniqueFilename,
        ContentType: result.ContentType,
        Size:        result.Size,
        Width:       result.Width,
        Height:      result.Height,
    }, nil
}

func generateUniqueFilename(original string) string {
    ext := filepath.Ext(original)
    base := strings.TrimSuffix(filepath.Base(original), ext)
    
    base = sanitizeFilename(base)
    
    randomBytes := make([]byte, 8)
    rand.Read(randomBytes)
    suffix := hex.EncodeToString(randomBytes)
    
    return fmt.Sprintf("%s_%s%s", base, suffix, ext)
}

func sanitizeFilename(name string) string {
    name = strings.Map(func(r rune) rune {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
            return r
        }
        return '_'
    }, name)
    
    if len(name) > 50 {
        name = name[:50]
    }
    
    return name
}
```

---

## API Endpoint

### Upload Image
```
POST /api/events/:event_id/images
Authorization: Required (Event Manager or Admin)
Content-Type: multipart/form-data

Form Data:
- file: (binary image data)

Response 201 Created:
{
    "image": {
        "path": "images/123/logo_a1b2c3d4.png",
        "public_url": "https://rsvp.example.com/assets/images/123/logo_a1b2c3d4.png",
        "filename": "logo_a1b2c3d4.png",
        "content_type": "image/png",
        "size": 45678,
        "width": 800,
        "height": 600
    }
}

Response 400 Bad Request:
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Image size exceeds 5MB",
        "field": "file"
    }
}
```

### Delete Image
```
DELETE /api/events/:event_id/images/:filename
Authorization: Required (Event Manager or Admin)

Response 204 No Content
```

---

## Tasks

### Phase 1: Validation (TDD)
- [ ] Define ImageValidator struct
- [ ] Write test for size validation
- [ ] Write test for dimension validation
- [ ] Write test for type validation (magic bytes)
- [ ] Write test for each allowed format
- [ ] Write test for invalid formats
- [ ] Implement Validate method
- [ ] Implement detectContentType
- [ ] Run tests (should pass)

### Phase 2: EXIF Stripping (TDD)
- [ ] Write test for stripEXIF with JPEG
- [ ] Write test for stripEXIF with PNG
- [ ] Write test for stripEXIF with GIF
- [ ] Write test for stripEXIF with invalid data
- [ ] Implement stripEXIF
- [ ] Run tests (should pass)

### Phase 3: Upload Service (TDD)
- [ ] Define ImageService interface
- [ ] Write test for UploadImage success
- [ ] Write test for UploadImage validation error
- [ ] Write test for UploadImage storage error
- [ ] Write test for generateUniqueFilename
- [ ] Write test for sanitizeFilename
- [ ] Implement UploadImage
- [ ] Implement filename helpers
- [ ] Run tests (should pass)

### Phase 4: Handler Layer (TDD)
- [ ] Create image upload handler
- [ ] Write test for POST handler
- [ ] Write test for multipart parsing
- [ ] Write test for unauthorized access
- [ ] Write test for event ownership check
- [ ] Implement upload handler
- [ ] Run tests (should pass)

### Phase 5: Integration Testing
- [ ] Test full upload flow
- [ ] Test with various image formats
- [ ] Test with oversized images
- [ ] Test with invalid images
- [ ] Test EXIF stripping
- [ ] Test filename uniqueness
- [ ] Test concurrent uploads

---

## Validation Rules

### File Type Validation
- Allowed: JPEG, PNG, GIF, WebP
- Validation: Magic byte detection (not extension)
- Error: "Only JPEG, PNG, GIF, and WebP images are allowed"

### File Size Validation
- Maximum: 5MB (5,242,880 bytes)
- Error: "Image size exceeds 5MB"

### Dimension Validation
- Maximum: 4096x4096 pixels
- Error: "Image dimensions exceed 4096x4096 pixels"

### Image Validity
- Must decode successfully
- Must have valid dimensions
- Error: "File is not a valid image"

### Filename Sanitization
- Alphanumeric, hyphens, underscores only
- Maximum 50 characters (before suffix)
- Random 16-character suffix added
- Original extension preserved

---

## Security Considerations

### EXIF Data Removal
- Strip all metadata
- Prevents location tracking
- Prevents camera information leakage
- Re-encode image to remove EXIF

### Path Traversal Prevention
- Sanitize filenames
- Remove directory separators
- Use filepath.Clean()
- Restrict to designated directories

### Content Type Validation
- Validate by magic bytes
- Ignore file extension
- Prevent polyglot files
- Reject non-image files

---

## Error Handling

| Error Condition | Error Type | HTTP Status | Message |
|----------------|------------|-------------|---------|
| File too large | `ValidationError` | 400 | "Image size exceeds 5MB" |
| Invalid format | `ValidationError` | 400 | "Only JPEG, PNG, GIF, and WebP images are allowed" |
| Dimensions exceeded | `ValidationError` | 400 | "Image dimensions exceed 4096x4096 pixels" |
| Invalid image | `ValidationError` | 400 | "File is not a valid image" |
| Storage error | `InternalError` | 500 | "Failed to store image" |
| Not event owner | `ForbiddenError` | 403 | "You can only upload images for your own events" |

---

## Testing Strategy

### Unit Tests

```go
func TestImageValidator_Validate(t *testing.T) {
    validator := NewImageValidator()
    
    tests := []struct {
        name    string
        data    []byte
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid JPEG",
            data:    loadTestImage(t, "test.jpg"),
            wantErr: false,
        },
        {
            name:    "valid PNG",
            data:    loadTestImage(t, "test.png"),
            wantErr: false,
        },
        {
            name:    "file too large",
            data:    make([]byte, 6*1024*1024),
            wantErr: true,
            errMsg:  "exceeds",
        },
        {
            name:    "invalid format",
            data:    []byte("not an image"),
            wantErr: true,
            errMsg:  "not a valid image",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := validator.Validate(tt.data)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && tt.errMsg != "" {
                if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("Error message = %v, want to contain %v", err, tt.errMsg)
                }
            }
            if !tt.wantErr && result == nil {
                t.Error("Expected validation result")
            }
        })
    }
}

func TestDetectContentType(t *testing.T) {
    tests := []struct {
        name string
        data []byte
        want string
    }{
        {
            name: "JPEG",
            data: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
            want: "image/jpeg",
        },
        {
            name: "PNG",
            data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
            want: "image/png",
        },
        {
            name: "GIF87a",
            data: []byte("GIF87a"),
            want: "image/gif",
        },
        {
            name: "GIF89a",
            data: []byte("GIF89a"),
            want: "image/gif",
        },
        {
            name: "invalid",
            data: []byte("not an image"),
            want: "",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := detectContentType(tt.data)
            if got != tt.want {
                t.Errorf("detectContentType() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestStripEXIF(t *testing.T) {
    originalData := loadTestImage(t, "with_exif.jpg")
    
    stripped, err := stripEXIF(originalData, "jpeg")
    if err != nil {
        t.Fatalf("stripEXIF() error = %v", err)
    }
    
    if len(stripped) >= len(originalData) {
        t.Error("Expected stripped image to be smaller")
    }
    
    img, _, err := image.Decode(bytes.NewReader(stripped))
    if err != nil {
        t.Fatalf("Stripped image is not valid: %v", err)
    }
    
    if img == nil {
        t.Error("Expected valid image after EXIF stripping")
    }
}
```

### Integration Tests

```go
func TestImageService_UploadImage_Integration(t *testing.T) {
    tempDir := t.TempDir()
    provider := NewLocalProvider(tempDir, "http://localhost:8080")
    service := NewImageService(provider)
    
    imageData := loadTestImage(t, "test.jpg")
    
    ctx := context.Background()
    metadata, err := service.UploadImage(ctx, 123, "logo.jpg", bytes.NewReader(imageData))
    if err != nil {
        t.Fatalf("UploadImage() error = %v", err)
    }
    
    if metadata.Path == "" {
        t.Error("Expected path to be set")
    }
    
    if metadata.PublicURL == "" {
        t.Error("Expected public URL to be set")
    }
    
    if metadata.Width == 0 || metadata.Height == 0 {
        t.Error("Expected dimensions to be set")
    }
    
    if !strings.Contains(metadata.Path, "images/123/") {
        t.Errorf("Path = %s, want to contain images/123/", metadata.Path)
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
```

---

## Tasks

### Phase 1: Validation (TDD)
- [ ] Define ImageValidator struct
- [ ] Write test for size validation
- [ ] Write test for type validation
- [ ] Write test for dimension validation
- [ ] Write test for detectContentType
- [ ] Write test for each image format
- [ ] Implement Validate method
- [ ] Implement detectContentType
- [ ] Run tests (should pass)

### Phase 2: EXIF Stripping (TDD)
- [ ] Write test for stripEXIF with JPEG
- [ ] Write test for stripEXIF with PNG
- [ ] Write test for stripEXIF with GIF
- [ ] Write test for stripEXIF preserves image
- [ ] Implement stripEXIF
- [ ] Run tests (should pass)

### Phase 3: Upload Service (TDD)
- [ ] Define ImageService interface
- [ ] Write test for UploadImage success
- [ ] Write test for UploadImage validation errors
- [ ] Write test for generateUniqueFilename
- [ ] Write test for sanitizeFilename
- [ ] Implement UploadImage
- [ ] Implement filename helpers
- [ ] Run tests (should pass)

### Phase 4: Handler Layer (TDD)
- [ ] Create image upload handler
- [ ] Write test for POST handler
- [ ] Write test for multipart form parsing
- [ ] Write test for event ownership check
- [ ] Write test for unauthorized access
- [ ] Implement upload handler
- [ ] Run tests (should pass)

### Phase 5: Integration Testing
- [ ] Test full upload flow
- [ ] Test with each image format
- [ ] Test with oversized images
- [ ] Test with invalid images
- [ ] Test EXIF stripping
- [ ] Test filename uniqueness
- [ ] Test concurrent uploads

---

## Dependencies

**Depends on:**
- Story 08: Storage Provider (for storage interface)
- Epic 01: Auth (for RBAC)

**Blocks:**
- Story 10: Asset Serving (needs uploaded images)
- Story 03: Default Templates (can use uploaded images)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] ImageValidator implemented
- [ ] ImageService implemented
- [ ] Upload handler implemented
- [ ] All unit tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] EXIF stripping verified
- [ ] Security validation verified
- [ ] Documentation updated
- [ ] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11.6 (Image Upload), Section 12 (Asset Storage)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md) - Section 5
- **Story 08:** [06_STORY_08_storage_provider.md](06_STORY_08_storage_provider.md)
