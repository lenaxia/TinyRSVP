package uxserver

import (
	"html/template"
	"strings"
	"testing"
)

func TestBuildTemplateFuncMap_Dict(t *testing.T) {
	funcMap := BuildTemplateFuncMap()
	dictFn, ok := funcMap["dict"]
	if !ok {
		t.Fatal("expected 'dict' function in funcMap")
	}

	tests := []struct {
		name      string
		args      []interface{}
		wantKeys  []string
		wantError bool
	}{
		{
			name:     "valid string key/value pairs",
			args:     []interface{}{"a", 1, "b", "two", "c", true},
			wantKeys: []string{"a", "b", "c"},
		},
		{
			name:     "single pair",
			args:     []interface{}{"only", 42},
			wantKeys: []string{"only"},
		},
		{
			name:      "odd number of args returns error",
			args:      []interface{}{"a", 1, "b"},
			wantError: true,
		},
		{
			name:      "non-string key returns error",
			args:      []interface{}{123, "value"},
			wantError: true,
		},
		{
			name:     "empty args returns empty map",
			args:     []interface{}{},
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dictFn.(func(...interface{}) (map[string]interface{}, error))
			m, err := result(tt.args...)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(m) != len(tt.wantKeys) {
				t.Fatalf("map length = %d, want %d (keys: %v)", len(m), len(tt.wantKeys), tt.wantKeys)
			}
			for _, k := range tt.wantKeys {
				if _, ok := m[k]; !ok {
					t.Errorf("expected key %q in map", k)
				}
			}
		})
	}
}

func TestBuildTemplateFuncMap_DictValueIntegrity(t *testing.T) {
	funcMap := BuildTemplateFuncMap()
	dictFn := funcMap["dict"].(func(...interface{}) (map[string]interface{}, error))

	m, err := dictFn("count", 42, "name", "Alice", "active", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m["count"] != 42 {
		t.Errorf("count = %v, want 42", m["count"])
	}
	if m["name"] != "Alice" {
		t.Errorf("name = %v, want Alice", m["name"])
	}
	if m["active"] != true {
		t.Errorf("active = %v, want true", m["active"])
	}
}

func TestBuildTemplateFuncMap_IncludesRequiredFunctions(t *testing.T) {
	funcMap := BuildTemplateFuncMap()

	required := []string{
		"add", "sub", "mul", "div",
		"until", "iterate",
		"lower", "upper",
		"formatDateTime", "formatTime",
		"dict", "safeHTML",
		"timezoneAbbr",
	}

	for _, name := range required {
		if _, ok := funcMap[name]; !ok {
			t.Errorf("expected function %q in funcMap", name)
		}
	}
}

func TestBuildTemplateFuncMap_DivByZero(t *testing.T) {
	funcMap := BuildTemplateFuncMap()
	divFn := funcMap["div"].(func(a, b int) int)

	if result := divFn(10, 0); result != 0 {
		t.Errorf("div(10, 0) = %d, want 0 (safe default)", result)
	}
	if result := divFn(10, 2); result != 5 {
		t.Errorf("div(10, 2) = %d, want 5", result)
	}
}

func TestBuildTemplateFuncMap_TemplateExecutable(t *testing.T) {
	funcMap := BuildTemplateFuncMap()

	tmpl, err := template.New("test").Funcs(funcMap).Parse(
		`{{range $k, $v := (dict "Title" "Hello" "Count" 5)}}{{$k}}={{$v}} {{end}}`,
	)
	if err != nil {
		t.Fatalf("parse template with funcMap: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("execute template with funcMap: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Title=Hello") {
		t.Errorf("template output missing Title=Hello: %q", output)
	}
	if !strings.Contains(output, "Count=5") {
		t.Errorf("template output missing Count=5: %q", output)
	}
}

func TestBuildTemplateFuncMap_TimezoneAbbr(t *testing.T) {
	funcMap := BuildTemplateFuncMap()
	tzFn := funcMap["timezoneAbbr"].(func(string) string)

	result := tzFn("America/Los_Angeles")
	if result == "" {
		t.Error("timezoneAbbr returned empty string for valid timezone")
	}
	if !isValidTimezoneAbbr(result) {
		t.Logf("timezoneAbbr returned %q (may be IANA fallback if timezone DB unavailable)", result)
	}

	invalid := tzFn("Not/A/Real_Tz")
	if invalid != "Not/A/Real_Tz" {
		t.Errorf("timezoneAbbr for invalid tz = %q, want fallback to input", invalid)
	}
}

func isValidTimezoneAbbr(s string) bool {
	if s == "" {
		return false
	}
	upper := strings.ToUpper(s)
	return strings.Contains(upper, "PST") ||
		strings.Contains(upper, "PDT") ||
		strings.Contains(upper, "MST") ||
		strings.Contains(upper, "MDT") ||
		strings.Contains(upper, "CST") ||
		strings.Contains(upper, "CDT") ||
		strings.Contains(upper, "EST") ||
		strings.Contains(upper, "EDT") ||
		len(s) <= 5
}

func TestBuildTemplateFuncMap_Dict_DuplicateKeys(t *testing.T) {
	funcMap := BuildTemplateFuncMap()
	dictFn := funcMap["dict"].(func(...interface{}) (map[string]interface{}, error))

	m, err := dictFn("k", "first", "k", "second", "k", "third")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v, ok := m["k"]; !ok || v != "third" {
		t.Errorf("duplicate key last-wins: got %v (ok=%v), want \"third\"", v, ok)
	}
	if len(m) != 1 {
		t.Errorf("map length = %d, want 1 (single key after duplicates)", len(m))
	}
}

func TestBuildTemplateFuncMap_Div_NegativeAndOverflow(t *testing.T) {
	funcMap := BuildTemplateFuncMap()
	divFn := funcMap["div"].(func(a, b int) int)

	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"positive/positive", 10, 3, 3},
		{"negative/positive", -10, 3, -3},
		{"positive/negative", 10, -3, -3},
		{"negative/negative", -10, -3, 3},
		{"zero dividend", 0, 5, 0},
		{"zero divisor", 5, 0, 0},
		{"both zero", 0, 0, 0},
		{"large dividend", 1 << 30, 2, 1 << 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := divFn(tt.a, tt.b); got != tt.want {
				t.Errorf("div(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestBuildTemplateFuncMap_TimezoneAbbr_KnownZones(t *testing.T) {
	funcMap := BuildTemplateFuncMap()
	tzFn := funcMap["timezoneAbbr"].(func(string) string)

	knownZones := []string{
		"America/Los_Angeles",
		"America/New_York",
		"America/Chicago",
		"Europe/London",
		"Europe/Paris",
		"Asia/Tokyo",
		"UTC",
	}

	for _, zone := range knownZones {
		t.Run(zone, func(t *testing.T) {
			result := tzFn(zone)
			if result == "" {
				t.Errorf("timezoneAbbr(%q) returned empty", zone)
			}
			if result == zone {
				t.Logf("timezoneAbbr(%q) returned the input unchanged — timezone DB may be unavailable in this environment", zone)
			}
		})
	}
}
