# User Story: Plus Ones Input UI

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 6 hours
**Actual Effort:** 1 hour

---

## User Story

As a **guest**, I want **to specify how many plus ones I'm bringing** so that **the host knows the total guest count**.

---

## Acceptance Criteria

- [ ] Plus ones input field displayed on RSVP form
- [ ] Field only shown if invite.max_plus_ones > 0
- [ ] Input type number with min=0, max=invite.max_plus_ones
- [ ] Field disabled if response is "no"
- [ ] Clear label showing limit (e.g., "Guests (up to 2)")
- [ ] Client-side validation before submission
- [ ] Server-side validation on submission
- [ ] Error messages displayed inline
- [ ] Mobile-friendly input (large touch target)
- [ ] Accessible (keyboard navigation, screen reader)
- [ ] Works without JavaScript (form validation)

---

## Technical Details

### HTML Structure

```html
<div class="form-group plus-ones-group" data-max-plus-ones="{{.Invite.MaxPlusOnes}}">
    <label for="plus_ones">
        Guests
        <span class="limit-text">(up to {{.Invite.MaxPlusOnes}})</span>
    </label>
    
    <input 
        type="number" 
        id="plus_ones" 
        name="plus_ones" 
        min="0" 
        max="{{.Invite.MaxPlusOnes}}"
        value="0"
        aria-describedby="plus_ones_help"
        required
    />
    
    <small id="plus_ones_help" class="form-text">
        How many additional guests will you bring?
    </small>
    
    <div class="error-message" role="alert" aria-live="polite"></div>
</div>
```

### JavaScript Enhancement (Progressive)

```javascript
// Progressive enhancement for better UX
document.addEventListener('DOMContentLoaded', function() {
    const responseInputs = document.querySelectorAll('input[name="response"]');
    const plusOnesInput = document.getElementById('plus_ones');
    const plusOnesGroup = document.querySelector('.plus-ones-group');
    
    if (!responseInputs.length || !plusOnesInput) return;
    
    function updatePlusOnesState() {
        const selectedResponse = document.querySelector('input[name="response"]:checked');
        
        if (selectedResponse && selectedResponse.value === 'no') {
            plusOnesInput.value = '0';
            plusOnesInput.disabled = true;
            plusOnesGroup.classList.add('disabled');
        } else {
            plusOnesInput.disabled = false;
            plusOnesGroup.classList.remove('disabled');
        }
    }
    
    responseInputs.forEach(input => {
        input.addEventListener('change', updatePlusOnesState);
    });
    
    // Validate on input
    plusOnesInput.addEventListener('input', function() {
        const value = parseInt(this.value);
        const max = parseInt(this.max);
        const errorDiv = plusOnesGroup.querySelector('.error-message');
        
        if (value < 0) {
            errorDiv.textContent = 'Cannot be negative';
            this.setCustomValidity('Cannot be negative');
        } else if (value > max) {
            errorDiv.textContent = `You can bring up to ${max} guest(s)`;
            this.setCustomValidity(`Maximum ${max} guests`);
        } else {
            errorDiv.textContent = '';
            this.setCustomValidity('');
        }
    });
    
    updatePlusOnesState();
});
```

---

## Tasks

### Phase 1: HTML/Template
- [ ] Add plus ones input to RSVP form template
- [ ] Add conditional rendering (only if max_plus_ones > 0)
- [ ] Add proper labels and help text
- [ ] Add ARIA attributes for accessibility
- [ ] Test template rendering

### Phase 2: CSS Styling
- [ ] Style input field for mobile (44px min height)
- [ ] Style label and help text
- [ ] Style error message display
- [ ] Add disabled state styling
- [ ] Add focus indicators
- [ ] Test on multiple screen sizes

### Phase 3: JavaScript Enhancement
- [ ] Implement auto-disable on "no" response
- [ ] Implement client-side validation
- [ ] Implement inline error display
- [ ] Add input sanitization
- [ ] Test without JavaScript (graceful degradation)

### Phase 4: Accessibility
- [ ] Test keyboard navigation
- [ ] Test with screen reader
- [ ] Verify ARIA labels
- [ ] Test focus management
- [ ] Verify color contrast

### Phase 5: Integration Testing
- [ ] Test with various max_plus_ones values
- [ ] Test response type changes
- [ ] Test form submission
- [ ] Test error scenarios
- [ ] Test mobile devices

---

## UI States

### State 1: Response = "Yes" or "Maybe"
```
┌─────────────────────────────────┐
│ Guests (up to 2)                │
│ ┌─────────────────────────────┐ │
│ │ [  2  ]  ▲▼                 │ │
│ └─────────────────────────────┘ │
│ How many additional guests      │
│ will you bring?                 │
└─────────────────────────────────┘
```

