package css

import (
	"os"
	"strings"
	"testing"
)

func TestCSSVariablesIntegrationWithTemplates(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	exampleHTML := `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/static/css/variables.css">
    <style>
        .button-primary {
            background-color: var(--color-primary-600);
            color: var(--color-background);
            padding: var(--spacing-3) var(--spacing-6);
            border-radius: var(--radius-md);
            font-size: var(--font-size-base);
            font-weight: var(--font-weight-medium);
            transition: all var(--transition-base);
            box-shadow: var(--shadow-sm);
        }
        
        .button-primary:hover {
            background-color: var(--color-primary-700);
            box-shadow: var(--shadow-md);
        }
        
        .card {
            background-color: var(--color-surface);
            border: 1px solid var(--color-border);
            border-radius: var(--radius-lg);
            padding: var(--spacing-6);
            box-shadow: var(--shadow-base);
        }
        
        .heading {
            font-size: var(--font-size-2xl);
            font-weight: var(--font-weight-bold);
            color: var(--color-text-primary);
            line-height: var(--line-height-tight);
            margin-bottom: var(--spacing-4);
        }
        
        .text-secondary {
            color: var(--color-text-secondary);
            font-size: var(--font-size-sm);
        }
    </style>
</head>
<body>
    <div class="card">
        <h1 class="heading">Event Title</h1>
        <p class="text-secondary">Event details go here</p>
        <button class="button-primary">RSVP Now</button>
    </div>
</body>
</html>
`

	requiredVarsInExample := []string{
		"--color-primary-600",
		"--color-primary-700",
		"--color-background",
		"--color-surface",
		"--color-border",
		"--color-text-primary",
		"--color-text-secondary",
		"--spacing-3",
		"--spacing-4",
		"--spacing-6",
		"--radius-md",
		"--radius-lg",
		"--font-size-base",
		"--font-size-sm",
		"--font-size-2xl",
		"--font-weight-medium",
		"--font-weight-bold",
		"--line-height-tight",
		"--transition-base",
		"--shadow-sm",
		"--shadow-md",
		"--shadow-base",
	}

	cssContent := string(content)
	for _, variable := range requiredVarsInExample {
		if !strings.Contains(cssContent, variable) {
			t.Errorf("Example HTML uses %s but it's not defined in variables.css", variable)
		}
	}

	if !strings.Contains(exampleHTML, "var(--") {
		t.Error("Example HTML should use CSS variables")
	}
}

func TestCSSVariablesResponsiveDesign(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	exampleResponsiveCSS := `
.container {
    width: 100%;
    max-width: var(--container-xl);
    margin: 0 auto;
    padding: var(--spacing-4);
}

@media (min-width: 768px) {
    .container {
        padding: var(--spacing-6);
    }
}

@media (min-width: 1024px) {
    .container {
        padding: var(--spacing-8);
    }
}

.grid {
    display: grid;
    gap: var(--spacing-4);
    grid-template-columns: 1fr;
}

@media (min-width: 768px) {
    .grid {
        grid-template-columns: repeat(2, 1fr);
        gap: var(--spacing-6);
    }
}

@media (min-width: 1024px) {
    .grid {
        grid-template-columns: repeat(3, 1fr);
        gap: var(--spacing-8);
    }
}
`

	requiredVarsInResponsive := []string{
		"--container-xl",
		"--spacing-4",
		"--spacing-6",
		"--spacing-8",
	}

	cssContent := string(content)
	for _, variable := range requiredVarsInResponsive {
		if !strings.Contains(cssContent, variable) {
			t.Errorf("Responsive example uses %s but it's not defined in variables.css", variable)
		}
	}

	if !strings.Contains(exampleResponsiveCSS, "@media") {
		t.Error("Responsive example should include media queries")
	}
}

func TestCSSVariablesFormStyling(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	_ = `
.form-input {
    width: 100%;
    padding: var(--spacing-3);
    font-size: var(--font-size-base);
    font-family: var(--font-family-sans);
    color: var(--color-text-primary);
    background-color: var(--color-background);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-base);
    transition: border-color var(--transition-base);
}

.form-input:focus {
    outline: none;
    border-color: var(--color-border-focus);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:disabled {
    background-color: var(--color-gray-100);
    color: var(--color-text-disabled);
    cursor: not-allowed;
}

.form-label {
    display: block;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    color: var(--color-text-primary);
    margin-bottom: var(--spacing-2);
}

.form-error {
    color: var(--color-error);
    font-size: var(--font-size-sm);
    margin-top: var(--spacing-1);
}

.form-success {
    color: var(--color-success);
    font-size: var(--font-size-sm);
    margin-top: var(--spacing-1);
}
`

	requiredVarsInForm := []string{
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--font-size-base",
		"--font-size-sm",
		"--font-family-sans",
		"--font-weight-medium",
		"--color-text-primary",
		"--color-text-disabled",
		"--color-background",
		"--color-border",
		"--color-border-focus",
		"--color-gray-100",
		"--color-error",
		"--color-success",
		"--radius-base",
		"--transition-base",
	}

	cssContent := string(content)
	for _, variable := range requiredVarsInForm {
		if !strings.Contains(cssContent, variable) {
			t.Errorf("Form example uses %s but it's not defined in variables.css", variable)
		}
	}
}

