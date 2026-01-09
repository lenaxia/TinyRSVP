package css

import (
	"strings"
	"testing"
)

func TestEventFormCSS(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:    "contains event form class",
			content: eventFormCSS,
			expected: []string{
				".event-form",
			},
		},
		{
			name:    "contains form section styles",
			content: eventFormCSS,
			expected: []string{
				".form-section",
				".form-section-title",
			},
		},
		{
			name:    "contains form actions styles",
			content: eventFormCSS,
			expected: []string{
				".form-actions",
			},
		},
		{
			name:    "contains question item styles",
			content: eventFormCSS,
			expected: []string{
				".question-item",
			},
		},
		{
			name:    "contains responsive styles",
			content: eventFormCSS,
			expected: []string{
				"@media",
				"min-width",
			},
		},
		{
			name:    "contains spacing variables",
			content: eventFormCSS,
			expected: []string{
				"var(--spacing-",
			},
		},
		{
			name:    "contains alert styles",
			content: eventFormCSS,
			expected: []string{
				".alert",
				".alert-error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(tt.content, exp) {
					t.Errorf("Expected CSS to contain %q", exp)
				}
			}
		})
	}
}

func TestEventFormCSSResponsive(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:    "contains tablet breakpoint",
			content: eventFormCSS,
			expected: []string{
				"768px",
			},
		},
		{
			name:    "contains desktop breakpoint",
			content: eventFormCSS,
			expected: []string{
				"1024px",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(tt.content, exp) {
					t.Errorf("Expected CSS to contain %q", exp)
				}
			}
		})
	}
}

const eventFormCSS = `.event-form {
    max-width: 800px;
    margin: 0 auto;
}

.form-section {
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    padding: var(--spacing-6);
    margin-bottom: var(--spacing-6);
}

.form-section-title {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-semibold);
    color: var(--color-text-primary);
    margin-bottom: var(--spacing-4);
    padding-bottom: var(--spacing-3);
    border-bottom: 2px solid var(--color-border);
}

.form-actions {
    display: flex;
    gap: var(--spacing-3);
    justify-content: flex-end;
    padding: var(--spacing-6) 0;
    border-top: 1px solid var(--color-border);
    margin-top: var(--spacing-6);
}

.question-item {
    background: var(--color-background);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: var(--spacing-4);
    margin-bottom: var(--spacing-4);
}

.question-item .form-group:last-of-type {
    margin-bottom: var(--spacing-3);
}

.remove-question {
    margin-top: var(--spacing-2);
}

.alert {
    padding: var(--spacing-4);
    border-radius: var(--radius-md);
    margin-bottom: var(--spacing-6);
    border: 1px solid;
}

.alert-error {
    background-color: var(--color-error-bg);
    border-color: var(--color-error);
    color: var(--color-error-text);
}

.alert strong {
    font-weight: var(--font-weight-semibold);
}

@media (max-width: 767px) {
    .form-section {
        padding: var(--spacing-4);
        margin-bottom: var(--spacing-4);
    }

    .form-actions {
        flex-direction: column;
    }

    .form-actions .btn {
        width: 100%;
    }
}

@media (min-width: 768px) {
    .form-inline {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: var(--spacing-4);
    }
}

@media (min-width: 1024px) {
    .event-form {
        max-width: 900px;
    }
}
`
