# STORY: Dashboard Recent Events Clickable Links

**Epic:** 10 - Technical Debt & Improvements
**Story ID:** 10_STORY_03
**Priority:** Medium
**Estimated Effort:** 2 hours

## User Story

As a user, I want to click on events in the Recent Events section of the dashboard so that I can quickly navigate to events that were recently created or modified.

## Current Issues

1. Recent Events section shows event activity but items are not clickable
2. Event cancellations don't appear in the activity feed
3. No way to navigate directly to an event from the dashboard

## Acceptance Criteria

- [ ] Recent Events items are clickable
- [ ] Clicking an event navigates to event detail page
- [ ] Event cancellations appear in activity feed
- [ ] Activity feed shows: created, updated, published, cancelled events
- [ ] Each activity item shows event title and action type
- [ ] Hover state indicates items are clickable

## Technical Approach

### 1. Update Dashboard Service
- Include event cancellations in activity query
- Add event_id to activity items
- Sort by timestamp descending

### 2. Update Dashboard Template
```html
<div class="activity-item" onclick="window.location.href='/events/{{.EventID}}'" style="cursor: pointer;">
    <div class="activity-item-icon">{{.Icon}}</div>
    <div class="activity-item-content">
        <div class="activity-item-title">
            <a href="/events/{{.EventID}}">{{.EventTitle}}</a>
        </div>
        <div class="activity-item-description">{{.Action}}</div>
        <div class="activity-item-time">{{.Timestamp}}</div>
    </div>
</div>
```

### 3. Add CSS
```css
.activity-item {
    cursor: pointer;
    transition: background-color 0.2s ease;
}

.activity-item:hover {
    background-color: var(--color-surface);
}

.activity-item-title a {
    color: var(--color-text-primary);
    text-decoration: none;
    font-weight: var(--font-weight-semibold);
}

.activity-item-title a:hover {
    color: var(--color-primary-600);
    text-decoration: underline;
}
```

## Tasks

- [ ] Update dashboard service to include cancelled events
- [ ] Add event_id to activity items
- [ ] Make activity items clickable
- [ ] Add hover states
- [ ] Test navigation from dashboard
- [ ] Verify all event types appear (created, updated, published, cancelled)

## Dependencies

None

## Status

- **Status:** ✅ Complete (verified 2026-08-01: `ActivityItem.EventID` populated; clickable links render; worklog 0160)
- **Assigned:** Unassigned
- **Started:** N/A
- **Completed:** N/A
