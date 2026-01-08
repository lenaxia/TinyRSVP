# User Story: Display Preference Questions on RSVP Page

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 6 hours

---

## User Story

As a **guest**, I want **to see preference questions on the RSVP page** so that **I can provide additional information to the host**.

---

## Acceptance Criteria

- [x] Questions loaded from database for event
- [x] Questions displayed in order (by display_order)
- [x] Question types rendered correctly (text/select/boolean)
- [x] Required questions marked with asterisk
- [ ] Help text displayed if provided
- [x] Mobile-responsive layout
- [x] Accessible (ARIA labels, keyboard navigation)
- [x] Questions only shown if event has questions
- [x] Works without JavaScript

---

## Technical Details

### Template Data

```go
type RSVPPageData struct {
    Event     *models.Event
    Invite    *models.Invite
    Questions []*models.PreferenceQuestion
    // ... other fields
}
```

### Question Rendering

```html
{{range .Questions}}
<div class="question-group" data-question-id="{{.ID}}" data-type="{{.Type}}">
    <label for="question_{{.ID}}">
        {{.QuestionText}}
        {{if .Required}}<span class="required">*</span>{{end}}
    </label>
    
    {{if eq .Type "text"}}
        <textarea 
            id="question_{{.ID}}" 
            name="answers[{{.ID}}][text]"
            maxlength="500"
            {{if .Required}}required{{end}}
            aria-describedby="help_{{.ID}}"
        ></textarea>
    {{else if eq .Type "select"}}
        <select 
            id="question_{{.ID}}" 
            name="answers[{{.ID}}][option]"
            {{if .Required}}required{{end}}
            aria-describedby="help_{{.ID}}"
        >
            <option value="">-- Select --</option>
            {{range .Options}}
            <option value="{{.}}">{{.}}</option>
            {{end}}
        </select>
    {{else if eq .Type "boolean"}}
        <div class="radio-group">
            <label>
                <input type="radio" name="answers[{{.ID}}][boolean]" value="true" {{if .Required}}required{{end}}>
                Yes
            </label>
            <label>
                <input type="radio" name="answers[{{.ID}}][boolean]" value="false" {{if .Required}}required{{end}}>
                No
            </label>
        </div>
    {{end}}
    
    {{if .HelpText}}
    <small id="help_{{.ID}}" class="help-text">{{.HelpText}}</small>
    {{end}}
</div>
{{end}}
```

---

## Tasks

- [x] Load questions in handler
- [x] Pass questions to template
- [x] Create question rendering template
- [x] Style question inputs (mobile-first)
- [x] Add accessibility attributes
- [x] Test with various question types
- [x] Test required vs optional questions
- [x] Test without JavaScript

---

## Dependencies

**Depends on:**
- Story 02_STORY_05: Preference Questions (questions must exist)
- Story 01: RSVP Page (needs page to display on)

**Blocks:**
- Story 06: Answer Submission (needs UI to submit from)

---

## Definition of Done

- [x] All acceptance criteria met (except help text - not in current design)
- [x] Questions render correctly
- [x] Mobile-responsive
- [x] Accessible
- [x] Tests passing
- [x] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
