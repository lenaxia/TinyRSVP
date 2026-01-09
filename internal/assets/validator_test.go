package assets

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "JPEG magic bytes",
			data: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			want: "image/jpeg",
		},
		{
			name: "PNG magic bytes",
			data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			want: "image/png",
		},
		{
			name: "GIF87a magic bytes",
			data: append([]byte("GIF87a"), make([]byte, 6)...),
			want: "image/gif",
		},
		{
			name: "GIF89a magic bytes",
			data: append([]byte("GIF89a"), make([]byte, 6)...),
			want: "image/gif",
		},
		{
			name: "WebP magic bytes",
			data: append([]byte("RIFF"), append([]byte{0x00, 0x00, 0x00, 0x00}, []byte("WEBP")...)...),
			want: "image/webp",
		},
		{
			name: "invalid - too short",
			data: []byte{0xFF, 0xD8},
			want: "",
		},
		{
			name: "invalid - not an image",
			data: []byte("This is not an image file at all"),
			want: "",
		},
		{
			name: "invalid - empty",
			data: []byte{},
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

func TestImageValidator_Validate_Size(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name    string
		size    int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid size - 1MB",
			size:    1 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "valid size - exactly 5MB",
			size:    5 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "invalid size - exceeds 5MB",
			size:    5*1024*1024 + 1,
			wantErr: true,
			errMsg:  "exceeds",
		},
		{
			name:    "invalid size - 10MB",
			size:    10 * 1024 * 1024,
			wantErr: true,
			errMsg:  "exceeds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createTestJPEG(t, 100, 100)
			if len(data) < tt.size {
				padding := make([]byte, tt.size-len(data))
				data = append(data, padding...)
			} else {
				data = data[:tt.size]
			}

			_, err := validator.Validate(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestImageValidator_Validate_ContentType(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid JPEG",
			data:    createTestJPEG(t, 100, 100),
			wantErr: false,
		},
		{
			name:    "valid PNG",
			data:    createTestPNG(t, 100, 100),
			wantErr: false,
		},
		{
			name:    "valid GIF",
			data:    createTestGIF(t, 100, 100),
			wantErr: false,
		},
		{
			name:    "invalid - text file",
			data:    []byte("This is a text file, not an image"),
			wantErr: true,
			errMsg:  "allowed",
		},
		{
			name:    "invalid - empty",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Validate(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestImageValidator_Validate_Dimensions(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid dimensions - 100x100",
			width:   100,
			height:  100,
			wantErr: false,
		},
		{
			name:    "valid dimensions - 4096x4096",
			width:   4096,
			height:  4096,
			wantErr: false,
		},
		{
			name:    "invalid dimensions - width exceeds",
			width:   4097,
			height:  100,
			wantErr: true,
			errMsg:  "dimensions exceed",
		},
		{
			name:    "invalid dimensions - height exceeds",
			width:   100,
			height:  4097,
			wantErr: true,
			errMsg:  "dimensions exceed",
		},
		{
			name:    "invalid dimensions - both exceed",
			width:   5000,
			height:  5000,
			wantErr: true,
			errMsg:  "dimensions exceed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createTestJPEG(t, tt.width, tt.height)

			result, err := validator.Validate(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr {
				if result.Width != tt.width {
					t.Errorf("Width = %v, want %v", result.Width, tt.width)
				}
				if result.Height != tt.height {
					t.Errorf("Height = %v, want %v", result.Height, tt.height)
				}
			}
		})
	}
}

func TestImageValidator_Validate_Complete(t *testing.T) {
	validator := NewImageValidator()

	data := createTestJPEG(t, 800, 600)

	result, err := validator.Validate(data)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.ContentType != "image/jpeg" {
		t.Errorf("ContentType = %v, want image/jpeg", result.ContentType)
	}
	if result.Format != "jpeg" {
		t.Errorf("Format = %v, want jpeg", result.Format)
	}
	if result.Width != 800 {
		t.Errorf("Width = %v, want 800", result.Width)
	}
	if result.Height != 600 {
		t.Errorf("Height = %v, want 600", result.Height)
	}
	if result.Size != int64(len(data)) {
		t.Errorf("Size = %v, want %v", result.Size, len(data))
	}
}

func TestStripEXIF_JPEG(t *testing.T) {
	originalData := createTestJPEG(t, 200, 150)

	stripped, err := stripEXIF(originalData, "jpeg")
	if err != nil {
		t.Fatalf("stripEXIF() error = %v", err)
	}

	if len(stripped) == 0 {
		t.Error("Expected non-empty stripped image")
	}

	img, format, err := image.Decode(bytes.NewReader(stripped))
	if err != nil {
		t.Fatalf("Stripped image is not valid: %v", err)
	}

	if format != "jpeg" {
		t.Errorf("Format = %v, want jpeg", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("Dimensions = %dx%d, want 200x150", bounds.Dx(), bounds.Dy())
	}
}

func TestStripEXIF_PNG(t *testing.T) {
	originalData := createTestPNG(t, 200, 150)

	stripped, err := stripEXIF(originalData, "png")
	if err != nil {
		t.Fatalf("stripEXIF() error = %v", err)
	}

	if len(stripped) == 0 {
		t.Error("Expected non-empty stripped image")
	}

	img, format, err := image.Decode(bytes.NewReader(stripped))
	if err != nil {
		t.Fatalf("Stripped image is not valid: %v", err)
	}

	if format != "png" {
		t.Errorf("Format = %v, want png", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("Dimensions = %dx%d, want 200x150", bounds.Dx(), bounds.Dy())
	}
}

func TestStripEXIF_GIF(t *testing.T) {
	originalData := createTestGIF(t, 200, 150)

	stripped, err := stripEXIF(originalData, "gif")
	if err != nil {
		t.Fatalf("stripEXIF() error = %v", err)
	}

	if len(stripped) == 0 {
		t.Error("Expected non-empty stripped image")
	}

	img, format, err := image.Decode(bytes.NewReader(stripped))
	if err != nil {
		t.Fatalf("Stripped image is not valid: %v", err)
	}

	if format != "gif" {
		t.Errorf("Format = %v, want gif", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("Dimensions = %dx%d, want 200x150", bounds.Dx(), bounds.Dy())
	}
}

func TestStripEXIF_InvalidData(t *testing.T) {
	invalidData := []byte("not an image")

	_, err := stripEXIF(invalidData, "jpeg")
	if err == nil {
		t.Error("Expected error for invalid image data")
	}
}

func TestStripEXIF_UnsupportedFormat(t *testing.T) {
	data := createTestJPEG(t, 100, 100)

	result, err := stripEXIF(data, "webp")
	if err != nil {
		t.Fatalf("stripEXIF() error = %v", err)
	}

	if !bytes.Equal(result, data) {
		t.Error("Expected original data for unsupported format")
	}
}

func TestGenerateUniqueFilename(t *testing.T) {
	tests := []struct {
		name     string
		original string
	}{
		{
			name:     "simple filename",
			original: "photo.jpg",
		},
		{
			name:     "filename with spaces",
			original: "my photo.png",
		},
		{
			name:     "filename with special chars",
			original: "photo@#$%.gif",
		},
		{
			name:     "long filename",
			original: "this_is_a_very_long_filename_that_exceeds_fifty_characters_in_length.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateUniqueFilename(tt.original)

			if result == "" {
				t.Error("Expected non-empty filename")
			}

			if !strings.HasSuffix(result, ".jpg") && !strings.HasSuffix(result, ".png") && !strings.HasSuffix(result, ".gif") {
				t.Errorf("Expected filename to preserve extension, got %v", result)
			}

			if len(result) > 100 {
				t.Errorf("Filename too long: %d characters", len(result))
			}

			result2 := generateUniqueFilename(tt.original)
			if result == result2 {
				t.Error("Expected unique filenames for multiple calls")
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "alphanumeric only",
			input: "photo123",
			want:  "photo123",
		},
		{
			name:  "with hyphens and underscores",
			input: "my-photo_2024",
			want:  "my-photo_2024",
		},
		{
			name:  "with spaces",
			input: "my photo",
			want:  "my_photo",
		},
		{
			name:  "with special characters",
			input: "photo@#$%^&*()",
			want:  "photo_________",
		},
		{
			name:  "mixed case",
			input: "MyPhoto",
			want:  "MyPhoto",
		},
		{
			name:  "exceeds 50 chars",
			input: "this_is_a_very_long_filename_that_definitely_exceeds_the_fifty_character_limit",
			want:  "this_is_a_very_long_filename_that_definitely_excee",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createTestJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("Failed to create test JPEG: %v", err)
	}
	return buf.Bytes()
}

func createTestPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to create test PNG: %v", err)
	}
	return buf.Bytes()
}

func createTestGIF(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewPaletted(image.Rect(0, 0, width, height), color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
	})

	var buf bytes.Buffer
	if err := gif.Encode(&buf, img, nil); err != nil {
		t.Fatalf("Failed to create test GIF: %v", err)
	}
	return buf.Bytes()
}
