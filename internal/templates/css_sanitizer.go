package templates

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type CSSSanitizer interface {
	Sanitize(css string) (string, error)
	Validate(css string) error
}

type cssSanitizer struct {
	dangerousPatterns []*regexp.Regexp
}

func NewCSSSanitizer() CSSSanitizer {
	return &cssSanitizer{
		dangerousPatterns: compileDangerousPatterns(),
	}
}

func compileDangerousPatterns() []*regexp.Regexp {
	patterns := []string{
		`javascript\s*:`,
		`expression\s*\(`,
		`behavior\s*:`,
		`@import`,
		`vbscript\s*:`,
		`data\s*:\s*text/html`,
		`-moz-binding`,
		`@charset`,
		`<script`,
		`</script`,
	}

	compiled := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		compiled[i] = regexp.MustCompile(`(?i)` + pattern)
	}

	return compiled
}

func (s *cssSanitizer) Validate(css string) error {
	for _, pattern := range s.dangerousPatterns {
		if pattern.MatchString(css) {
			patternStr := pattern.String()
			patternStr = strings.TrimPrefix(patternStr, "(?i)")
			patternStr = strings.ReplaceAll(patternStr, `\s*`, "")
			patternStr = strings.ReplaceAll(patternStr, `\(`, "(")
			patternStr = strings.ReplaceAll(patternStr, `\)`, ")")

			return &models.ValidationError{
				Field:   "css_content",
				Message: fmt.Sprintf("CSS contains dangerous pattern: %s", patternStr),
			}
		}
	}

	return nil
}

func (s *cssSanitizer) Sanitize(css string) (string, error) {
	if err := s.Validate(css); err != nil {
		return "", err
	}

	css = removeComments(css)
	css = normalizeWhitespace(css)

	return css, nil
}

func removeComments(css string) string {
	commentPattern := regexp.MustCompile(`(?s)/\*.*?\*/`)
	return commentPattern.ReplaceAllString(css, "")
}

func normalizeWhitespace(css string) string {
	css = regexp.MustCompile(`\s+`).ReplaceAllString(css, " ")
	css = strings.TrimSpace(css)
	return css
}
