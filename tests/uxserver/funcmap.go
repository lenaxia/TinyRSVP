package uxserver

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// BuildTemplateFuncMap returns the template function map used by all UX test
// servers. Mirrors the funcMap used in production (cmd/server/main.go) so
// templates parse identically in tests.
func BuildTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"until": func(count int) []int {
			result := make([]int, count)
			for i := range result {
				result[i] = i
			}
			return result
		},
		"iterate": func(count int) []int {
			result := make([]int, count)
			for i := range result {
				result[i] = i
			}
			return result
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"formatDateTime": func(t time.Time) string {
			return t.Format("Monday, January 2, 2006 at 3:04 PM MST")
		},
		"formatTime": func(t time.Time) string {
			return t.Format("3:04 PM MST")
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires even number of arguments")
			}
			d := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				d[key] = values[i+1]
			}
			return d, nil
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"timezoneAbbr": func(iana string) string {
			loc, err := time.LoadLocation(iana)
			if err != nil {
				return iana
			}
			return time.Now().In(loc).Format("MST")
		},
	}
}
