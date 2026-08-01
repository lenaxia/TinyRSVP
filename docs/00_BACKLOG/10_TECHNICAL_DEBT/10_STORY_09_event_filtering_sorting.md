# Epic 10: Technical Debt & Improvements
## Story 06: Event List Filtering and Sorting

### User Story
As an event manager, I want to filter events by status, search by text, and sort by different criteria so that I can quickly find the events I'm looking for.

### Problem
The events list page (`/events`) has filter controls (status dropdown, search input, sort dropdown) but they are non-functional:
1. Status filter dropdown doesn't update the page when changed
2. Search input doesn't filter events
3. Sort dropdown doesn't reorder events
4. No JavaScript is wired up to handle these interactions

The backend handler (`ListEventsPage` in `events_web.go`) does support:
- Status filtering via `?status=` query parameter
- Pagination via `?page=` query parameter

But missing:
- Search/text filtering
- Sort ordering
- JavaScript to trigger page updates

### Acceptance Criteria
- [ ] Add JavaScript to handle status filter changes
  - On change, reload page with `?status=<value>` query parameter
  - Preserve other query parameters (page, sort, search)
- [ ] Add JavaScript to handle search input
  - Debounce input (300ms delay)
  - Reload page with `?search=<value>` query parameter
  - Preserve other query parameters
- [ ] Add JavaScript to handle sort dropdown changes
  - On change, reload page with `?sort=<value>` query parameter
  - Preserve other query parameters
- [ ] Update backend `ListEventsPage` handler to support:
  - Search parameter (filter by title/description)
  - Sort parameter (date, name, rsvp count)
- [ ] Update `events.Service.ListEvents` to support search and sort
- [ ] Add tests for filtering, searching, and sorting
- [ ] Ensure query parameters are preserved in pagination links

### Technical Implementation

#### Frontend (JavaScript)
Create `static/js/event_filters.js`:
```javascript
document.addEventListener('DOMContentLoaded', function() {
    const statusFilter = document.getElementById('status-filter');
    const searchInput = document.querySelector('.search-input');
    const sortSelect = document.getElementById('sort-select');
    
    function updateURL() {
        const params = new URLSearchParams(window.location.search);
        
        if (statusFilter.value !== 'all') {
            params.set('status', statusFilter.value);
        } else {
            params.delete('status');
        }
        
        if (searchInput.value) {
            params.set('search', searchInput.value);
        } else {
            params.delete('search');
        }
        
        if (sortSelect.value !== 'date') {
            params.set('sort', sortSelect.value);
        } else {
            params.delete('sort');
        }
        
        params.delete('page');
        
        window.location.search = params.toString();
    }
    
    statusFilter.addEventListener('change', updateURL);
    sortSelect.addEventListener('change', updateURL);
    
    let searchTimeout;
    searchInput.addEventListener('input', function() {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(updateURL, 300);
    });
});
```

#### Backend Updates

1. Update `ListEventsPage` in `events_web.go`:
   - Add search parameter handling
   - Add sort parameter handling
   - Pass to service layer

2. Update `events.ListFilters` struct:
   - Add `Search string` field
   - Add `SortBy string` field
   - Add `SortOrder string` field

3. Update `events.Service.ListEvents`:
   - Implement search filtering (title, description)
   - Implement sorting (date, name, rsvp count)

4. Update pagination links in template:
   - Preserve status, search, sort parameters

### Files to Modify
- `static/js/event_filters.js` (create new)
- `templates/web/event_list.html` (add script tag)
- `internal/handlers/events_web.go` (add search/sort handling)
- `internal/events/service.go` (add search/sort logic)
- `internal/events/filters.go` or similar (update ListFilters struct)
- Tests for all changes

### Status
- Status: ⚠️ In Progress (partial) — verified 2026-08-01
  - **Done:** Backend `?status=` filtering works (`events_web.go:106` → `ListFilters.Status`).
  - **Broken:** Frontend status/search/sort controls navigate to `/not-implemented` (`event_list.html:48,64,70`) instead of submitting `?status=`.
  - **Not started:** Search and sort (no `Search`/`SortBy`/`SortOrder` fields on `ListFilters`).
- Priority: High (UAT feedback - broken functionality)
- Assigned: Unassigned
- Created: 2026-01-10
