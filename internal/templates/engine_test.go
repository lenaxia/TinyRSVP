package templates

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}
}

func TestEngine_Parse(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name     string
		tmplStr  string
		wantErr  bool
	}{
		{
			name:    "simple text",
			tmplStr: "<h1>Hello World</h1>",
			wantErr: false,
		},
		{
			name:    "with variable",
			tmplStr: "<h1>{{.Title}}</h1>",
			wantErr: false,
		},
		{
			name:    "with function",
			tmplStr: "<h1>{{.Title | upper}}</h1>",
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			tmplStr: "{{.Title",
			wantErr: true,
		},
		{
			name:    "empty template",
			tmplStr: "",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tt.tmplStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tmpl == nil {
				t.Error("Parse() returned nil template without error")
			}
		})
	}
}

func TestEngine_Execute(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name    string
		tmplStr string
		data    interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "simple text",
			tmplStr: "<h1>Hello World</h1>",
			data:    nil,
			want:    "<h1>Hello World</h1>",
			wantErr: false,
		},
		{
			name:    "variable substitution",
			tmplStr: "<h1>{{.Title}}</h1>",
			data:    struct{ Title string }{Title: "Test Event"},
			want:    "<h1>Test Event</h1>",
			wantErr: false,
		},
		{
			name:    "XSS prevention - script tag",
			tmplStr: "<p>{{.Description}}</p>",
			data:    struct{ Description string }{Description: "<script>alert('xss')</script>"},
			want:    "<p>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</p>",
			wantErr: false,
		},
		{
			name:    "XSS prevention - img onerror",
			tmplStr: "<div>{{.Content}}</div>",
			data:    struct{ Content string }{Content: "<img src=x onerror=alert('xss')>"},
			want:    "<div>&lt;img src=x onerror=alert(&#39;xss&#39;)&gt;</div>",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tt.tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			var buf bytes.Buffer
			err = engine.Execute(&buf, tmpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			got := buf.String()
			if got != tt.want {
				t.Errorf("Execute() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEngine_ExecuteToString(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name    string
		tmplStr string
		data    interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "simple text",
			tmplStr: "<h1>Hello World</h1>",
			data:    nil,
			want:    "<h1>Hello World</h1>",
			wantErr: false,
		},
		{
			name:    "variable substitution",
			tmplStr: "<h1>{{.Title}}</h1>",
			data:    struct{ Title string }{Title: "Test Event"},
			want:    "<h1>Test Event</h1>",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tt.tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			got, err := engine.ExecuteToString(tmpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if got != tt.want {
				t.Errorf("ExecuteToString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEngine_CustomFunctions_Upper(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("{{.Text | upper}}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	data := struct{ Text string }{Text: "hello world"}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	want := "HELLO WORLD"
	if got != want {
		t.Errorf("upper function = %q, want %q", got, want)
	}
}

func TestEngine_CustomFunctions_Lower(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("{{.Text | lower}}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	data := struct{ Text string }{Text: "HELLO WORLD"}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	want := "hello world"
	if got != want {
		t.Errorf("lower function = %q, want %q", got, want)
	}
}

func TestEngine_CustomFunctions_Title(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("{{.Text | title}}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	data := struct{ Text string }{Text: "hello world"}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	want := "Hello World"
	if got != want {
		t.Errorf("title function = %q, want %q", got, want)
	}
}

func TestEngine_CustomFunctions_FormatDate(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("{{formatDate .Time \"2006-01-02\"}}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	testTime := time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC)
	data := struct{ Time time.Time }{Time: testTime}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	want := "2026-06-15"
	if got != want {
		t.Errorf("formatDate function = %q, want %q", got, want)
	}
}

func TestEngine_CustomFunctions_FormatTime(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("{{formatTime .Time}}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	testTime := time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC)
	data := struct{ Time time.Time }{Time: testTime}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	want := "2:30 PM"
	if got != want {
		t.Errorf("formatTime function = %q, want %q", got, want)
	}
}

func TestEngine_CustomFunctions_FormatDateTime(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("{{formatDateTime .Time}}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	testTime := time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC)
	data := struct{ Time time.Time }{Time: testTime}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	want := "Monday, June 15, 2026 at 2:30 PM"
	if got != want {
		t.Errorf("formatDateTime function = %q, want %q", got, want)
	}
}

func TestEngine_CustomFunctions_Truncate(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name   string
		text   string
		length int
		want   string
	}{
		{
			name:   "shorter than limit",
			text:   "Hello",
			length: 10,
			want:   "Hello",
		},
		{
			name:   "exactly at limit",
			text:   "Hello",
			length: 5,
			want:   "Hello",
		},
		{
			name:   "longer than limit",
			text:   "Hello World",
			length: 5,
			want:   "Hello...",
		},
		{
			name:   "empty string",
			text:   "",
			length: 5,
			want:   "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse("{{truncate .Text .Length}}")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			data := struct {
				Text   string
				Length int
			}{Text: tt.text, Length: tt.length}
			
			got, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}
			
			if got != tt.want {
				t.Errorf("truncate function = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEngine_CustomFunctions_Default(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name         string
		value        string
		defaultValue string
		want         string
	}{
		{
			name:         "non-empty value",
			value:        "Hello",
			defaultValue: "Default",
			want:         "Hello",
		},
		{
			name:         "empty value",
			value:        "",
			defaultValue: "Default",
			want:         "Default",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse("{{default .Value .Default}}")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			data := struct {
				Value   string
				Default string
			}{Value: tt.value, Default: tt.defaultValue}
			
			got, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}
			
			if got != tt.want {
				t.Errorf("default function = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEngine_CustomFunctions_SafeHTML(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("{{safeHTML .HTML}}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	data := struct{ HTML string }{HTML: "<b>Bold</b>"}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	want := "<b>Bold</b>"
	if got != want {
		t.Errorf("safeHTML function = %q, want %q", got, want)
	}
}

func TestEngine_CustomFunctions_SafeURL(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("<a href=\"{{safeURL .URL}}\">Link</a>")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	data := struct{ URL string }{URL: "https://example.com/path?query=value"}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if !strings.Contains(got, "https://example.com/path?query=value") {
		t.Errorf("safeURL function did not preserve URL: %q", got)
	}
}

func TestEngine_CustomFunctions_SafeCSS(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("<style>{{safeCSS .CSS}}</style>")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	data := struct{ CSS string }{CSS: "color: red; font-size: 14px;"}
	got, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	want := "<style>color: red; font-size: 14px;</style>"
	if got != want {
		t.Errorf("safeCSS function = %q, want %q", got, want)
	}
}

func TestEngine_XSSPrevention(t *testing.T) {
	engine := NewEngine()
	
	xssPayloads := []struct {
		name    string
		payload string
	}{
		{
			name:    "script tag",
			payload: "<script>alert('xss')</script>",
		},
		{
			name:    "img onerror",
			payload: "<img src=x onerror=alert('xss')>",
		},
		{
			name:    "svg onload",
			payload: "<svg onload=alert('xss')>",
		},
		{
			name:    "javascript protocol",
			payload: "javascript:alert('xss')",
		},
		{
			name:    "iframe with javascript",
			payload: "<iframe src='javascript:alert(\"xss\")'></iframe>",
		},
		{
			name:    "onclick attribute",
			payload: "<div onclick='alert(\"xss\")'>Click</div>",
		},
	}
	
	tmplStr := "<div>{{.Payload}}</div>"
	
	for _, tt := range xssPayloads {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			data := struct{ Payload string }{Payload: tt.payload}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}
			
			if !strings.Contains(result, "&lt;") && !strings.Contains(result, "&#") {
				t.Errorf("Expected HTML escaping for %s: %s", tt.name, result)
			}
		})
	}
}

func TestEngine_ParseError(t *testing.T) {
	engine := NewEngine()
	
	invalidTemplates := []string{
		"{{.Title",
		"{{end}}",
		"{{range}}",
		"{{if}}",
	}
	
	for _, tmplStr := range invalidTemplates {
		t.Run(tmplStr, func(t *testing.T) {
			_, err := engine.Parse(tmplStr)
			if err == nil {
				t.Error("Parse() expected error for invalid template")
			}
		})
	}
}

func TestEngine_ExecuteError(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("{{.Method.NonExistent}}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	data := struct {
		Method string
	}{
		Method: "test",
	}
	
	var buf bytes.Buffer
	err = engine.Execute(&buf, tmpl, data)
	if err == nil {
		t.Error("Execute() expected error when accessing non-existent field")
	}
}

func TestEngine_NilTemplate(t *testing.T) {
	engine := NewEngine()
	
	var buf bytes.Buffer
	err := engine.Execute(&buf, nil, nil)
	if err == nil {
		t.Error("Execute() expected error for nil template")
	}
}

func TestEngine_NilWriter(t *testing.T) {
	engine := NewEngine()
	
	tmpl, err := engine.Parse("<h1>Test</h1>")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	err = engine.Execute(nil, tmpl, nil)
	if err == nil {
		t.Error("Execute() expected error for nil writer")
	}
}

func TestEngine_ComplexTemplate(t *testing.T) {
	engine := NewEngine()
	
	tmplStr := `
<html>
<head><title>{{.Title | upper}}</title></head>
<body>
	<h1>{{.Title}}</h1>
	<p>{{.Description}}</p>
	<p>Date: {{formatDate .Date "January 2, 2006"}}</p>
	<p>Time: {{formatTime .Date}}</p>
	{{if .Items}}
	<ul>
	{{range .Items}}
		<li>{{.}}</li>
	{{end}}
	</ul>
	{{end}}
</body>
</html>
`
	
	tmpl, err := engine.Parse(tmplStr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	data := struct {
		Title       string
		Description string
		Date        time.Time
		Items       []string
	}{
		Title:       "Test Event",
		Description: "A test event description",
		Date:        time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC),
		Items:       []string{"Item 1", "Item 2", "Item 3"},
	}
	
	result, err := engine.ExecuteToString(tmpl, data)
	if err != nil {
		t.Fatalf("ExecuteToString() error = %v", err)
	}
	
	if !strings.Contains(result, "TEST EVENT") {
		t.Error("Expected uppercase title")
	}
	if !strings.Contains(result, "Test Event") {
		t.Error("Expected original title")
	}
	if !strings.Contains(result, "June 15, 2026") {
		t.Error("Expected formatted date")
	}
	if !strings.Contains(result, "2:30 PM") {
		t.Error("Expected formatted time")
	}
	if !strings.Contains(result, "Item 1") || !strings.Contains(result, "Item 2") {
		t.Error("Expected items in list")
	}
}

func BenchmarkEngine_Parse(b *testing.B) {
	engine := NewEngine()
	tmplStr := "<h1>{{.Title}}</h1><p>{{.Description}}</p>"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Parse(tmplStr)
	}
}

func BenchmarkEngine_Execute(b *testing.B) {
	engine := NewEngine()
	tmpl, _ := engine.Parse("<h1>{{.Title}}</h1><p>{{.Description}}</p>")
	data := struct {
		Title       string
		Description string
	}{
		Title:       "Test Event",
		Description: "A test description",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = engine.Execute(&buf, tmpl, data)
	}
}

func BenchmarkEngine_ExecuteToString(b *testing.B) {
	engine := NewEngine()
	tmpl, _ := engine.Parse("<h1>{{.Title}}</h1><p>{{.Description}}</p>")
	data := struct {
		Title       string
		Description string
	}{
		Title:       "Test Event",
		Description: "A test description",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ExecuteToString(tmpl, data)
	}
}

func TestEngine_ThreadSafety(t *testing.T) {
	engine := NewEngine()
	tmpl, err := engine.Parse("<h1>{{.Title}}</h1>")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			data := struct{ Title string }{Title: "Test"}
			for j := 0; j < 100; j++ {
				_, err := engine.ExecuteToString(tmpl, data)
				if err != nil {
					t.Errorf("Goroutine %d: ExecuteToString() error = %v", id, err)
				}
			}
			done <- true
		}(i)
	}
	
	for i := 0; i < 10; i++ {
		<-done
	}
}
