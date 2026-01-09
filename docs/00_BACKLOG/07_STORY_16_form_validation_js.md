# User Story: Client-Side Form Validation

**Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **user**, I want **immediate feedback on form errors** so that **I can correct mistakes before submitting**.

---

## Acceptance Criteria

- [ ] Real-time validation on blur
- [ ] Email format validation
- [ ] Required field validation
- [ ] Date/time validation
- [ ] Custom validation messages
- [ ] Error display near fields
- [ ] Success indicators
- [ ] Works without JavaScript (HTML5 validation fallback)

---

## Technical Details

```javascript
const FormValidator = {
    validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    },
    
    validateRequired(value) {
        return value.trim().length > 0;
    },
    
    showError(field, message) {
        const errorEl = field.nextElementSibling;
        errorEl.textContent = message;
        field.classList.add('error');
    },
    
    clearError(field) {
        const errorEl = field.nextElementSibling;
        errorEl.textContent = '';
        field.classList.remove('error');
    }
};
```

---

## Tasks

- [ ] Create validation JavaScript module
- [ ] Add email validation
- [ ] Add required field validation
- [ ] Add custom validators
- [ ] Display error messages
- [ ] Test all validation rules
- [ ] Test without JavaScript

---

## Dependencies

**Depends on:** 07_STORY_06_forms.md

**Blocks:** None

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Validation working
- [ ] Fallback to HTML5 validation
- [ ] Changes committed to git

---

## References

- **Epic:** [07_EPIC_frontend.md](07_EPIC_frontend.md)
