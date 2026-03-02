package email

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
	textTemplate "text/template"
	"time"
)

type TemplateRenderer interface {
	RenderHTML(ctx context.Context, templateName string, data interface{}) (string, error)
	RenderText(ctx context.Context, templateName string, data interface{}) (string, error)
	LoadTemplates() error
	ReloadTemplates() error
}

type TemplateConfig struct {
	TemplateDir  string
	CacheEnabled bool
}

type templateRenderer struct {
	config        *TemplateConfig
	htmlTemplates map[string]*template.Template
	textTemplates map[string]*textTemplate.Template
	mu            sync.RWMutex
}

func NewTemplateRenderer(config *TemplateConfig) (TemplateRenderer, error) {
	if config.TemplateDir == "" {
		return nil, fmt.Errorf("template directory is required")
	}

	if _, err := os.Stat(config.TemplateDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("template directory does not exist: %s", config.TemplateDir)
	}

	r := &templateRenderer{
		config:        config,
		htmlTemplates: make(map[string]*template.Template),
		textTemplates: make(map[string]*textTemplate.Template),
	}

	if err := r.LoadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return r, nil
}

func (r *templateRenderer) RenderHTML(ctx context.Context, templateName string, data interface{}) (string, error) {
	r.mu.RLock()
	tmpl, ok := r.htmlTemplates[templateName]
	r.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("HTML template not found: %s", templateName)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.String(), nil
}

func (r *templateRenderer) RenderText(ctx context.Context, templateName string, data interface{}) (string, error) {
	r.mu.RLock()
	tmpl, ok := r.textTemplates[templateName]
	r.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("text template not found: %s", templateName)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return buf.String(), nil
}

func (r *templateRenderer) LoadTemplates() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	htmlPattern := filepath.Join(r.config.TemplateDir, "*.html")
	htmlFiles, err := filepath.Glob(htmlPattern)
	if err != nil {
		return fmt.Errorf("failed to glob HTML templates: %w", err)
	}

	for _, file := range htmlFiles {
		name := filepath.Base(file)
		name = strings.TrimSuffix(name, ".html")

		tmpl, err := template.New(filepath.Base(file)).Funcs(templateFuncs()).ParseFiles(file)
		if err != nil {
			return fmt.Errorf("failed to parse HTML template %s: %w", file, err)
		}

		r.htmlTemplates[name] = tmpl
	}

	textPattern := filepath.Join(r.config.TemplateDir, "*.txt")
	textFiles, err := filepath.Glob(textPattern)
	if err != nil {
		return fmt.Errorf("failed to glob text templates: %w", err)
	}

	for _, file := range textFiles {
		name := filepath.Base(file)
		name = strings.TrimSuffix(name, ".txt")

		tmpl, err := textTemplate.New(filepath.Base(file)).Funcs(textTemplateFuncs()).ParseFiles(file)
		if err != nil {
			return fmt.Errorf("failed to parse text template %s: %w", file, err)
		}

		r.textTemplates[name] = tmpl
	}

	return nil
}

func (r *templateRenderer) ReloadTemplates() error {
	return r.LoadTemplates()
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatDate": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("Monday, January 2, 2006 at 3:04 PM")
		},
		"title": strings.Title,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}
}

func textTemplateFuncs() textTemplate.FuncMap {
	return textTemplate.FuncMap{
		"formatDate": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("Monday, January 2, 2006 at 3:04 PM")
		},
		"title": strings.Title,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}
}