func TestCSSVariablesAlertComponents(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	_ = `
.alert {
    padding: var(--spacing-4);
    border-radius: var(--radius-md);
    border-left: 4px solid;
    font-size: var(--font-size-sm);
}

.alert-success {
    background-color: var(--color-success-light);
    border-color: var(--color-success);
    color: var(--color-success);
}

.alert-warning {
    background-color: var(--color-warning-light);
    border-color: var(--color-warning);
    color: var(--color-warning);
}

.alert-error {
    background-color: var(--color-error-light);
    border-color: var(--color-error);
    color: var(--color-error);
}

.alert-info {
    background-color: var(--color-info-light);
    border-color: var(--color-info);
    color: var(--color-info);
}
`

	requiredVarsInAlert := []string{
		"--spacing-4",
		"--radius-md",
		"--font-size-sm",
		"--color-success",
		"--color-success-light",
		"--color-warning",
		"--color-warning-light",
		"--color-error",
		"--color-error-light",
		"--color-info",
		"--color-info-light",
	}

	cssContent := string(content)
	for _, variable := range requiredVarsInAlert {
		if !strings.Contains(cssContent, variable) {
			t.Errorf("Alert example uses %s but it's not defined in variables.css", variable)
		}
	}
}

func TestCSSVariablesModalComponents(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	_ = `
.modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    z-index: var(--z-index-modal-backdrop);
}

.modal {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background-color: var(--color-background);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-2xl);
    padding: var(--spacing-6);
    max-width: var(--container-md);
    width: 90%;
    z-index: var(--z-index-modal);
}

.modal-header {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    color: var(--color-text-primary);
    margin-bottom: var(--spacing-4);
}

.modal-body {
    color: var(--color-text-secondary);
    line-height: var(--line-height-normal);
    margin-bottom: var(--spacing-6);
}
`

	requiredVarsInModal := []string{
		"--z-index-modal-backdrop",
		"--z-index-modal",
		"--color-background",
		"--color-text-primary",
		"--color-text-secondary",
		"--radius-xl",
		"--shadow-2xl",
		"--spacing-4",
		"--spacing-6",
		"--font-size-xl",
		"--font-weight-bold",
		"--line-height-normal",
		"--container-md",
	}

	cssContent := string(content)
	for _, variable := range requiredVarsInModal {
		if !strings.Contains(cssContent, variable) {
			t.Errorf("Modal example uses %s but it's not defined in variables.css", variable)
		}
	}
}

func TestCSSVariablesNavigationComponents(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	_ = `
.navbar {
    position: sticky;
    top: 0;
    background-color: var(--color-background);
    border-bottom: 1px solid var(--color-border);
    padding: var(--spacing-4) var(--spacing-6);
    z-index: var(--z-index-sticky);
    box-shadow: var(--shadow-sm);
}

.nav-link {
    color: var(--color-text-secondary);
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-medium);
    padding: var(--spacing-2) var(--spacing-4);
    border-radius: var(--radius-base);
    transition: all var(--transition-fast);
}

.nav-link:hover {
    color: var(--color-text-primary);
    background-color: var(--color-gray-100);
}

.nav-link.active {
    color: var(--color-primary-600);
    background-color: var(--color-primary-50);
}
`

	requiredVarsInNav := []string{
		"--color-background",
		"--color-border",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-primary-50",
		"--color-primary-600",
		"--color-gray-100",
		"--spacing-2",
		"--spacing-4",
		"--spacing-6",
		"--font-size-base",
		"--font-weight-medium",
		"--radius-base",
		"--shadow-sm",
		"--z-index-sticky",
		"--transition-fast",
	}

	cssContent := string(content)
	for _, variable := range requiredVarsInNav {
		if !strings.Contains(cssContent, variable) {
			t.Errorf("Navigation example uses %s but it's not defined in variables.css", variable)
		}
	}
}

func TestCSSVariablesCompleteCoverage(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	cssContent := string(content)

	allRequiredCategories := map[string][]string{
		"Primary Colors": {
			"--color-primary-50", "--color-primary-500", "--color-primary-900",
		},
		"Semantic Colors": {
			"--color-success", "--color-warning", "--color-error", "--color-info",
		},
		"Gray Scale": {
			"--color-gray-50", "--color-gray-500", "--color-gray-900",
		},
		"Functional Colors": {
			"--color-background", "--color-text-primary", "--color-border",
		},
		"Spacing": {
			"--spacing-0", "--spacing-4", "--spacing-8", "--spacing-16",
		},
		"Typography": {
			"--font-size-base", "--font-weight-normal", "--line-height-normal", "--font-family-sans",
		},
		"Border Radius": {
			"--radius-none", "--radius-base", "--radius-lg", "--radius-full",
		},
		"Shadows": {
			"--shadow-sm", "--shadow-base", "--shadow-lg",
		},
		"Transitions": {
			"--transition-fast", "--transition-base", "--transition-slow",
		},
		"Z-Index": {
			"--z-index-dropdown", "--z-index-modal", "--z-index-tooltip",
		},
		"Breakpoints": {
			"--breakpoint-sm", "--breakpoint-md", "--breakpoint-lg",
		},
		"Containers": {
			"--container-sm", "--container-md", "--container-lg",
		},
	}

	for category, variables := range allRequiredCategories {
		for _, variable := range variables {
			if !strings.Contains(cssContent, variable) {
				t.Errorf("Category '%s': missing variable %s", category, variable)
			}
		}
	}
}
