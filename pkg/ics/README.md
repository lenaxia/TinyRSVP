# ICS Package

## Purpose

Generates RFC 5545 compliant iCalendar (.ics) files for event invitations.

## Structure

- `generator.go` - ICS file generator implementation
- `generator_test.go` - Comprehensive test suite

## Interface

```go
type Generator interface {
    Generate(event *models.Event, rsvpURL string) ([]byte, error)
}
```

## Usage

```go
import "github.com/lenaxia/tinyrsvp/pkg/ics"

gen := ics.NewGenerator()
icsData, err := gen.Generate(event, "https://example.com/rsvp/token123")
if err != nil {
    return err
}
```

## Features

- RFC 5545 compliant VCALENDAR format
- Timezone support (TZID parameter)
- Optional end time
- Location and description fields
- RSVP URL embedded in description
- 24-hour reminder alarm (VALARM)
- Proper escaping of special characters
- Sequence number for event updates

## Special Character Escaping

The generator properly escapes ICS special characters:
- Backslash `\` → `\\`
- Comma `,` → `\,`
- Semicolon `;` → `\;`
- Newline `\n` → `\\n`

## Output Format

```
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//TinyRSVP//EN
METHOD:REQUEST
BEGIN:VEVENT
UID:<event_id>@tinyrsvp
DTSTAMP:<timestamp>
DTSTART;TZID=<timezone>:<start_time>
DTEND;TZID=<timezone>:<end_time>
SUMMARY:<event_title>
LOCATION:<location>
DESCRIPTION:<description>\n\nRSVP: <rsvp_url>
STATUS:CONFIRMED
SEQUENCE:<ics_sequence>
BEGIN:VALARM
TRIGGER:-PT24H
ACTION:DISPLAY
DESCRIPTION:Reminder: <event_title> tomorrow
END:VALARM
END:VEVENT
END:VCALENDAR
```

## Testing

```bash
# Run all ICS tests
go test -timeout 30s ./pkg/ics/...

# Run with coverage
go test -timeout 30s -cover ./pkg/ics/...
```

## Test Coverage

- Basic event generation
- Minimal event (only required fields)
- Event with end time
- Special character escaping
- RSVP URL inclusion
- Alarm generation
- Valid ICS format structure
- Helper function tests (escapeICS, formatICSTime)

## Dependencies

- `internal/models` - Event model
- Standard library only (no external dependencies)

## Related Files

- Email service: `internal/email/service.go`
- Email templates: `templates/email/rsvp_confirmation.{html,txt}`
- RSVP service: `internal/rsvp/service.go`
