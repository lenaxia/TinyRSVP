# JavaScript Files

This directory contains client-side JavaScript for progressive enhancement of TinyRSVP forms.

## Files

### form_validation.js

Client-side form validation module that provides immediate feedback to users before form submission.

**Features:**
- Real-time validation on blur events
- Email format validation
- Required field validation
- Date/time validation and range checking
- Number validation with min/max constraints
- Custom error messages via `data-error-message` attribute
- Success indicators for valid fields
- Accessibility support (ARIA attributes)
- Progressive enhancement (works without JavaScript via HTML5 validation)

**Usage:**

The validator automatically initializes on page load for all forms with the `novalidate` attribute:

```html
<form method="POST" action="/events" novalidate>
    <!-- form fields -->
</form>
<script src="/static/js/form_validation.js"></script>
```

**Validation Rules:**

1. **Required Fields**: Validates that required fields are not empty
2. **Email**: Validates email format using regex pattern
3. **DateTime**: Validates datetime-local inputs
4. **Number**: Validates numeric inputs with optional min/max constraints
5. **MaxLength**: Validates maximum character length
6. **Date Ranges**: Validates that end dates are after start dates
7. **RSVP Deadline**: Validates that RSVP deadline is before event start

**Custom Error Messages:**

Add `data-error-message` attribute to override default error messages:

```html
<input type="email" name="email" required data-error-message="Please provide a valid email address">
```

**Accessibility:**

- Adds `aria-invalid="true"` to fields with errors
- Creates error messages with `role="alert"`
- Links error messages via `aria-describedby`
- Focuses and scrolls to first error on form submission

**Progressive Enhancement:**

Forms work without JavaScript using HTML5 validation attributes:
- `required`
- `type="email"`
- `type="datetime-local"`
- `min` and `max` for numbers
- `maxlength` for text inputs

## Testing

Tests are located in this directory:
- `form_validation_test.go` - Unit tests for JavaScript structure and features
- `form_validation_integration_test.go` - Integration tests with HTML templates and CSS

Run tests:
```bash
go test -timeout 30s ./static/js/... -v
```

## File Size

The form_validation.js file is kept under 25KB to ensure fast loading on all connections.

## Browser Compatibility

The JavaScript uses modern ES6+ features supported by all current browsers:
- const/let
- Arrow functions
- Template literals
- querySelector/querySelectorAll
- classList API
- addEventListener

For older browsers, HTML5 validation provides fallback functionality.
