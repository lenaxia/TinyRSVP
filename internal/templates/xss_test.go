package templates

import (
	"strings"
	"testing"
)

var owaspXSSVectors = []struct {
	name    string
	payload string
}{
	{"basic script tag", "<script>alert('xss')</script>"},
	{"img onerror", "<img src=x onerror=alert('xss')>"},
	{"svg onload", "<svg onload=alert('xss')>"},
	{"body onload", "<body onload=alert('xss')>"},
	{"input onfocus autofocus", "<input onfocus=alert('xss') autofocus>"},
	{"select onfocus autofocus", "<select onfocus=alert('xss') autofocus>"},
	{"javascript url lowercase", "javascript:alert('xss')"},
	{"javascript url mixed case", "jAvAsCrIpT:alert('xss')"},
	{"javascript url html entities", "&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;alert('xss')"},
	{"data url text/html", "data:text/html,<script>alert('xss')</script>"},
	{"data url base64", "data:text/html;base64,PHNjcmlwdD5hbGVydCgneHNzJyk8L3NjcmlwdD4="},
	{"svg with script", "<svg><script>alert('xss')</script></svg>"},
	{"svg animate onbegin", "<svg><animate onbegin=alert('xss') attributeName=x dur=1s>"},
	{"img src decimal entities", "<IMG SRC=&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;alert('xss')>"},
	{"img src hex entities", "<IMG SRC=&#x6A;&#x61;&#x76;&#x61;&#x73;&#x63;&#x72;&#x69;&#x70;&#x74;&#x3A;alert('xss')>"},
	{"mutation xss noscript", "<noscript><p title=\"</noscript><img src=x onerror=alert('xss')\">"},
	{"iframe javascript", "<iframe src='javascript:alert(\"xss\")'></iframe>"},
	{"onclick attribute", "<div onclick='alert(\"xss\")'>Click</div>"},
	{"onmouseover", "<div onmouseover='alert(\"xss\")'>Hover</div>"},
	{"onerror in attribute", "\" onerror=\"alert('xss')"},
	{"style with expression", "<style>body{background:url('javascript:alert(1)')}</style>"},
	{"link href javascript", "<link rel=stylesheet href='javascript:alert(1)'>"},
	{"meta refresh javascript", "<meta http-equiv=refresh content='0;url=javascript:alert(1)'>"},
	{"object data", "<object data='javascript:alert(1)'>"},
	{"embed src", "<embed src='javascript:alert(1)'>"},
	{"form action javascript", "<form action='javascript:alert(1)'><input type=submit></form>"},
	{"button formaction", "<button formaction='javascript:alert(1)'>Click</button>"},
	{"input formaction", "<input type=submit formaction='javascript:alert(1)'>"},
	{"video onerror", "<video onerror=alert('xss')><source></video>"},
	{"audio onerror", "<audio onerror=alert('xss')><source></audio>"},
	{"details ontoggle", "<details ontoggle=alert('xss') open>"},
	{"marquee onstart", "<marquee onstart=alert('xss')>"},
	{"isindex action", "<isindex action='javascript:alert(1)'>"},
	{"table background", "<table background='javascript:alert(1)'>"},
	{"bgsound src", "<bgsound src='javascript:alert(1)'>"},
	{"base href", "<base href='javascript:alert(1)'>"},
	{"applet code", "<applet code='javascript:alert(1)'>"},
	{"xml with script", "<?xml version='1.0'?><script>alert('xss')</script>"},
	{"import with script", "<import implementation='#default#time2' onbegin='alert(1)'>"},
}

func TestXSSPrevention_OWASPVectors_HTMLContext(t *testing.T) {
	engine := NewEngine()
	tmplStr := "<div>{{.Payload}}</div>"

	for _, vector := range owaspXSSVectors {
		t.Run(vector.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			data := struct{ Payload string }{Payload: vector.payload}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if strings.Contains(vector.payload, "<") && !strings.Contains(result, "&lt;") {
				t.Errorf("Expected HTML escaping for %s\nPayload: %s\nOutput: %s",
					vector.name, vector.payload, result)
			}
		})
	}
}

