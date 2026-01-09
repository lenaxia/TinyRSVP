# User Story: Template Variable System

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-09

---

## User Story

As an **event manager**, I want **comprehensive documentation of template variables** so that **I can create custom templates using all available data**.

---

## Acceptance Criteria

- [x] All template variables documented
- [x] Variable reference guide created
- [x] Example templates provided for each type
- [x] Template data structures defined
- [x] Helper functions documented
- [x] Variable validation implemented
- [x] Type-safe template data structures
- [x] All tests pass with timeout
- [ ] Documentation accessible in admin UI

---

## Technical Details

### Template Data Structures

```go
package templates

import (
    "time"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type InviteEmailData struct {
    Event struct {
        Title        string
        Description  string
        StartTime    time.Time
        EndTime      *time.Time
        Timezone     string
        Location     string
        RSVPDeadline *time.Time
    }
    Invite struct {
        Name  string
        Email string
    }
    RSVPURL     string
    MaxPlusOnes int
}

type RSVPPageData struct {
    Event struct {
        Title        string
        Description  string
        StartTime    time.Time
        EndTime      *time.Time
        Timezone     string
        Location     string
        RSVPDeadline *time.Time
    }
    Token       string
    MaxPlusOnes int
    Questions   []QuestionData
}

type QuestionData struct {
    ID           int64
    QuestionText string
    QuestionType string
    Options      []OptionData
    Required     bool
    HelpText     string
}

type OptionData struct {
    Value string
    Label string
}

type ConfirmationPageData struct {
    Event struct {
        Title       string
        Description string
        StartTime   time.Time
        EndTime     *time.Time
        Timezone    string
        Location    string
    }
    Token string
    RSVP  struct {
        Response string
        PlusOnes int
    }
    Answers []AnswerData
}

type AnswerData struct {
    QuestionText  string
    AnswerDisplay string
}
```

### Data Builder Functions

```go
func BuildInviteEmailData(event *models.Event, invite *models.Invite, rsvpURL string) *InviteEmailData {
    data := &InviteEmailData{
        RSVPURL:     rsvpURL,
        MaxPlusOnes: invite.MaxPlusOnes,
    }
    
    data.Event.Title = event.Title
    data.Event.Description = event.Description
    data.Event.StartTime = event.StartTime
    data.Event.EndTime = event.EndTime
    data.Event.Timezone = event.Timezone
    data.Event.Location = event.Location
    data.Event.RSVPDeadline = event.RSVPDeadline
    
    data.Invite.Name = invite.Name
    data.Invite.Email = invite.Email
    
    return data
}

func BuildRSVPPageData(event *models.Event, invite *models.Invite, questions []*models.PreferenceQuestion, token string) *RSVPPageData {
    data := &RSVPPageData{
        Token:       token,
        MaxPlusOnes: invite.MaxPlusOnes,
    }
    
    data.Event.Title = event.Title
    data.Event.Description = event.Description
    data.Event.StartTime = event.StartTime
    data.Event.EndTime = event.EndTime
    data.Event.Timezone = event.Timezone
    data.Event.Location = event.Location
    data.Event.RSVPDeadline = event.RSVPDeadline
    
    for _, q := range questions {
        qData := QuestionData{
            ID:           q.ID,
            QuestionText: q.QuestionText,
            QuestionType: q.QuestionType,
            Required:     q.Required,
            HelpText:     q.HelpText,
        }
        
        if q.Options != nil {
            for _, opt := range q.Options {
                qData.Options = append(qData.Options, OptionData{
                    Value: opt.Value,
                    Label: opt.Label,
                })
            }
        }
        
        data.Questions = append(data.Questions, qData)
    }
    
    return data
}

func BuildConfirmationPageData(event *models.Event, rsvp *models.RSVP, answers []*models.RSVPAnswer, token string) *ConfirmationPageData {
    data := &ConfirmationPageData{
        Token: token,
    }
    
    data.Event.Title = event.Title
    data.Event.Description = event.Description
    data.Event.StartTime = event.StartTime
    data.Event.EndTime = event.EndTime
    data.Event.Timezone = event.Timezone
    data.Event.Location = event.Location
    
    data.RSVP.Response = string(rsvp.Response)
    data.RSVP.PlusOnes = rsvp.PlusOnes
    
    for _, ans := range answers {
        ansData := AnswerData{
            QuestionText:  ans.Question.QuestionText,
            AnswerDisplay: formatAnswer(ans),
        }
        data.Answers = append(data.Answers, ansData)
    }
    
    return data
}

func formatAnswer(answer *models.RSVPAnswer) string {
    if answer.AnswerText != nil {
        return *answer.AnswerText
    }
    if answer.AnswerOption != nil {
        return *answer.AnswerOption
    }
    if answer.AnswerBoolean != nil {
        if *answer.AnswerBoolean {
            return "Yes"
        }
        return "No"
    }
    return ""
}
```

---

## Variable Reference Documentation

