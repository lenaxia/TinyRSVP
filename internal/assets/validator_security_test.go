package assets

import (
	"bytes"
	"image"
	"strings"
	"testing"
)

func TestDetectContentType_MaliciousFiles(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name: "SVG file - should be rejected",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <circle cx="50" cy="50" r="40" fill="red"/>
</svg>`),
			want: "",
		},
		{
			name: "SVG with embedded script - XSS vector",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <script>alert('XSS')</script>
  <circle cx="50" cy="50" r="40" fill="red"/>
</svg>`),
			want: "",
		},
		{
			name: "HTML file pretending to be image",
			data: []byte(`<!DOCTYPE html>
<html><body><img src="x" onerror="alert('XSS')"></body></html>`),
			want: "",
		},
		{
			name: "JavaScript file",
			data: []byte(`alert('malicious code');`),
			want: "",
		},
		{
			name: "PHP file",
			data: []byte(`<?php system($_GET['cmd']); ?>`),
			want: "",
		},
		{
			name: "Executable (ELF)",
			data: []byte{0x7F, 0x45, 0x4C, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
			want: "",
		},
		{
			name: "Windows executable (PE)",
			data: []byte{0x4D, 0x5A, 0x90, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00},
			want: "",
		},
		{
			name: "Mach-O executable",
			data: []byte{0xFE, 0xED, 0xFA, 0xCE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
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

func TestImageValidator_Validate_RenamedExecutables(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ELF executable renamed as .jpg",
			data:    []byte{0x7F, 0x45, 0x4C, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
			errMsg:  "allowed",
		},
		{
			name:    "PE executable renamed as .png",
			data:    []byte{0x4D, 0x5A, 0x90, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00},
			wantErr: true,
			errMsg:  "allowed",
		},
		{
			name:    "Shell script renamed as .gif",
			data:    []byte("#!/bin/bash\nrm -rf /"),
			wantErr: true,
			errMsg:  "allowed",
		},
		{
			name:    "ZIP file renamed as .jpg",
			data:    []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00},
			wantErr: true,
			errMsg:  "allowed",
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

func TestImageValidator_Validate_PolyglotFiles(t *testing.T) {
	validator := NewImageValidator()

	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	scriptPayload := []byte(`<script>alert('XSS')</script>`)
	
	polyglot := append(jpegHeader, scriptPayload...)
	
	_, err := validator.Validate(polyglot)
	if err == nil {
		t.Error("Expected validation to fail for polyglot file with script payload")
	}
}

func TestImageValidator_Validate_EmbeddedScripts(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name: "image with <script> tag",
			data: func() []byte {
				img := createTestJPEG(t, 100, 100)
				return append(img, []byte("<script>alert('xss')</script>")...)
			}(),
			wantErr: true,
		},
		{
			name: "image with javascript: protocol",
			data: func() []byte {
				img := createTestPNG(t, 100, 100)
				return append(img, []byte("javascript:alert(1)")...)
			}(),
			wantErr: true,
		},
		{
			name: "image with data: protocol",
			data: func() []byte {
				img := createTestGIF(t, 100, 100)
				return append(img, []byte("data:text/html,<script>alert(1)</script>")...)
			}(),
			wantErr: true,
		},
		{
			name: "image with onerror handler",
			data: func() []byte {
				img := createTestJPEG(t, 100, 100)
				return append(img, []byte(`onerror="alert(1)"`)...)
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Validate(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageValidator_Validate_SVGRejection(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "basic SVG",
			data: []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><circle cx="50" cy="50" r="40"/></svg>`),
		},
		{
			name: "SVG with script",
			data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
		},
		{
			name: "SVG with event handler",
			data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle onload="alert(1)"/></svg>`),
		},
		{
			name: "SVG with external resource",
			data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"><image href="http://evil.com/malware.exe"/></svg>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Validate(tt.data)
			if err == nil {
				t.Error("Expected SVG to be rejected")
			}
			if err != nil && !strings.Contains(err.Error(), "allowed") {
				t.Errorf("Expected 'allowed' error, got: %v", err)
			}
		})
	}
}

func TestImageValidator_Validate_MismatchedExtension(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name        string
		data        []byte
		description string
		wantErr     bool
	}{
		{
			name:        "JPEG data is valid",
			data:        createTestJPEG(t, 100, 100),
			description: "Valid JPEG should pass",
			wantErr:     false,
		},
		{
			name:        "PNG data is valid",
			data:        createTestPNG(t, 100, 100),
			description: "Valid PNG should pass",
			wantErr:     false,
		},
		{
			name:        "Text file with .jpg extension",
			data:        []byte("This is just text, not an image"),
			description: "Text file should be rejected regardless of extension",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.Validate(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v (%s)", err, tt.wantErr, tt.description)
			}
		})
	}
}

func TestImageValidator_Validate_BoundaryConditions(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name    string
		size    int
		width   int
		height  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "exactly at size limit",
			size:    5 * 1024 * 1024,
			width:   100,
			height:  100,
			wantErr: false,
		},
		{
			name:    "one byte over size limit",
			size:    5*1024*1024 + 1,
			width:   100,
			height:  100,
			wantErr: true,
			errMsg:  "exceeds",
		},
		{
			name:    "exactly at dimension limit",
			size:    1024 * 1024,
			width:   4096,
			height:  4096,
			wantErr: false,
		},
		{
			name:    "one pixel over width limit",
			size:    1024 * 1024,
			width:   4097,
			height:  100,
			wantErr: true,
			errMsg:  "dimensions exceed",
		},
		{
			name:    "one pixel over height limit",
			size:    1024 * 1024,
			width:   100,
			height:  4097,
			wantErr: true,
			errMsg:  "dimensions exceed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createTestJPEG(t, tt.width, tt.height)
			
			if len(data) < tt.size {
				padding := make([]byte, tt.size-len(data))
				data = append(data, padding...)
			} else if len(data) > tt.size {
				data = data[:tt.size]
			}

			_, err := validator.Validate(data)
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

func TestStripEXIF_PreservesImageQuality(t *testing.T) {
	tests := []struct {
		name   string
		format string
		width  int
		height int
	}{
		{
			name:   "JPEG quality preservation",
			format: "jpeg",
			width:  800,
			height: 600,
		},
		{
			name:   "PNG quality preservation",
			format: "png",
			width:  800,
			height: 600,
		},
		{
			name:   "GIF quality preservation",
			format: "gif",
			width:  800,
			height: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var originalData []byte
			switch tt.format {
			case "jpeg":
				originalData = createTestJPEG(t, tt.width, tt.height)
			case "png":
				originalData = createTestPNG(t, tt.width, tt.height)
			case "gif":
				originalData = createTestGIF(t, tt.width, tt.height)
			}

			stripped, err := stripEXIF(originalData, tt.format)
			if err != nil {
				t.Fatalf("stripEXIF() error = %v", err)
			}

			if len(stripped) == 0 {
				t.Error("Stripped image has zero size")
			}

			img, format, err := image.Decode(bytes.NewReader(stripped))
			if err != nil {
				t.Fatalf("Failed to decode stripped image: %v", err)
			}

			if img == nil {
				t.Error("Stripped image is nil")
			}

			if format != tt.format {
				t.Errorf("Format = %v, want %v", format, tt.format)
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
				t.Errorf("Dimensions = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.width, tt.height)
			}
		})
	}
}

func TestDetectContentType_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "empty data",
			data: []byte{},
			want: "",
		},
		{
			name: "very short data",
			data: []byte{0xFF},
			want: "",
		},
		{
			name: "null bytes",
			data: make([]byte, 20),
			want: "",
		},
		{
			name: "random bytes",
			data: []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, 0x11, 0x22, 0x33, 0x44},
			want: "",
		},
		{
			name: "partial JPEG header",
			data: []byte{0xFF, 0xD8},
			want: "",
		},
		{
			name: "partial PNG header",
			data: []byte{0x89, 0x50, 0x4E, 0x47},
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
