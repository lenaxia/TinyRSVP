# Domain 6: Template & Asset Management - Low-Level Design

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Implementation Ready  
**HLD Reference:** [Section 11 - Templates](../02_REVISED_HLD.md#11-templates--customization), [Section 12 - Asset Storage](../02_REVISED_HLD.md#12-asset-storage)

---

## 1. Overview

### 1.1 Purpose

Manages templates for emails and web pages, asset storage (images), and template rendering with XSS prevention.

### 1.2 Responsibilities

- Template CRUD operations
- Template validation and security (XSS prevention)
- Go html/template integration
- Template variable interpolation
- Image upload and validation
- Storage provider abstraction (local FS, future S3)
- Asset access control
- Asset deletion policy

### 1.3 Design Principles

- **XSS Prevention** - Auto-escaping via html/template
- **Validated** - Parse templates before saving
- **Pluggable Storage** - Abstract storage provider
- **Type-Safe** - Strongly-typed template data
- **Secure by Default** - Sanitize all inputs

---

## 2. Package Structure

```
internal/
├── templates/
│   ├── service.go              # Template service
│   ├── service_test.go
│   ├── renderer.go             # Template rendering
│   ├── renderer_test.go
│   ├── validator.go            # Template validation
│   └── validator_test.go
├── storage/
│   ├── provider.go             # Storage interface
│   ├── local.go                # Local FS implementation
│   ├── local_test.go
│   └── s3.go                   # S3 implementation (v1+)
```

---

## 3. Interfaces

### 3.1 Template Service Interface

```go
package templates

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type Service interface {
    CreateTemplate(ctx context.Context, template *models.Template) error
    GetTemplate(ctx context.Context, id int64) (*models.Template, error)
    GetDefaultTemplate(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
    UpdateTemplate(ctx context.Context, template *models.Template) error
    DeleteTemplate(ctx context.Context, id int64) error
    SetDefault(ctx context.Context, id int64) error
    ListTemplates(ctx context.Context, templateType *models.TemplateType) ([]*models.Template, error)
}
```

### 3.2 Template Renderer Interface

```go
package templates

import (
    "io"
)

type Renderer interface {
    RenderHTML(templateContent string, data interface{}) (string, error)
    RenderText(templateContent string, data interface{}) (string, error)
    RenderToWriter(w io.Writer, templateContent string, data interface{}) error
}
```

### 3.3 Storage Provider Interface

```go
package storage

import (
    "context"
    "io"
)

type Provider interface {
    PutObject(ctx context.Context, path string, data io.Reader, contentType string) error
    GetObject(ctx context.Context, path string) (io.ReadCloser, error)
    DeleteObject(ctx context.Context, path string) error
    GetPublicURL(ctx context.Context, path string) (string, error)
    ListObjects(ctx context.Context, prefix string) ([]string, error)
}
```

---

## 4. Implementation

### 4.1 Template Renderer

```go
package templates

import (
    "bytes"
    "fmt"
    "html/template"
    "io"
)

type renderer struct {
    funcMap template.FuncMap
}

func NewRenderer() Renderer {
    return &renderer{
        funcMap: template.FuncMap{
            "upper": strings.ToUpper,
            "lower": strings.ToLower,
        },
    }
}

func (r *renderer) RenderHTML(templateContent string, data interface{}) (string, error) {
    tmpl, err := template.New("").Funcs(r.funcMap).Parse(templateContent)
    if err != nil {
        return "", fmt.Errorf("failed to parse template: %w", err)
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("failed to execute template: %w", err)
    }
    
    return buf.String(), nil
}
```

### 4.2 Local Storage Provider

```go
package storage

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
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
    fullPath := filepath.Join(p.basePath, path)
    
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    
    file, err := os.Create(fullPath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()
    
    if _, err := io.Copy(file, data); err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }
    
    return nil
}

func (p *localProvider) GetObject(ctx context.Context, path string) (io.ReadCloser, error) {
    fullPath := filepath.Join(p.basePath, path)
    return os.Open(fullPath)
}

func (p *localProvider) GetPublicURL(ctx context.Context, path string) (string, error) {
    return fmt.Sprintf("%s/assets/%s", p.baseURL, path), nil
}
```

---

## 5. Image Validation

```go
package templates

import (
    "bytes"
    "fmt"
    "image"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
)

func ValidateImage(data []byte) error {
    if len(data) > 5*1024*1024 {
        return fmt.Errorf("image size exceeds 5MB")
    }
    
    img, format, err := image.DecodeConfig(bytes.NewReader(data))
    if err != nil {
        return fmt.Errorf("invalid image: %w", err)
    }
    
    if img.Width > 4096 || img.Height > 4096 {
        return fmt.Errorf("image dimensions exceed 4096x4096")
    }
    
    allowedFormats := map[string]bool{
        "jpeg": true,
        "png":  true,
        "gif":  true,
        "webp": true,
    }
    
    if !allowedFormats[format] {
        return fmt.Errorf("unsupported image format: %s", format)
    }
    
    return nil
}
```

---

## 6. Mock Implementations

### 6.1 Mock Template Service

```go
package templates

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type MockService struct {
    CreateTemplateFunc     func(ctx context.Context, template *models.Template) error
    GetTemplateFunc        func(ctx context.Context, id int64) (*models.Template, error)
    GetDefaultTemplateFunc func(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
    UpdateTemplateFunc     func(ctx context.Context, template *models.Template) error
    DeleteTemplateFunc     func(ctx context.Context, id int64) error
    SetDefaultFunc         func(ctx context.Context, id int64) error
    ListTemplatesFunc      func(ctx context.Context, templateType *models.TemplateType) ([]*models.Template, error)
}

func (m *MockService) CreateTemplate(ctx context.Context, template *models.Template) error {
    if m.CreateTemplateFunc != nil {
        return m.CreateTemplateFunc(ctx, template)
    }
    return nil
}

func (m *MockService) GetTemplate(ctx context.Context, id int64) (*models.Template, error) {
    if m.GetTemplateFunc != nil {
        return m.GetTemplateFunc(ctx, id)
    }
    return &models.Template{ID: id}, nil
}

func (m *MockService) GetDefaultTemplate(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
    if m.GetDefaultTemplateFunc != nil {
        return m.GetDefaultTemplateFunc(ctx, templateType)
    }
    return &models.Template{Type: templateType, IsDefault: true}, nil
}

func (m *MockService) UpdateTemplate(ctx context.Context, template *models.Template) error {
    if m.UpdateTemplateFunc != nil {
        return m.UpdateTemplateFunc(ctx, template)
    }
    return nil
}

func (m *MockService) DeleteTemplate(ctx context.Context, id int64) error {
    if m.DeleteTemplateFunc != nil {
        return m.DeleteTemplateFunc(ctx, id)
    }
    return nil
}

func (m *MockService) SetDefault(ctx context.Context, id int64) error {
    if m.SetDefaultFunc != nil {
        return m.SetDefaultFunc(ctx, id)
    }
    return nil
}

func (m *MockService) ListTemplates(ctx context.Context, templateType *models.TemplateType) ([]*models.Template, error) {
    if m.ListTemplatesFunc != nil {
        return m.ListTemplatesFunc(ctx, templateType)
    }
    return []*models.Template{}, nil
}
```

### 6.2 Mock Template Renderer

```go
package templates

import "io"

type MockRenderer struct {
    RenderHTMLFunc     func(templateContent string, data interface{}) (string, error)
    RenderTextFunc     func(templateContent string, data interface{}) (string, error)
    RenderToWriterFunc func(w io.Writer, templateContent string, data interface{}) error
}

func (m *MockRenderer) RenderHTML(templateContent string, data interface{}) (string, error) {
    if m.RenderHTMLFunc != nil {
        return m.RenderHTMLFunc(templateContent, data)
    }
    return "<html>mock</html>", nil
}

func (m *MockRenderer) RenderText(templateContent string, data interface{}) (string, error) {
    if m.RenderTextFunc != nil {
        return m.RenderTextFunc(templateContent, data)
    }
    return "mock text", nil
}

func (m *MockRenderer) RenderToWriter(w io.Writer, templateContent string, data interface{}) error {
    if m.RenderToWriterFunc != nil {
        return m.RenderToWriterFunc(w, templateContent, data)
    }
    w.Write([]byte("mock"))
    return nil
}
```

---

## 7. Dependencies

**Internal:**
- Domain 1 (Auth) - Access control
- Domain 7 (Database) - Template storage

**Dependents:**
- Domain 5 (Email) - Email templates
- Domain 8 (API) - Template endpoints

---

## 8. Testing

```go
func TestTemplateRenderer_RenderHTML(t *testing.T) {
    renderer := NewRenderer()
    
    tmpl := `<h1>{{.Title}}</h1><p>{{.Description}}</p>`
    data := struct {
        Title       string
        Description string
    }{
        Title:       "Test Event",
        Description: "<script>alert('xss')</script>",
    }
    
    result, err := renderer.RenderHTML(tmpl, data)
    if err != nil {
        t.Fatal(err)
    }
    
    if bytes.Contains([]byte(result), []byte("<script>")) {
        t.Error("XSS not prevented")
    }
}
```

---

**Document Status:** ✅ Complete

**Next Domain:** [Domain 8: API & HTTP Handlers](08_API_LLD.md)