### State 2: Response = "No"
```
┌─────────────────────────────────┐
│ Guests (up to 2)     [DISABLED] │
│ ┌─────────────────────────────┐ │
│ │ [  0  ]  ▲▼                 │ │
│ └─────────────────────────────┘ │
│ Cannot bring guests when        │
│ declining                       │
└─────────────────────────────────┘
```

### State 3: Validation Error
```
┌─────────────────────────────────┐
│ Guests (up to 2)                │
│ ┌─────────────────────────────┐ │
│ │ [  5  ]  ▲▼                 │ │ ← Red border
│ └─────────────────────────────┘ │
│ ⚠ You can bring up to 2 guest(s)│ ← Red text
└─────────────────────────────────┘
```

---

## Mobile CSS

```css
.plus-ones-group {
    margin-bottom: 1.5rem;
}

.plus-ones-group label {
    display: block;
    font-weight: 600;
    margin-bottom: 0.5rem;
    font-size: 1rem;
}

.plus-ones-group .limit-text {
    color: #666;
    font-weight: 400;
    font-size: 0.875rem;
}

.plus-ones-group input[type="number"] {
    width: 100%;
    min-height: 44px;
    padding: 0.75rem;
    font-size: 1rem;
    border: 2px solid #ddd;
    border-radius: 4px;
    transition: border-color 0.2s;
}

.plus-ones-group input[type="number"]:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
}

.plus-ones-group input[type="number"]:disabled {
    background-color: #f5f5f5;
    color: #999;
    cursor: not-allowed;
}

.plus-ones-group.disabled {
    opacity: 0.6;
}

.plus-ones-group .form-text {
    display: block;
    margin-top: 0.5rem;
    font-size: 0.875rem;
    color: #666;
}

.plus-ones-group .error-message {
    display: block;
    margin-top: 0.5rem;
    font-size: 0.875rem;
    color: #dc3545;
    font-weight: 500;
}

.plus-ones-group .error-message:empty {
    display: none;
}

.plus-ones-group input.invalid {
    border-color: #dc3545;
}

/* Desktop styles */
@media (min-width: 768px) {
    .plus-ones-group input[type="number"] {
        max-width: 200px;
    }
}
```

---

## Validation Rules

### Client-Side
- Min value: 0
- Max value: invite.max_plus_ones
- Must be integer
- Auto-set to 0 if response is "no"

### Server-Side
- Same as client-side
- Additional check against invite status
- Additional check against event status

---

## Testing Strategy

### Manual Testing Checklist
- [ ] Input accepts valid values (0 to max)
- [ ] Input rejects negative values
- [ ] Input rejects values > max
- [ ] Input rejects non-integer values
- [ ] Input disabled when response = "no"
- [ ] Input enabled when response = "yes" or "maybe"
- [ ] Error messages display correctly
- [ ] Form submits with valid values
- [ ] Form blocks submission with invalid values
- [ ] Works on mobile devices
- [ ] Works with keyboard only
- [ ] Works with screen reader

### Automated Tests

```javascript
describe('Plus Ones Input', () => {
    it('should disable input when response is no', () => {
        cy.visit('/rsvp/testtoken');
        cy.get('input[value="no"]').click();
        cy.get('#plus_ones').should('be.disabled');
        cy.get('#plus_ones').should('have.value', '0');
    });
    
    it('should enable input when response is yes', () => {
        cy.visit('/rsvp/testtoken');
        cy.get('input[value="yes"]').click();
        cy.get('#plus_ones').should('not.be.disabled');
    });
    
    it('should show error for exceeding limit', () => {
        cy.visit('/rsvp/testtoken');
        cy.get('input[value="yes"]').click();
        cy.get('#plus_ones').clear().type('10');
        cy.get('.error-message').should('contain', 'up to 2 guest(s)');
    });
    
    it('should allow valid plus ones', () => {
        cy.visit('/rsvp/testtoken');
        cy.get('input[value="yes"]').click();
        cy.get('#plus_ones').clear().type('2');
        cy.get('.error-message').should('be.empty');
    });
});
```

---

## Dependencies

**Depends on:**
- Story 01: RSVP Page (needs page to add UI to)
- Story 03: Plus Ones Validation (needs validation logic)

**Blocks:**
- Story 02: RSVP Submission (needs UI for input)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] HTML template implemented
- [ ] CSS styling complete
- [ ] JavaScript enhancement working
- [ ] Client-side validation working
- [ ] Accessibility requirements met
- [ ] Mobile-responsive
- [ ] Works without JavaScript
- [ ] Manual testing complete
- [ ] Automated tests passing
- [ ] Documentation updated
- [ ] Code reviewed

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **Frontend:** README-LLM.md Section "Frontend: Plain CSS + Vanilla JavaScript"
- **Accessibility:** WCAG 2.1 AA Guidelines