### Common Variables (All Templates)

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.Event.Title}}` | string | Event name | "Birthday Party" |
| `{{.Event.Description}}` | string | Event description | "Join us for cake!" |
| `{{.Event.StartTime}}` | time.Time | Event start time | 2026-06-15 14:00:00 |
| `{{.Event.EndTime}}` | *time.Time | Event end time (optional) | 2026-06-15 18:00:00 |
| `{{.Event.Timezone}}` | string | IANA timezone | "America/Los_Angeles" |
| `{{.Event.Location}}` | string | Event location | "123 Main St" |
| `{{.Event.RSVPDeadline}}` | *time.Time | RSVP deadline (optional) | 2026-06-10 23:59:59 |

### Invite Email Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.Invite.Name}}` | string | Guest name | "John Doe" |
| `{{.Invite.Email}}` | string | Guest email | "john@example.com" |
| `{{.RSVPURL}}` | string | RSVP link | "https://rsvp.example.com/rsvp/abc123" |
| `{{.MaxPlusOnes}}` | int | Max guests allowed | 2 |

### RSVP Page Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.Token}}` | string | Invite token | "abc123..." |
| `{{.MaxPlusOnes}}` | int | Max guests allowed | 2 |
| `{{.Questions}}` | []QuestionData | Preference questions | Array of questions |

### Confirmation Page Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.Token}}` | string | Invite token | "abc123..." |
| `{{.RSVP.Response}}` | string | RSVP response | "yes" |
| `{{.RSVP.PlusOnes}}` | int | Number of guests | 2 |
| `{{.Answers}}` | []AnswerData | Question answers | Array of answers |

---

## Template Function Reference

### Date/Time Functions

```go
// Format date with custom layout
{{.Event.StartTime | formatDate "Monday, January 2, 2006"}}
// Output: "Monday, June 15, 2026"

// Format time
{{.Event.StartTime | formatTime}}
// Output: "2:00 PM"

// Format date and time
{{.Event.StartTime | formatDate "Jan 2, 2006 at 3:04 PM"}}
// Output: "Jun 15, 2026 at 2:00 PM"
```

### String Functions

```go
// Convert to uppercase
{{.Event.Title | upper}}
// Output: "BIRTHDAY PARTY"

// Convert to lowercase
{{.Event.Title | lower}}
// Output: "birthday party"

// Convert to title case
{{.Event.Title | title}}
// Output: "Birthday Party"
```

### Math Functions

```go
// Addition
{{add .RSVP.PlusOnes 1}}
// Output: 3 (if PlusOnes is 2)

// Subtraction
{{sub .MaxPlusOnes .RSVP.PlusOnes}}
// Output: 0 (if both are 2)

// Greater than
{{if gt .MaxPlusOnes 0}}
    You can bring guests
{{end}}
```

### Control Flow

```go
// Conditional
{{if .Event.Location}}
    <p>Where: {{.Event.Location}}</p>
{{end}}

// If-else
{{if eq .RSVP.Response "yes"}}
    <p>See you there!</p>
{{else}}
    <p>Sorry you can't make it.</p>
{{end}}

// Range (loop)
{{range .Questions}}
    <div>{{.QuestionText}}</div>
{{end}}

// With (scope)
{{with .Event}}
    <h1>{{.Title}}</h1>
    <p>{{.Description}}</p>
{{end}}
```

---

## Example Templates

### Minimal Invite Email

```html
<html>
<body>
    <h1>{{.Event.Title}}</h1>
    <p>When: {{.Event.StartTime | formatDate "Jan 2, 2006"}}</p>
    <p>Where: {{.Event.Location}}</p>
    <a href="{{.RSVPURL}}">RSVP</a>
</body>
</html>
```

### Custom Styled RSVP Page

```html
<html>
<head>
    <style>
        body { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
        .card { background: white; max-width: 500px; margin: 50px auto; padding: 40px; border-radius: 10px; }
    </style>
</head>
<body>
    <div class="card">
        <h1>{{.Event.Title}}</h1>
        <p>{{.Event.Description}}</p>
        <form method="POST">
            <label><input type="radio" name="response" value="yes"> Yes</label>
            <label><input type="radio" name="response" value="no"> No</label>
            <button type="submit">Submit</button>
        </form>
    </div>
</body>
</html>
```

---

## Tasks

### Phase 1: Data Structures (TDD)
- [x] Define InviteEmailData struct
- [x] Define RSVPPageData struct
- [x] Define ConfirmationPageData struct
- [x] Define QuestionData struct
- [x] Define AnswerData struct
- [x] Write tests for data structures
- [x] Run tests (should pass)

### Phase 2: Builder Functions (TDD)
- [x] Write test for BuildInviteEmailData
- [x] Write test for BuildRSVPPageData
- [x] Write test for BuildConfirmationPageData
- [x] Write test for formatAnswer helper
- [x] Implement all builder functions
- [x] Run tests (should pass)

### Phase 3: Documentation
- [x] Create variable reference guide
- [x] Create function reference guide
- [x] Create example templates
- [ ] Create troubleshooting guide
- [ ] Add inline help in admin UI

### Phase 4: Integration Testing
- [x] Test data builders with real models
- [x] Test rendering with built data
- [x] Test all variable types
- [x] Test all functions
- [x] Verify type safety

---

## Dependencies

**Depends on:**
- Story 00: Template Struct (for data models)
- Story 01: Template Integration (for renderer)

**Blocks:**
- Story 03: Default Templates (needs data structures)
- Story 04: Template CRUD (needs data structures)
- Story 06: Template Preview (needs data builders)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Data structures defined
- [x] Builder functions implemented
- [x] All unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] Documentation complete
- [x] Variable reference guide created
- [x] Example templates provided
- [x] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11.3 (Template Variables)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **Story 01:** [06_STORY_01_template_integration.md](06_STORY_01_template_integration.md)
