package auth

import (
	"fmt"
	"net/url"
	"strings"
)

func ValidateReturnURL(returnURL string) (string, error) {
	if returnURL == "" {
		return "/", nil
	}

	normalized := strings.ToLower(returnURL)

	if !strings.HasPrefix(returnURL, "/") {
		return "/", fmt.Errorf("return URL must start with /")
	}

	if strings.HasPrefix(returnURL, "//") {
		return "/", fmt.Errorf("protocol-relative URLs not allowed")
	}

	schemes := []string{"http:", "https:", "ftp:", "ftps:", "javascript:", "data:", "file:", "about:", "blob:", "mailto:"}
	for _, scheme := range schemes {
		if strings.Contains(normalized, scheme) {
			return "/", fmt.Errorf("scheme %s not allowed in return URL", scheme)
		}
	}

	if strings.ContainsAny(returnURL, "\t\n\r\x00\\") {
		return "/", fmt.Errorf("return URL contains invalid control characters")
	}

	if strings.Contains(strings.ToUpper(returnURL), "%0D") ||
		strings.Contains(strings.ToUpper(returnURL), "%0A") ||
		strings.Contains(strings.ToUpper(returnURL), "%00") {
		return "/", fmt.Errorf("return URL contains encoded control characters")
	}

	parsed, err := url.Parse(returnURL)
	if err != nil {
		return "/", fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "" {
		return "/", fmt.Errorf("absolute URLs not allowed")
	}

	if parsed.Host != "" {
		return "/", fmt.Errorf("external hosts not allowed")
	}

	if parsed.RawQuery != "" {
		normalizedQuery := strings.ToLower(parsed.RawQuery)
		for _, scheme := range schemes {
			if strings.Contains(normalizedQuery, scheme) {
				return "/", fmt.Errorf("scheme found in query parameter")
			}
		}
	}

	if parsed.Fragment != "" {
		normalizedFragment := strings.ToLower(parsed.Fragment)
		for _, scheme := range schemes {
			if strings.Contains(normalizedFragment, scheme) {
				return "/", fmt.Errorf("scheme found in fragment")
			}
		}
	}

	return returnURL, nil
}
