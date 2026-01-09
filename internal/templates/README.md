# Templates Package

## Purpose

Provides a secure template rendering engine with automatic XSS prevention using Go's `html/template` package.

## Components

### Engine (`engine.go`)

The core template engine that wraps Go's `html/template` with custom functions and security features.

**Key Features:**
- Automatic HTML escaping for XSS prevention
- Custom template functions for common operations
- Thread-safe template execution
- Support for both string and writer-based rendering

**API:**
```go
engine := NewEngine()
tmpl, err := engine.Parse(templateString)
result, err := engine.ExecuteToString(tmpl, data)
```

### Custom Template Functions

The engine provides the following custom functions:

#### String Functions
- `upper` - Convert string to uppercase
- `lower` - Convert string to lowercase  
- `title` - Convert string to title case

#### Date/Time Functions
- `formatDate` - Format time with custom layout: `{{formatDate .Time "2006-01-02"}}`
- `formatTime` - Format time as "3:04 PM": `{{formatTime .Time}}`
- `formatDateTime` - Format as "Monday, January 2, 2006 at 3:04 PM": `{{formatDateTime .Time}}`

#### Utility Functions
- `truncate` - Truncate string to length with "...": `{{truncate .Text 50}}`
- `default` - Provide default value for empty strings: `{{default .Value "default"}}`

#### Safety Functions (Use with Caution)
- `safeHTML` - Mark HTML as safe (bypasses escaping): `{{safeHTML .TrustedHTML}}`
- `safeURL` - Mark URL as safe: `{{safeURL .TrustedURL}}`
- `safeCSS` - Mark CSS as safe: `{{safeCSS .TrustedCSS}}`

**Warning:** Only use safety functions with trusted, validated content. Never use with user-provided data.

## Security

### XSS Prevention

The engine automatically escapes all template variables to prevent XSS attacks:

```go
// Input: <script>alert('xss')</script>
// Output: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;
```

### Safe Content

Use safety functions only when you have pre-validated and sanitized content:

```go
// SAFE: Content from trusted source
{{safeHTML .AdminGeneratedHTML}}

// UNSAFE: Never do this with user input
{{safeHTML .UserProvidedContent}}  // ❌ DANGEROUS
```

## Testing

### Unit Tests (`engine_test.go`)

Comprehensive unit tests covering:
- Template parsing and execution
- All custom functions
- XSS prevention with various attack vectors
- Error handling
- Thread safety

### Integration Tests (`engine_integration_test.go`)

Real-world scenario tests with:
- Event templates
- Email templates
- RSVP confirmation templates
- Complex nested data structures
- XSS prevention in user data

Run tests:
```bash
go test -timeout 30s ./internal/templates/...
```

## Usage Examples

### Basic Template

```go
engine := NewEngine()

tmpl, err := engine.Parse("<h1>{{.Title}}</h1>")
if err != nil {
    return err
}

data := struct{ Title string }{Title: "Hello World"}
result, err := engine.ExecuteToString(tmpl, data)
```

### Event Invitation

```go
tmplStr := `
Dear {{.Name}},

You're invited to {{.Event.Title}}!

Date: {{formatDateTime .Event.StartTime}}
Location: {{.Event.Location}}

RSVP: {{.RSVPLink}}
`

tmpl, _ := engine.Parse(tmplStr)
result, _ := engine.ExecuteToString(tmpl, inviteData)
```

### With Conditionals and Loops

```go
tmplStr := `
<h1>{{.Title | upper}}</h1>
{{if .Items}}
<ul>
{{range .Items}}
    <li>{{.Name}} - {{formatDate .Date "Jan 2, 2006"}}</li>
{{end}}
</ul>
{{else}}
<p>No items found.</p>
{{end}}
`
```

## Performance

The engine is designed for efficiency:
- Templates can be parsed once and reused
- Thread-safe for concurrent execution
- Minimal memory allocations
- Benchmarks included in test suite

## Integration

This package is used by:
- `internal/email` - Email template rendering
- Future: Web page rendering
- Future: PDF generation

## Dependencies

**Standard Library Only:**
- `html/template` - Core template engine with XSS protection
- `bytes` - Buffer management
- `strings` - String operations
- `time` - Date/time formatting
- `io` - Writer interface

## Future Enhancements

Potential additions (not yet implemented):
- Template caching
- Template inheritance/layouts
- Additional custom functions as needed
- Template validation utilities
