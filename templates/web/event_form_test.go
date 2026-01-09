package web

import (
	"strings"
	"testing"
)

func TestEventFormTemplate(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:    "contains form element",
			content: eventFormHTML,
			expected: []string{
				`<form`,
				`method="POST"`,
			},
		},
		{
			name:    "contains required fields",
			content: eventFormHTML,
			expected: []string{
				`name="title"`,
				`name="start_time"`,
				`name="timezone"`,
			},
		},
		{
			name:    "contains optional fields",
			content: eventFormHTML,
			expected: []string{
				`name="description"`,
				`name="location"`,
				`name="end_time"`,
				`name="rsvp_deadline"`,
				`name="max_plus_ones"`,
			},
		},
		{
			name:    "contains form labels",
			content: eventFormHTML,
			expected: []string{
				`class="form-label"`,
				`Event Title`,
				`Description`,
				`Start Date & Time`,
				`Timezone`,
			},
		},
		{
			name:    "contains required indicators",
			content: eventFormHTML,
			expected: []string{
				`class="form-required"`,
				`*`,
			},
		},
		{
			name:    "contains form validation classes",
			content: eventFormHTML,
			expected: []string{
				`class="form-group"`,
				`class="form-input"`,
				`class="form-textarea"`,
				`class="form-select"`,
			},
		},
		{
			name:    "contains submit buttons",
			content: eventFormHTML,
			expected: []string{
				`type="submit"`,
				`Save Draft`,
				`Publish Event`,
			},
		},
		{
			name:    "contains help text",
			content: eventFormHTML,
			expected: []string{
				`class="form-help-text"`,
			},
		},
		{
			name:    "contains error display",
			content: eventFormHTML,
			expected: []string{
				`class="form-error"`,
			},
		},
		{
			name:    "contains accessibility attributes",
			content: eventFormHTML,
			expected: []string{
				`aria-label`,
				`aria-describedby`,
				`required`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(tt.content, exp) {
					t.Errorf("Expected content to contain %q", exp)
				}
			}
		})
	}
}

