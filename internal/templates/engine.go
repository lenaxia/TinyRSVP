package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"
)

type Engine struct {
	funcMap template.FuncMap
}

func NewEngine() *Engine {
	return &Engine{
		funcMap: createFuncMap(),
	}
}

func (e *Engine) Parse(templateContent string) (*template.Template, error) {
	if templateContent == "" {
		return template.New("").Funcs(e.funcMap).Parse(templateContent)
	}
	
	tmpl, err := template.New("").Funcs(e.funcMap).Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	
	return tmpl, nil
}

func (e *Engine) Execute(w io.Writer, tmpl *template.Template, data interface{}) error {
	if w == nil {
		return fmt.Errorf("writer cannot be nil")
	}
	
	if tmpl == nil {
		return fmt.Errorf("template cannot be nil")
	}
	
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	
	return nil
}

func (e *Engine) ExecuteToString(tmpl *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := e.Execute(&buf, tmpl, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func createFuncMap() template.FuncMap {
	return template.FuncMap{
		"upper":          strings.ToUpper,
		"lower":          strings.ToLower,
		"title":          strings.Title,
		"formatDate":     formatDate,
		"formatTime":     formatTime,
		"formatDateTime": formatDateTime,
		"truncate":       truncate,
		"default":        defaultValue,
		"iterate":        iterate,
		"add":            add,
	}
}

func formatDate(t time.Time, layout string) string {
	return t.Format(layout)
}

func formatTime(t time.Time) string {
	return t.Format("3:04 PM")
}

func formatDateTime(t time.Time) string {
	return t.Format("Monday, January 2, 2006 at 3:04 PM")
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

func defaultValue(value, defaultVal string) string {
	if value == "" {
		return defaultVal
	}
	return value
}

func iterate(count int) []int {
	result := make([]int, count)
	for i := 0; i < count; i++ {
		result[i] = i
	}
	return result
}

func add(a, b int) int {
	return a + b
}
