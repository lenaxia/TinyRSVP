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

- [ ] Questions loaded from database for event
- [ ] Questions displayed in order (by display_order)
- [ ] Question types rendered correctly (text/select/boolean)
- [ ] Required questions marked with asterisk
- [ ] Help text displayed if provided
- [ ] Mobile-responsive layout
- [ ] Accessible (ARIA labels, keyboard navigation)
- [ ] Questions only shown if event has questions
- [ ] Works without JavaScript

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

- [ ] Load questions in handler
- [ ] Pass questions to template
- [ ] Create question rendering template
- [ ] Style question inputs (mobile-first)
- [ ] Add accessibility attributes
- [ ] Test with various question types
- [ ] Test required vs optional questions
- [ ] Test without JavaScript

---

## Dependencies

**Depends on:**
- Story 02_STORY_05: Preference Questions (questions must exist)
- Story 01: RSVP Page (needs page to display on)

**Blocks:**
- Story 06: Answer Submission (needs UI to submit from)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Questions render correctly
- [ ] Mobile-responsive
- [ ] Accessible
- [ ] Tests passing
- [ ] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