func TestEventFormResponsive(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:    "contains responsive classes",
			content: eventFormHTML,
			expected: []string{
				`class="dashboard"`,
				`class="dashboard-main"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(tt.content, exp) {
					t.Errorf("Expected content to contain %q", exp)
				}
			}
		})
	}
}

func TestEventFormQuestions(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:    "contains preference questions section",
			content: eventFormHTML,
			expected: []string{
				`Preference Questions`,
				`Add Question`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(tt.content, exp) {
					t.Errorf("Expected content to contain %q", exp)
				}
			}
		})
	}
}

func TestEventFormTimezone(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:    "contains common timezones",
			content: eventFormHTML,
			expected: []string{
				`America/Los_Angeles`,
				`America/New_York`,
				`UTC`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(tt.content, exp) {
					t.Errorf("Expected content to contain %q", exp)
				}
			}
		})
	}
}

const eventFormHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Create Event - TinyRSVP</title>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/typography.css">
    <link rel="stylesheet" href="/static/css/colors.css">
    <link rel="stylesheet" href="/static/css/spacing.css">
    <link rel="stylesheet" href="/static/css/grid.css">
    <link rel="stylesheet" href="/static/css/buttons.css">
    <link rel="stylesheet" href="/static/css/navigation.css">
    <link rel="stylesheet" href="/static/css/forms.css">
    <link rel="stylesheet" href="/static/css/dashboard.css">
    <link rel="stylesheet" href="/static/css/event_form.css">
</head>
<body>
    <div class="dashboard">
        <aside class="dashboard-sidebar">
            <nav class="nav">
                <a href="/" class="logo">TinyRSVP</a>
                <ul class="nav-menu">
                    <li><a href="/dashboard" class="nav-link">Dashboard</a></li>
                    <li><a href="/events" class="nav-link active">Events</a></li>
                    <li><a href="/invites" class="nav-link">Invites</a></li>
                    <li><a href="/settings" class="nav-link">Settings</a></li>
                </ul>
            </nav>
        </aside>

        <main class="dashboard-main">
            <header class="dashboard-header">
                <h1>Create Event</h1>
            </header>

            <form method="POST" action="/events" class="event-form">
                <section class="form-section">
                    <h2 class="form-section-title">Event Details</h2>
                    
                    <div class="form-group">
                        <label for="title" class="form-label">
                            Event Title<span class="form-required" aria-label="required">*</span>
                        </label>
                        <input 
                            type="text" 
                            id="title" 
                            name="title" 
                            class="form-input" 
                            required
                            aria-describedby="title-help"
                            maxlength="200"
                        >
                        <span class="form-help-text" id="title-help">Give your event a clear, descriptive name</span>
                        <span class="form-error" id="title-error"></span>
                    </div>

                    <div class="form-group">
                        <label for="description" class="form-label">Description</label>
                        <textarea 
                            id="description" 
                            name="description" 
                            class="form-textarea"
                            aria-describedby="description-help"
                            rows="5"
                        ></textarea>
                        <span class="form-help-text" id="description-help">Provide details about your event</span>
                        <span class="form-error" id="description-error"></span>
                    </div>

                    <div class="form-group">
                        <label for="location" class="form-label">Location</label>
                        <input 
                            type="text" 
                            id="location" 
                            name="location" 
                            class="form-input"
                            aria-describedby="location-help"
                        >
                        <span class="form-help-text" id="location-help">Physical address or virtual meeting link</span>
                        <span class="form-error" id="location-error"></span>
                    </div>
                </section>

                <section class="form-section">
                    <h2 class="form-section-title">Date & Time</h2>
                    
                    <div class="form-group">
                        <label for="start_time" class="form-label">
                            Start Date & Time<span class="form-required" aria-label="required">*</span>
                        </label>
                        <input 
                            type="datetime-local" 
                            id="start_time" 
                            name="start_time" 
                            class="form-input" 
                            required
                            aria-describedby="start_time-help"
                        >
                        <span class="form-help-text" id="start_time-help">When does your event start?</span>
                        <span class="form-error" id="start_time-error"></span>
                    </div>

                    <div class="form-group">
                        <label for="end_time" class="form-label">End Date & Time</label>
                        <input 
                            type="datetime-local" 
                            id="end_time" 
                            name="end_time" 
                            class="form-input"
                            aria-describedby="end_time-help"
                        >
                        <span class="form-help-text" id="end_time-help">When does your event end?</span>
                        <span class="form-error" id="end_time-error"></span>
                    </div>

                    <div class="form-group">
                        <label for="timezone" class="form-label">
                            Timezone<span class="form-required" aria-label="required">*</span>
                        </label>
                        <select 
                            id="timezone" 
                            name="timezone" 
                            class="form-select" 
                            required
                            aria-describedby="timezone-help"
                        >
                            <option value="">Select timezone...</option>
                            <option value="America/Los_Angeles">Pacific Time (PT)</option>
                            <option value="America/Denver">Mountain Time (MT)</option>
                            <option value="America/Chicago">Central Time (CT)</option>
                            <option value="America/New_York">Eastern Time (ET)</option>
                            <option value="UTC">UTC</option>
                            <option value="Europe/London">London (GMT)</option>
                            <option value="Europe/Paris">Paris (CET)</option>
                            <option value="Asia/Tokyo">Tokyo (JST)</option>
                            <option value="Australia/Sydney">Sydney (AEDT)</option>
                        </select>
                        <span class="form-help-text" id="timezone-help">Event timezone</span>
                        <span class="form-error" id="timezone-error"></span>
                    </div>
                </section>

                <section class="form-section">
                    <h2 class="form-section-title">RSVP Settings</h2>
                    
                    <div class="form-group">
                        <label for="rsvp_deadline" class="form-label">RSVP Deadline</label>
                        <input 
                            type="datetime-local" 
                            id="rsvp_deadline" 
                            name="rsvp_deadline" 
                            class="form-input"
                            aria-describedby="rsvp_deadline-help"
                        >
                        <span class="form-help-text" id="rsvp_deadline-help">Last date guests can respond</span>
                        <span class="form-error" id="rsvp_deadline-error"></span>
                    </div>

                    <div class="form-group">
                        <label for="max_plus_ones" class="form-label">Maximum Plus Ones</label>
                        <input 
                            type="number" 
                            id="max_plus_ones" 
                            name="max_plus_ones" 
                            class="form-input"
                            min="0"
                            max="10"
                            value="0"
                            aria-describedby="max_plus_ones-help"
                        >
                        <span class="form-help-text" id="max_plus_ones-help">How many guests can each invitee bring?</span>
                        <span class="form-error" id="max_plus_ones-error"></span>
                    </div>
                </section>

                <section class="form-section">
                    <h2 class="form-section-title">Preference Questions</h2>
                    <p class="form-help-text">Ask guests about dietary restrictions, meal preferences, etc.</p>
                    
                    <div id="questions-container">
                    </div>
                    
                    <button type="button" class="btn btn-secondary" id="add-question-btn">
                        Add Question
                    </button>
                </section>

                <div class="form-actions">
                    <button type="submit" name="action" value="draft" class="btn btn-secondary">
                        Save Draft
                    </button>
                    <button type="submit" name="action" value="publish" class="btn btn-primary">
                        Publish Event
                    </button>
                </div>
            </form>
        </main>
    </div>
</body>
</html>`
