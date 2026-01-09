# Templates Package

## Purpose

Provides a secure template rendering engine with automatic XSS prevention using Go's `html/template` package, along with default system templates and seeding functionality.

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

### Validator (`validator.go`)

Comprehensive template validation and security checking before templates are saved or used.

**Key Features:**
- Syntax validation (parse and execute with test data)
- Variable validation (whitelist enforcement)
- Size limits (HTML: 100KB, Text: 50KB, CSS: 50KB)
- Security checks (XSS prevention verification)

**API:**
```go
validator := NewValidator(engine)
err := validator.ValidateTemplate(template)
```

**Validation Methods:**
- `ValidateTemplate(tmpl)` - Complete validation
- `ValidateSyntax(content, type)` - Parse and execute check
- `ValidateVariables(content, allowedVars)` - Variable whitelist check
- `ValidateSize(content, maxBytes)` - Size limit check

### Seeder (`seeder.go`)

Provides default system templates that are automatically loaded into the database on application startup.

**Key Features:**
- Embeds default templates in the binary using `go:embed`
- Idempotent seeding (safe to run multiple times)
- Creates templates for all three types (invite_email, rsvp_page, confirmation_page)
- Mobile-responsive designs
- Email client compatibility

**API:**
```go
seeder := NewSeeder(templateRepo, systemUserID)
err := seeder.SeedDefaults(ctx)
```

**Default Templates:**
- `defaults/invite_email.html` - HTML email invitation
- `defaults/invite_email.txt` - Plain text email invitation
- `defaults/rsvp_page.html` - RSVP form page
- `defaults/confirmation_page.html` - RSVP confirmation page

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

### Unit Tests

**Engine Tests (`engine_test.go`):**
- Template parsing and execution
- All custom functions
- XSS prevention with various attack vectors
- Error handling
- Thread safety

**Validator Tests (`validator_test.go`):**
- Size validation (within/at/exceeding limits)
- Syntax validation (valid/invalid templates)
- Variable validation (allowed/undefined variables)
- Complete template validation

**Seeder Tests (`seeder_test.go`):**
- Template creation for all types
- Idempotent seeding (no duplicates)
- Template names and metadata
- Context cancellation handling
- Repository error handling
- Template content validation
- Variable usage verification

### Integration Tests

**Engine Integration (`engine_integration_test.go`):**
- Event templates
- Email templates
- RSVP confirmation templates
- Complex nested data structures
- XSS prevention in user data

**Validator Integration (`validator_integration_test.go`):**
- Real-world complete templates
- Edge cases (exact size limits, complex nesting)
- XSS prevention verification
- Advanced XSS payload testing

**Seeder Integration (`seeder_integration_test.go`):**
- Database seeding with real repository
- Idempotent seeding verification
- Template parseability checks
- Template renderability with test data
- Template validation with validator

Run tests:
```bash
go test -timeout 30s ./internal/templates/...
```

**Test Coverage:** 93 tests, all passing

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
