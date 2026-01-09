package web

import (
	"html/template"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLoadingStatesTemplateIntegration(t *testing.T) {
	cssContent, err := os.ReadFile("../../static/css/loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	jsContent, err := os.ReadFile("../../static/js/loading_states.js")
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	if len(cssContent) == 0 {
		t.Fatal("loading_states.css is empty")
	}

	if len(jsContent) == 0 {
		t.Fatal("loading_states.js is empty")
	}
}

func TestLoadingStatesInButtonTemplate(t *testing.T) {
	tmplContent := `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/buttons.css">
    <link rel="stylesheet" href="/static/css/loading_states.css">
</head>
<body>
    <button id="submit-btn" class="btn btn-primary">Submit</button>
    <script src="/static/js/loading_states.js"></script>
    <script>
        document.getElementById('submit-btn').addEventListener('click', function() {
            LoadingStates.showButtonLoading(this);
            setTimeout(() => LoadingStates.hideButtonLoading(this), 2000);
        });
    </script>
</body>
</html>
`

	tmpl, err := template.New("test").Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	w := httptest.NewRecorder()
	err = tmpl.Execute(w, nil)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := w.Body.String()

	if !strings.Contains(html, "loading_states.css") {
		t.Error("Template should include loading_states.css")
	}

	if !strings.Contains(html, "loading_states.js") {
		t.Error("Template should include loading_states.js")
	}

	if !strings.Contains(html, "LoadingStates.showButtonLoading") {
		t.Error("Template should use LoadingStates API")
	}
}

func TestLoadingStatesInFormTemplate(t *testing.T) {
	tmplContent := `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/forms.css">
    <link rel="stylesheet" href="/static/css/loading_states.css">
</head>
<body>
    <form id="event-form">
        <div class="form-group">
            <label class="form-label">Event Name</label>
            <input type="text" class="form-input" name="name" required>
        </div>
        <button type="submit" class="btn btn-primary">Create Event</button>
    </form>
    <script src="/static/js/loading_states.js"></script>
    <script>
        document.getElementById('event-form').addEventListener('submit', function(e) {
            e.preventDefault();
            const btn = this.querySelector('button[type="submit"]');
            LoadingStates.showButtonLoading(btn);
        });
    </script>
</body>
</html>
`

	tmpl, err := template.New("test").Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	w := httptest.NewRecorder()
	err = tmpl.Execute(w, nil)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := w.Body.String()

	if !strings.Contains(html, "loading_states.css") {
		t.Error("Form template should include loading_states.css")
	}

	if !strings.Contains(html, "LoadingStates.showButtonLoading") {
		t.Error("Form template should use LoadingStates for submit button")
	}
}

func TestLoadingStatesSkeletonInListTemplate(t *testing.T) {
	tmplContent := `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/loading_states.css">
</head>
<body>
    <div id="event-list">
        {{if .Loading}}
            <div class="skeleton skeleton-heading"></div>
            <div class="skeleton skeleton-text"></div>
            <div class="skeleton skeleton-text"></div>
            <div class="skeleton skeleton-text" style="width: 60%;"></div>
        {{else}}
            {{range .Events}}
                <div class="event-item">{{.Name}}</div>
            {{end}}
        {{end}}
    </div>
</body>
</html>
`

	tmpl, err := template.New("test").Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	t.Run("loading state", func(t *testing.T) {
		data := struct {
			Loading bool
			Events  []struct{ Name string }
		}{
			Loading: true,
			Events:  nil,
		}

		w := httptest.NewRecorder()
		err = tmpl.Execute(w, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := w.Body.String()

		if !strings.Contains(html, "skeleton") {
			t.Error("Loading state should show skeleton screens")
		}
	})

	t.Run("loaded state", func(t *testing.T) {
		data := struct {
			Loading bool
			Events  []struct{ Name string }
		}{
			Loading: false,
			Events: []struct{ Name string }{
				{Name: "Event 1"},
				{Name: "Event 2"},
			},
		}

		w := httptest.NewRecorder()
		err = tmpl.Execute(w, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := w.Body.String()

		if strings.Contains(html, "skeleton") {
			t.Error("Loaded state should not show skeleton screens")
		}

		if !strings.Contains(html, "Event 1") {
			t.Error("Loaded state should show actual content")
		}
	})
}

func TestLoadingStatesProgressInTemplate(t *testing.T) {
	tmplContent := `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/loading_states.css">
</head>
<body>
    <div class="progress">
        <div id="upload-progress" class="progress-bar" style="width: {{.Progress}}%;" role="progressbar" aria-valuenow="{{.Progress}}" aria-valuemin="0" aria-valuemax="100"></div>
    </div>
    <script src="/static/js/loading_states.js"></script>
</body>
</html>
`

	tmpl, err := template.New("test").Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := struct {
		Progress int
	}{
		Progress: 65,
	}

	w := httptest.NewRecorder()
	err = tmpl.Execute(w, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := w.Body.String()

	if !strings.Contains(html, "progress-bar") {
		t.Error("Template should include progress bar")
	}

	if !strings.Contains(html, "65%") {
		t.Error("Template should render progress percentage")
	}

	if !strings.Contains(html, "aria-valuenow=\"65\"") {
		t.Error("Template should include ARIA progress value")
	}
}

func TestLoadingStatesOverlayInTemplate(t *testing.T) {
	tmplContent := `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/loading_states.css">
</head>
<body>
    <button id="action-btn" class="btn btn-primary">Perform Action</button>
    <script src="/static/js/loading_states.js"></script>
    <script>
        document.getElementById('action-btn').addEventListener('click', function() {
            LoadingStates.showOverlay({label: 'Processing request'});
            setTimeout(() => LoadingStates.hideOverlay(), 2000);
        });
    </script>
</body>
</html>
`

	tmpl, err := template.New("test").Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	w := httptest.NewRecorder()
	err = tmpl.Execute(w, nil)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := w.Body.String()

	if !strings.Contains(html, "LoadingStates.showOverlay") {
		t.Error("Template should use LoadingStates.showOverlay")
	}

	if !strings.Contains(html, "LoadingStates.hideOverlay") {
		t.Error("Template should use LoadingStates.hideOverlay")
	}
}

func TestLoadingStatesARIAInTemplate(t *testing.T) {
	tmplContent := `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/static/css/loading_states.css">
</head>
<body>
    <div id="content" aria-busy="false">
        <p>Content here</p>
    </div>
    <script src="/static/js/loading_states.js"></script>
    <script>
        LoadingStates.setLoadingState('#content', true);
    </script>
</body>
</html>
`

	tmpl, err := template.New("test").Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	w := httptest.NewRecorder()
	err = tmpl.Execute(w, nil)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := w.Body.String()

	if !strings.Contains(html, "aria-busy") {
		t.Error("Template should include aria-busy attribute")
	}

	if !strings.Contains(html, "LoadingStates.setLoadingState") {
		t.Error("Template should use LoadingStates.setLoadingState")
	}
}

func TestLoadingStatesMultipleComponentsInTemplate(t *testing.T) {
	tmplContent := `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/buttons.css">
    <link rel="stylesheet" href="/static/css/loading_states.css">
</head>
<body>
    <button id="btn1" class="btn btn-primary">Button 1</button>
    <button id="btn2" class="btn btn-secondary">Button 2</button>
    <div id="spinner-container"></div>
    <div class="progress">
        <div id="progress1" class="progress-bar" style="width: 0%;"></div>
    </div>
    <script src="/static/js/loading_states.js"></script>
</body>
</html>
`

	tmpl, err := template.New("test").Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	w := httptest.NewRecorder()
	err = tmpl.Execute(w, nil)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := w.Body.String()

	components := []string{
		"btn btn-primary",
		"btn btn-secondary",
		"progress-bar",
		"loading_states.js",
		"loading_states.css",
	}

	for _, component := range components {
		if !strings.Contains(html, component) {
			t.Errorf("Template should include component: %s", component)
		}
	}
}
