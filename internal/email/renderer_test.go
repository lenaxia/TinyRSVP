package email

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewTemplateRenderer_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `<html><body>Hello {{.Name}}</body></html>`
	if err := os.WriteFile(filepath.Join(tempDir, "test.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	txtContent := `Hello {{.Name}}`
	if err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte(txtContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}
	
	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}
}

func TestNewTemplateRenderer_EmptyTemplateDir(t *testing.T) {
	config := &TemplateConfig{
		TemplateDir:  "",
		CacheEnabled: true,
	}
	
	_, err := NewTemplateRenderer(config)
	if err == nil {
		t.Error("Expected error for empty template directory")
	}
	
	if !strings.Contains(err.Error(), "template directory is required") {
		t.Errorf("Expected error message about template directory, got: %v", err)
	}
}

func TestNewTemplateRenderer_InvalidTemplateDir(t *testing.T) {
	config := &TemplateConfig{
		TemplateDir:  "/nonexistent/path",
		CacheEnabled: true,
	}
	
	_, err := NewTemplateRenderer(config)
	if err == nil {
		t.Error("Expected error for invalid template directory")
	}
}

func TestTemplateRenderer_RenderHTML_Success(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `<html><body>Hello {{.Name}}, welcome to {{.Event}}</body></html>`
	if err := os.WriteFile(filepath.Join(tempDir, "greeting.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct {
		Name  string
		Event string
	}{
		Name:  "John Doe",
		Event: "Birthday Party",
	}
	
	html, err := renderer.RenderHTML(context.Background(), "greeting", data)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	
	if !strings.Contains(html, "John Doe") {
		t.Error("Rendered HTML missing expected name")
	}
	
	if !strings.Contains(html, "Birthday Party") {
		t.Error("Rendered HTML missing expected event")
	}
}

func TestTemplateRenderer_RenderHTML_MissingTemplate(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct{ Name string }{Name: "Test"}
	
	_, err = renderer.RenderHTML(context.Background(), "nonexistent", data)
	if err == nil {
		t.Error("Expected error for missing template")
	}
	
	if !strings.Contains(err.Error(), "template not found") {
		t.Errorf("Expected 'template not found' error, got: %v", err)
	}
}

func TestTemplateRenderer_RenderText_Success(t *testing.T) {
	tempDir := t.TempDir()
	
	txtContent := `Hello {{.Name}}, welcome to {{.Event}}`
	if err := os.WriteFile(filepath.Join(tempDir, "greeting.txt"), []byte(txtContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct {
		Name  string
		Event string
	}{
		Name:  "Jane Smith",
		Event: "Conference",
	}
	
	text, err := renderer.RenderText(context.Background(), "greeting", data)
	if err != nil {
		t.Fatalf("RenderText() error = %v", err)
	}
	
	if !strings.Contains(text, "Jane Smith") {
		t.Error("Rendered text missing expected name")
	}
	
	if !strings.Contains(text, "Conference") {
		t.Error("Rendered text missing expected event")
	}
}

func TestTemplateRenderer_RenderText_MissingTemplate(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct{ Name string }{Name: "Test"}
	
	_, err = renderer.RenderText(context.Background(), "nonexistent", data)
	if err == nil {
		t.Error("Expected error for missing template")
	}
	
	if !strings.Contains(err.Error(), "template not found") {
		t.Errorf("Expected 'template not found' error, got: %v", err)
	}
}

func TestTemplateRenderer_TemplateFunctions_FormatDate(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `{{formatDate .Date "2006-01-02"}}`
	if err := os.WriteFile(filepath.Join(tempDir, "date.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct {
		Date time.Time
	}{
		Date: time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC),
	}
	
	html, err := renderer.RenderHTML(context.Background(), "date", data)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	
	if !strings.Contains(html, "2026-01-15") {
		t.Errorf("Expected formatted date '2026-01-15', got: %s", html)
	}
}

func TestTemplateRenderer_TemplateFunctions_FormatDateTime(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `{{formatDateTime .Date}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datetime.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct {
		Date time.Time
	}{
		Date: time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC),
	}
	
	html, err := renderer.RenderHTML(context.Background(), "datetime", data)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	
	if !strings.Contains(html, "Thursday") {
		t.Errorf("Expected day name 'Thursday' in formatted datetime, got: %s", html)
	}
	
	if !strings.Contains(html, "January 15, 2026") {
		t.Errorf("Expected date in formatted datetime, got: %s", html)
	}
}

func TestTemplateRenderer_TemplateFunctions_Title(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `{{title .Text}}`
	if err := os.WriteFile(filepath.Join(tempDir, "title.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct {
		Text string
	}{
		Text: "hello world",
	}
	
	html, err := renderer.RenderHTML(context.Background(), "title", data)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	
	if !strings.Contains(html, "Hello World") {
		t.Errorf("Expected 'Hello World', got: %s", html)
	}
}

func TestTemplateRenderer_TemplateFunctions_Upper(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `{{upper .Text}}`
	if err := os.WriteFile(filepath.Join(tempDir, "upper.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct {
		Text string
	}{
		Text: "hello",
	}
	
	html, err := renderer.RenderHTML(context.Background(), "upper", data)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	
	if !strings.Contains(html, "HELLO") {
		t.Errorf("Expected 'HELLO', got: %s", html)
	}
}

func TestTemplateRenderer_TemplateFunctions_Lower(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `{{lower .Text}}`
	if err := os.WriteFile(filepath.Join(tempDir, "lower.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct {
		Text string
	}{
		Text: "HELLO",
	}
	
	html, err := renderer.RenderHTML(context.Background(), "lower", data)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	
	if !strings.Contains(html, "hello") {
		t.Errorf("Expected 'hello', got: %s", html)
	}
}

func TestTemplateRenderer_ReloadTemplates(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `<html><body>Version 1</body></html>`
	htmlFile := filepath.Join(tempDir, "reload.html")
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	html, err := renderer.RenderHTML(context.Background(), "reload", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	if !strings.Contains(html, "Version 1") {
		t.Error("Expected 'Version 1' in initial render")
	}
	
	htmlContent = `<html><body>Version 2</body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	if err := renderer.ReloadTemplates(); err != nil {
		t.Fatalf("ReloadTemplates() error = %v", err)
	}
	
	html, err = renderer.RenderHTML(context.Background(), "reload", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	if !strings.Contains(html, "Version 2") {
		t.Error("Expected 'Version 2' after reload")
	}
}

func TestTemplateRenderer_InvalidTemplateData(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `{{.MissingField}}`
	if err := os.WriteFile(filepath.Join(tempDir, "invalid.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	data := struct{ Name string }{Name: "Test"}
	
	_, err = renderer.RenderHTML(context.Background(), "invalid", data)
	if err == nil {
		t.Error("Expected error for invalid template data")
	}
	
	if !strings.Contains(err.Error(), "MissingField") {
		t.Errorf("Expected error about MissingField, got: %v", err)
	}
}

func TestTemplateRenderer_ConcurrentRendering(t *testing.T) {
	tempDir := t.TempDir()
	
	htmlContent := `<html><body>Hello {{.Name}}</body></html>`
	if err := os.WriteFile(filepath.Join(tempDir, "concurrent.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	config := &TemplateConfig{
		TemplateDir:  tempDir,
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatal(err)
	}
	
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			data := struct{ Name string }{Name: "User"}
			_, err := renderer.RenderHTML(context.Background(), "concurrent", data)
			if err != nil {
				t.Errorf("Concurrent render %d failed: %v", id, err)
			}
		}(i)
	}
	
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestTemplateRenderer_WithRealTemplates(t *testing.T) {
	if _, err := os.Stat("../../templates/email"); os.IsNotExist(err) {
		t.Skip("Real templates directory not found")
	}
	
	config := &TemplateConfig{
		TemplateDir:  "../../templates/email",
		CacheEnabled: true,
	}
	
	renderer, err := NewTemplateRenderer(config)
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}
	
	data := struct {
		GuestName     string
		Response      string
		PlusOnes      int
		EventTitle    string
		EventDate     string
		EventLocation string
		UpdateURL     string
		Answers       []struct {
			Question string
			Answer   string
		}
	}{
		GuestName:     "John Doe",
		Response:      "accepted",
		PlusOnes:      2,
		EventTitle:    "Birthday Party",
		EventDate:     "January 15, 2026 at 7:00 PM",
		EventLocation: "123 Main St",
		UpdateURL:     "https://example.com/update",
		Answers: []struct {
			Question string
			Answer   string
		}{
			{Question: "Dietary Restrictions", Answer: "Vegetarian"},
		},
	}
	
	html, err := renderer.RenderHTML(context.Background(), "rsvp_confirmation", data)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	
	if !strings.Contains(html, "John Doe") {
		t.Error("HTML missing guest name")
	}
	
	text, err := renderer.RenderText(context.Background(), "rsvp_confirmation", data)
	if err != nil {
		t.Fatalf("RenderText() error = %v", err)
	}
	
	if !strings.Contains(text, "John Doe") {
		t.Error("Text missing guest name")
	}
}
