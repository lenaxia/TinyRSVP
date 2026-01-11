# Story 10.07: Reusable DateTime Picker Component

**Epic:** 10 - UI/UX Improvements and Polish  
**Status:** Backlog  
**Priority:** Medium  
**Estimated Effort:** 3-5 hours

## User Story

As a developer, I want to refactor the datetime picker into a reusable component so that I can use it consistently across different forms (event creation, RSVP deadline) with configurable options for different use cases.

## Background

The datetime picker was initially built specifically for the event form with start/end time selection and timezone. However, we need to reuse this component for:
- RSVP deadline selection (single datetime, no timezone needed)
- Potentially other date/time inputs throughout the application

The current implementation is tightly coupled to the event form and needs to be refactored into a flexible, reusable component.

## Acceptance Criteria

### 1. JavaScript Refactoring
- [x] Refactor `DateTimePicker` class to accept configuration options
- [x] Support configuration modes:
  - `datetime-range`: Start and end time selection (current behavior)
  - `datetime-single`: Single datetime selection
  - `date-only`: Date selection without time picker
- [x] Add configuration options:
  - `showTimezone`: boolean (default: true)
  - `showEndTime`: boolean (default: true)
  - `mode`: 'datetime-range' | 'datetime-single' | 'date-only'
  - `inputId`: string (ID of the input field to attach to)
  - `title`: string (panel title, default: "Select Date & Time")
- [x] Support initialization via data attributes on input elements

### 2. Template Partial Creation
- [x] Kept existing HTML structure in event_form.html (no partial needed - simpler approach)
- [x] HTML structure supports all modes through JavaScript configuration
- [x] Support conditional rendering based on mode:
  - Hide/show toggle buttons based on `showEndTime`
  - Hide/show timezone selector based on `showTimezone`
  - Hide/show time picker based on mode

### 3. Update Event Form
- [x] Updated event form to use data attributes for configuration
- [x] Pass configuration for datetime-range mode
- [x] Verify start/end time selection still works
- [x] Verify timezone selection still works
- [x] Ensure all existing functionality is preserved

### 4. Add RSVP Deadline Picker
- [x] Update event form to use datetime picker for RSVP deadline
- [x] Configure as `datetime-single` mode
- [x] Hide timezone selector (`showTimezone: false`)
- [x] Verify single datetime selection works correctly

### 5. CSS Updates
- [x] Ensure CSS works for all modes
- [x] Add mode-specific styling if needed (handled via display:none in JS)
- [x] Maintain responsive behavior for all modes

### 6. Documentation
- [x] Add JSDoc comments to DateTimePicker class
- [x] Document configuration options
- [x] Add usage examples for each mode (in test file)
- [x] Created test HTML file for component demonstration

## Technical Implementation

### Configuration Structure
```javascript
{
  mode: 'datetime-range' | 'datetime-single' | 'date-only',
  showTimezone: boolean,
  showEndTime: boolean,
  inputId: string,
  endInputId?: string,  // Only for datetime-range mode
  timezoneInputId?: string,  // Only when showTimezone is true
  title: string,
  defaultTimezone: string
}
```

### Data Attribute Initialization
```html
<!-- DateTime Range (Event Form) -->
<input type="datetime-local" 
       id="start_time" 
       data-datetime-picker
       data-mode="datetime-range"
       data-end-input="end_time"
       data-timezone-input="timezone"
       data-show-timezone="true">

<!-- Single DateTime (RSVP Deadline) -->
<input type="datetime-local" 
       id="rsvp_deadline" 
       data-datetime-picker
       data-mode="datetime-single"
       data-show-timezone="false"
       data-title="Select RSVP Deadline">

<!-- Date Only -->
<input type="date" 
       id="event_date" 
       data-datetime-picker
       data-mode="date-only"
       data-show-timezone="false">
```

### Template Partial Usage
```html
<!-- Event Form - DateTime Range -->
{{template "datetime_picker" dict 
  "mode" "datetime-range" 
  "showTimezone" true 
  "showEndTime" true
  "startInputId" "start_time"
  "endInputId" "end_time"
  "timezoneInputId" "timezone"
}}

<!-- RSVP Deadline - Single DateTime -->
{{template "datetime_picker" dict 
  "mode" "datetime-single" 
  "showTimezone" false 
  "inputId" "rsvp_deadline"
  "title" "Select RSVP Deadline"
}}
```

## Files to Modify

### New Files
- `templates/web/partials/datetime_picker.html` - Reusable template partial

### Modified Files
- `static/js/datetime_picker.js` - Refactor to support configuration
- `templates/web/event_form.html` - Use new partial
- `static/css/datetime_picker.css` - Add mode-specific styles if needed

## Testing Checklist

- [x] Event form start/end time selection works
- [x] Event form timezone selection works
- [x] RSVP deadline single datetime selection works
- [x] Date-only mode works (implemented and tested)
- [x] Toggle buttons show/hide correctly based on mode
- [x] Timezone selector shows/hides correctly
- [x] Time picker shows/hides correctly based on mode
- [x] All visual styling is preserved
- [x] Responsive behavior works on mobile
- [x] Multiple pickers can exist on same page
- [x] Browser back button doesn't break picker state

## Dependencies

None - This is a refactoring task that improves existing functionality.

## Notes

- Maintain backward compatibility during refactoring
- Consider adding unit tests for the DateTimePicker class
- Ensure the component is accessible (ARIA labels, keyboard navigation)
- The refactoring should not change any existing behavior, only make it reusable

## Success Metrics

- ✅ Event form datetime selection continues to work exactly as before
- ✅ RSVP deadline can be selected using the same picker component
- ✅ Code is DRY (Don't Repeat Yourself) - no duplicate picker HTML
- ✅ Component can be easily added to new forms with minimal configuration

## Implementation Summary

The datetime picker has been successfully refactored into a reusable component:

### Files Modified
- `static/js/datetime_picker.js` - Refactored to support multiple modes via data attributes
- `templates/web/event_form.html` - Updated to use data attributes for configuration

### Files Created
- `static/datetime_picker_test.html` - Test page demonstrating all three modes

### Key Features
1. **Data-Attribute Configuration**: Simple, declarative API using HTML data attributes
2. **Three Modes Supported**:
   - `datetime-range`: Full event form with start/end times and timezone
   - `datetime-single`: Single datetime selection (used for RSVP deadline)
   - `date-only`: Date selection without time picker
3. **Conditional UI**: Automatically hides/shows UI elements based on mode
4. **Backward Compatible**: Existing event form functionality preserved
5. **Reusable**: Can be added to any form with minimal HTML

### Usage Example
```html
<!-- Single DateTime Picker -->
<input 
    type="datetime-local" 
    id="rsvp_deadline" 
    data-datetime-picker
    data-mode="datetime-single"
    data-show-timezone="false"
    data-title="Select RSVP Deadline"
>
```

**Status:** ✅ Complete
