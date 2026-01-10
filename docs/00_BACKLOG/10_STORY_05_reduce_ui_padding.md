# Epic 10: Technical Debt & Improvements
## Story 05: Reduce Excessive Padding and Spacing in UI

### User Story
As a user, I want UI elements to have more reasonable padding and spacing so that the interface feels less bloated and more content can fit on screen.

### Problem
During UAT, it was identified that almost all UI elements (buttons, tables, cards, forms) have excessive padding, even for mobile views. This makes the interface feel unnecessarily spacious and reduces the amount of visible content.

### Acceptance Criteria
- [ ] Audit all CSS files for padding and spacing values
- [ ] Identify component-specific CSS files with excessive padding:
  - buttons.css
  - dashboard.css
  - forms.css
  - event_form.css
  - event_list.css
  - invite_list.css
  - rsvp_page.css
  - navigation.css
- [ ] Review and reduce padding in:
  - Stats cards
  - Action cards
  - Form inputs and labels
  - Buttons
  - Tables
  - Navigation elements
  - Dashboard layouts
- [ ] Check variables.css for base spacing scale
- [ ] Ensure mobile padding is appropriate (not desktop-sized)
- [ ] Test on multiple screen sizes after changes
- [ ] Verify accessibility (touch targets still 44px minimum)
- [ ] Update any tests that check for specific spacing

### Technical Notes
Current spacing scale in variables.css:
- spacing-0: 0
- spacing-1: 0.25rem (4px)
- spacing-2: 0.5rem (8px)
- spacing-3: 0.75rem (12px)
- spacing-4: 1rem (16px)
- spacing-5: 1.25rem (20px)
- spacing-6: 1.5rem (24px)
- spacing-8: 2rem (32px)
- spacing-10: 2.5rem (40px)
- spacing-12: 3rem (48px)
- spacing-16: 4rem (64px)
- spacing-20: 5rem (80px)
- spacing-24: 6rem (96px)

Likely issues:
- Cards using spacing-8 or higher for padding
- Forms using spacing-6+ for field spacing
- Buttons with excessive vertical padding
- Tables with too much cell padding

### Recommended Approach
1. Read variables.css to confirm spacing scale
2. Audit each component CSS file
3. Reduce padding values systematically:
   - Cards: from p-8/p-6 to p-4/p-3
   - Forms: from gap-6 to gap-3/gap-4
   - Buttons: from py-3/py-4 to py-2/py-3
   - Tables: from p-4 to p-2/p-3
4. Test each change on mobile and desktop
5. Commit changes incrementally by component

### Files to Audit
- static/css/variables.css (spacing scale)
- static/css/buttons.css
- static/css/dashboard.css
- static/css/forms.css
- static/css/event_form.css
- static/css/event_list.css
- static/css/invite_list.css
- static/css/rsvp_page.css
- static/css/navigation.css
- static/css/confirmation.css
- templates/web/*.html (inline classes)

### Status
- Status: Not Started
- Priority: High (UAT feedback)
- Assigned: Unassigned
- Created: 2026-01-10
