package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

type SecurityHeadersConfig struct {
	HSTSMaxAge            *int
	HSTSIncludeSubDomains bool
	HSTSPreload           bool
	
	CSPDefaultSrc     []string
	CSPScriptSrc      []string
	CSPStyleSrc       []string
	CSPImgSrc         []string
	CSPFontSrc        []string
	CSPConnectSrc     []string
	CSPFrameAncestors []string
	CSPBaseURI        []string
	CSPFormAction     []string
	CSPReportURI      string
	CSPReportOnly     bool
	
	XFrameOptions       string
	XContentTypeOptions string
	XXSSProtection      string
	ReferrerPolicy      string
	PermissionsPolicy   string
}

func SecurityHeaders(config *SecurityHeadersConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = &SecurityHeadersConfig{}
	}
	
	hstsHeader := buildHSTS(config)
	cspHeader := buildCSP(config)
	
	xFrameOptions := config.XFrameOptions
	if xFrameOptions == "" {
		xFrameOptions = "DENY"
	}
	
	xContentTypeOptions := config.XContentTypeOptions
	if xContentTypeOptions == "" {
		xContentTypeOptions = "nosniff"
	}
	
	xXSSProtection := config.XXSSProtection
	if xXSSProtection == "" {
		xXSSProtection = "1; mode=block"
	}
	
	referrerPolicy := config.ReferrerPolicy
	if referrerPolicy == "" {
		referrerPolicy = "strict-origin-when-cross-origin"
	}
	
	permissionsPolicy := config.PermissionsPolicy
	if permissionsPolicy == "" {
		permissionsPolicy = "geolocation=(), microphone=(), camera=()"
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hstsHeader != "" {
				w.Header().Set("Strict-Transport-Security", hstsHeader)
			}
			
			if config.CSPReportOnly {
				w.Header().Set("Content-Security-Policy-Report-Only", cspHeader)
			} else {
				w.Header().Set("Content-Security-Policy", cspHeader)
			}
			
			w.Header().Set("X-Content-Type-Options", xContentTypeOptions)
			w.Header().Set("X-Frame-Options", xFrameOptions)
			w.Header().Set("X-XSS-Protection", xXSSProtection)
			w.Header().Set("Referrer-Policy", referrerPolicy)
			w.Header().Set("Permissions-Policy", permissionsPolicy)
			
			next.ServeHTTP(w, r)
		})
	}
}

func buildHSTS(config *SecurityHeadersConfig) string {
	if config == nil {
		return "max-age=31536000; includeSubDomains"
	}
	
	var maxAge int
	includeSubDomains := config.HSTSIncludeSubDomains
	preload := config.HSTSPreload
	
	if config.HSTSMaxAge == nil {
		maxAge = 31536000
		includeSubDomains = true
	} else {
		maxAge = *config.HSTSMaxAge
		if maxAge == 0 {
			return ""
		}
	}
	
	parts := []string{"max-age=" + strconv.Itoa(maxAge)}
	
	if includeSubDomains {
		parts = append(parts, "includeSubDomains")
	}
	
	if preload {
		parts = append(parts, "preload")
	}
	
	return strings.Join(parts, "; ")
}

func buildCSP(config *SecurityHeadersConfig) string {
	directives := make([]string, 0, 9)
	
	var defaultSrc, scriptSrc, styleSrc, imgSrc, fontSrc, connectSrc, frameAncestors, baseURI, formAction []string
	var reportURI string
	
	if config != nil {
		defaultSrc = config.CSPDefaultSrc
		scriptSrc = config.CSPScriptSrc
		styleSrc = config.CSPStyleSrc
		imgSrc = config.CSPImgSrc
		fontSrc = config.CSPFontSrc
		connectSrc = config.CSPConnectSrc
		frameAncestors = config.CSPFrameAncestors
		baseURI = config.CSPBaseURI
		formAction = config.CSPFormAction
		reportURI = config.CSPReportURI
	}
	
	if len(defaultSrc) == 0 {
		defaultSrc = []string{"'self'"}
	}
	directives = append(directives, "default-src "+strings.Join(defaultSrc, " "))
	
	if len(scriptSrc) == 0 {
		scriptSrc = []string{"'self'", "'unsafe-inline'"}
	}
	directives = append(directives, "script-src "+strings.Join(scriptSrc, " "))
	
	if len(styleSrc) == 0 {
		styleSrc = []string{"'self'", "'unsafe-inline'"}
	}
	directives = append(directives, "style-src "+strings.Join(styleSrc, " "))
	
	if len(imgSrc) == 0 {
		imgSrc = []string{"'self'", "data:", "https:"}
	}
	directives = append(directives, "img-src "+strings.Join(imgSrc, " "))
	
	if len(fontSrc) == 0 {
		fontSrc = []string{"'self'"}
	}
	directives = append(directives, "font-src "+strings.Join(fontSrc, " "))
	
	if len(connectSrc) == 0 {
		connectSrc = []string{"'self'"}
	}
	directives = append(directives, "connect-src "+strings.Join(connectSrc, " "))
	
	if len(frameAncestors) == 0 {
		frameAncestors = []string{"'none'"}
	}
	directives = append(directives, "frame-ancestors "+strings.Join(frameAncestors, " "))
	
	if len(baseURI) == 0 {
		baseURI = []string{"'self'"}
	}
	directives = append(directives, "base-uri "+strings.Join(baseURI, " "))
	
	if len(formAction) == 0 {
		formAction = []string{"'self'"}
	}
	directives = append(directives, "form-action "+strings.Join(formAction, " "))
	
	if reportURI != "" {
		directives = append(directives, "report-uri "+reportURI)
	}
	
	return strings.Join(directives, "; ")
}