func TestXSSPrevention_OWASPVectors_AttributeContext(t *testing.T) {
	engine := NewEngine()
	tmplStr := "<img alt='{{.Payload}}'>"

	for _, vector := range owaspXSSVectors {
		t.Run(vector.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			data := struct{ Payload string }{Payload: vector.payload}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if strings.Contains(vector.payload, "<") && !strings.Contains(result, "&lt;") {
				t.Errorf("Expected HTML escaping in attribute for %s\nPayload: %s\nOutput: %s",
					vector.name, vector.payload, result)
			}
		})
	}
}

func TestXSSPrevention_OWASPVectors_URLContext(t *testing.T) {
	engine := NewEngine()
	tmplStr := "<a href='{{.Payload}}'>Link</a>"

	javascriptVectors := []struct {
		name    string
		payload string
	}{
		{"javascript url lowercase", "javascript:alert('xss')"},
		{"javascript url mixed case", "jAvAsCrIpT:alert('xss')"},
		{"data url text/html", "data:text/html,<script>alert('xss')</script>"},
		{"data url base64", "data:text/html;base64,PHNjcmlwdD5hbGVydCgneHNzJyk8L3NjcmlwdD4="},
	}

	for _, vector := range javascriptVectors {
		t.Run(vector.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			data := struct{ Payload string }{Payload: vector.payload}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if !strings.Contains(result, "#ZgotmplZ") && !strings.Contains(result, "%28") {
				t.Errorf("Expected URL sanitization for %s\nPayload: %s\nOutput: %s",
					vector.name, vector.payload, result)
			}
		})
	}
}

func TestXSSPrevention_OWASPVectors_JavaScriptContext(t *testing.T) {
	engine := NewEngine()
	tmplStr := "<script>var data = {{.Payload}};</script>"

	jsVectors := []struct {
		name    string
		payload string
	}{
		{"quote escape attempt", "'; alert('xss'); //"},
		{"double quote escape", "\"; alert('xss'); //"},
		{"newline escape", "\\n'; alert('xss'); //"},
		{"unicode escape", "\\u0027; alert('xss'); //"},
	}

	for _, vector := range jsVectors {
		t.Run(vector.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			data := struct{ Payload string }{Payload: vector.payload}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if strings.Contains(result, "'; alert") && !strings.Contains(result, "\"'; alert") && !strings.Contains(result, "\\\\n'; alert") {
				t.Errorf("JavaScript context allows code injection for %s\nPayload: %s\nOutput: %s",
					vector.name, vector.payload, result)
			}
		})
	}
}

