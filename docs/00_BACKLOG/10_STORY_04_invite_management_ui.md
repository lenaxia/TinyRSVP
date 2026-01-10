# STORY: Invite Management UI

**Epic:** 10 - Technical Debt & Improvements  
**Story ID:** 10_STORY_04  
**Priority:** High  
**Estimated Effort:** 6 hours

## User Story

As an event manager, I want a user-friendly interface to create and manage invites so that I don't have to use API endpoints directly.

## Current Issues

1. `/events/{id}/invites` page has "Import Invites" button but it's not wired up
2. No UI to create individual invites
3. No UI to manually add guests
4. Users must use API endpoints directly to create invites

## Acceptance Criteria

- [ ] Import Invites button opens modal or form
- [ ] Can upload CSV file with guest list
- [ ] Can create individual invites through UI
- [ ] Can manually add guest name and email
- [ ] Form validation for email addresses
- [ ] Success/error feedback after creation
- [ ] Newly created invites appear in list immediately

## Technical Approach

### Option A: Modal-Based (Recommended)
Use modals for both import and manual creation:

```html
<!-- Import Modal -->
<div class="modal" id="import-modal">
    <div class="modal-content">
        <h2>Import Invites</h2>
        <form id="import-form">
            <input type="file" accept=".csv" name="file">
            <button type="submit" class="btn btn-primary">Import</button>
        </form>
    </div>
</div>

<!-- Manual Create Modal -->
<div class="modal" id="create-modal">
    <div class="modal-content">
        <h2>Create Invite</h2>
        <form id="create-form">
            <input type="text" name="name" placeholder="Guest Name">
            <input type="email" name="email" placeholder="Email" required>
            <input type="number" name="max_plus_ones" min="0" max="10" value="0">
            <button type="submit" class="btn btn-primary">Create</button>
        </form>
    </div>
</div>
```

### Option B: Inline Forms
Add collapsible forms directly on the invite list page

### Option C: Separate Pages
Create `/events/{id}/invites/new` and `/events/{id}/invites/import` pages

## Implementation Steps

### 1. Create Modal CSS Component
- `static/css/modal.css`
- Backdrop, content, close button
- Responsive design
- Accessibility (focus trap, ESC to close)

### 2. Create Modal JavaScript
- `static/js/modal.js`
- Open/close functions
- Focus management
- ESC key handler

### 3. Wire Up Import Button
```javascript
document.querySelector('.import-btn').addEventListener('click', () => {
    Modal.open('import-modal');
});

document.querySelector('#import-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const response = await fetch(`/api/events/${eventId}/invites/import`, {
        method: 'POST',
        body: formData
    });
    if (response.ok) {
        Modal.close('import-modal');
        window.location.reload();
    }
});
```

### 4. Add Create Invite Button & Modal
- Add "Create Invite" button next to "Import Invites"
- Wire up to manual invite API endpoint
- Form validation
- Success feedback

### 5. Update Invite List Template
- Add modal HTML
- Add button event listeners
- Include modal.css and modal.js

## Tasks

- [ ] Create modal.css component
- [ ] Create modal.js component
- [ ] Write tests for modal component
- [ ] Add import modal to invite_list.html
- [ ] Add create modal to invite_list.html
- [ ] Wire up Import Invites button
- [ ] Wire up Create Invite button
- [ ] Add form validation
- [ ] Add success/error feedback
- [ ] Test file upload
- [ ] Test manual creation
- [ ] Test on mobile
- [ ] Update documentation

## Dependencies

- Existing API endpoints: `/api/events/{id}/invites/import`, `/api/events/{id}/invites/manual`

## Notes

- Modal should close on backdrop click
- Modal should close on ESC key
- Focus should return to trigger button on close
- Form should reset after successful submission
- Consider adding bulk create (multiple emails at once)

## Status

- **Status:** Not Started
- **Assigned:** Unassigned
- **Started:** N/A
- **Completed:** N/A