func TestXSSPrevention_ContextAwareEscaping(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		data     interface{}
		validate func(t *testing.T, result string)
	}{
		{
			name:     "HTML context escapes tags",
			template: "<div>{{.Content}}</div>",
			data:     struct{ Content string }{Content: "<script>alert('xss')</script>"},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "&lt;script&gt;") {
					t.Errorf("HTML context did not escape script tags: %s", result)
				}
			},
		},
		{
			name:     "Attribute context escapes quotes",
			template: "<img alt='{{.Alt}}'>",
			data:     struct{ Alt string }{Alt: "\" onerror=\"alert('xss')"},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "&#34;") {
					t.Errorf("Attribute context did not escape quotes: %s", result)
				}
			},
		},
		{
			name:     "URL context sanitizes javascript",
			template: "<a href='{{.URL}}'>Link</a>",
			data:     struct{ URL string }{URL: "javascript:alert('xss')"},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "#ZgotmplZ") {
					t.Errorf("URL context did not sanitize javascript: %s", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.ExecuteToString(tmpl, tt.data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestXSSPrevention_PolyglotPayload(t *testing.T) {
	engine := NewEngine()

	polyglot := "jaVasCript:/*-/*`/*\\`/*'/*\"/**/(/* */oNcliCk=alert() )//%0D%0A%0d%0a//</stYle/</titLe/</teXtarEa/</scRipt/--!>\\x3csVg/<sVg/oNloAd=alert()//\\x3e"

	contexts := []struct {
		name     string
		template string
	}{
		{"HTML", "<div>{{.Payload}}</div>"},
		{"Attribute", "<img alt='{{.Payload}}'>"},
		{"URL", "<a href='{{.Payload}}'>Link</a>"},
	}

	for _, ctx := range contexts {
		t.Run(ctx.name, func(t *testing.T) {
			tmpl, err := engine.Parse(ctx.template)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			data := struct{ Payload string }{Payload: polyglot}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if ctx.name == "URL" {
				if !strings.Contains(result, "#ZgotmplZ") {
					t.Errorf("Polyglot not sanitized in URL context\nOutput: %s", result)
				}
			} else {
				if !strings.Contains(result, "&lt;") {
					t.Errorf("Polyglot not escaped in %s context\nOutput: %s", ctx.name, result)
				}
			}
		})
	}
}

func TestXSSPrevention_EncodingBypass(t *testing.T) {
	engine := NewEngine()

	encodingVectors := []struct {
		name    string
		payload string
	}{
		{"HTML decimal entities", "&#60;&#115;&#99;&#114;&#105;&#112;&#116;&#62;alert('xss')&#60;&#47;&#115;&#99;&#114;&#105;&#112;&#116;&#62;"},
		{"HTML hex entities", "&#x3c;&#x73;&#x63;&#x72;&#x69;&#x70;&#x74;&#x3e;alert('xss')&#x3c;&#x2f;&#x73;&#x63;&#x72;&#x69;&#x70;&#x74;&#x3e;"},
		{"Mixed case tags", "<ScRiPt>alert('xss')</sCrIpT>"},
		{"Null byte injection", "<script\x00>alert('xss')</script>"},
		{"Tab in tag", "<script\t>alert('xss')</script>"},
		{"Newline in tag", "<script\n>alert('xss')</script>"},
	}

	tmplStr := "<div>{{.Payload}}</div>"

	for _, vector := range encodingVectors {
		t.Run(vector.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			data := struct{ Payload string }{Payload: vector.payload}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if strings.Contains(vector.payload, "<") && !strings.Contains(result, "&lt;") {
				t.Errorf("Encoding bypass not prevented for %s\nPayload: %s\nOutput: %s",
					vector.name, vector.payload, result)
			}
		})
	}
}

func TestXSSPrevention_MutationXSS(t *testing.T) {
	engine := NewEngine()

	mutationVectors := []struct {
		name    string
		payload string
	}{
		{"noscript mutation", "<noscript><p title=\"</noscript><img src=x onerror=alert('xss')\">"},
		{"style mutation", "<style><style/><img src=x onerror=alert('xss')>"},
		{"title mutation", "<title><title/><img src=x onerror=alert('xss')>"},
		{"textarea mutation", "<textarea><textarea/><img src=x onerror=alert('xss')>"},
	}

	tmplStr := "<div>{{.Payload}}</div>"

	for _, vector := range mutationVectors {
		t.Run(vector.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tmplStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			data := struct{ Payload string }{Payload: vector.payload}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if !strings.Contains(result, "&lt;") {
				t.Errorf("Mutation XSS not prevented for %s\nPayload: %s\nOutput: %s",
					vector.name, vector.payload, result)
			}
		})
	}
}

func TestXSSPrevention_VerifyEscaping(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name        string
		template    string
		payload     string
		mustContain string
	}{
		{
			name:        "script tags escaped",
			template:    "<div>{{.Input}}</div>",
			payload:     "<script>alert('xss')</script>",
			mustContain: "&lt;script&gt;",
		},
		{
			name:        "img tags escaped",
			template:    "<div>{{.Input}}</div>",
			payload:     "<img src=x onerror=alert('xss')>",
			mustContain: "&lt;img",
		},
		{
			name:        "svg tags escaped",
			template:    "<div>{{.Input}}</div>",
			payload:     "<svg onload=alert('xss')>",
			mustContain: "&lt;svg",
		},
		{
			name:        "quotes escaped in attributes",
			template:    "<img alt='{{.Input}}'>",
			payload:     "\" onerror=\"alert('xss')\"",
			mustContain: "&#34;",
		},
		{
			name:        "javascript URLs sanitized",
			template:    "<a href='{{.Input}}'>Link</a>",
			payload:     "javascript:alert('xss')",
			mustContain: "#ZgotmplZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			data := struct{ Input string }{Input: tt.payload}
			result, err := engine.ExecuteToString(tmpl, data)
			if err != nil {
				t.Fatalf("ExecuteToString() error = %v", err)
			}

			if !strings.Contains(result, tt.mustContain) {
				t.Errorf("Expected %q in output\nPayload: %s\nOutput: %s",
					tt.mustContain, tt.payload, result)
			}
		})
	}
}
