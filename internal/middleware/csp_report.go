package middleware

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type CSPReport struct {
	CSPReport CSPReportDetails `json:"csp-report"`
}

type CSPReportDetails struct {
	DocumentURI        string `json:"document-uri"`
	ViolatedDirective  string `json:"violated-directive"`
	BlockedURI         string `json:"blocked-uri"`
	SourceFile         string `json:"source-file,omitempty"`
	LineNumber         int    `json:"line-number,omitempty"`
	ColumnNumber       int    `json:"column-number,omitempty"`
	StatusCode         int    `json:"status-code,omitempty"`
	EffectiveDirective string `json:"effective-directive,omitempty"`
	OriginalPolicy     string `json:"original-policy,omitempty"`
	Disposition        string `json:"disposition,omitempty"`
	ScriptSample       string `json:"script-sample,omitempty"`
	Referrer           string `json:"referrer,omitempty"`
}

func CSPReportHandler(logger *log.Logger) http.Handler {
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/csp-report" && contentType != "application/json" {
			http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 10240))
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if len(body) == 0 {
			http.Error(w, "Empty request body", http.StatusBadRequest)
			return
		}

		if len(body) >= 10240 {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		var report CSPReport
		if err := json.Unmarshal(body, &report); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		logCSPViolation(logger, r, &report)

		w.WriteHeader(http.StatusNoContent)
	})
}

func logCSPViolation(logger *log.Logger, r *http.Request, report *CSPReport) {
	details := report.CSPReport

	var logParts []string
	logParts = append(logParts, "CSP violation reported")

	if details.DocumentURI != "" {
		logParts = append(logParts, "document-uri="+details.DocumentURI)
	}

	if details.ViolatedDirective != "" {
		logParts = append(logParts, "violated-directive="+details.ViolatedDirective)
	}

	if details.BlockedURI != "" {
		logParts = append(logParts, "blocked-uri="+details.BlockedURI)
	}

	if details.SourceFile != "" {
		logParts = append(logParts, "source-file="+details.SourceFile)
	}

	if details.LineNumber > 0 {
		logParts = append(logParts, "line-number="+string(rune(details.LineNumber+'0')))
	}

	if details.EffectiveDirective != "" {
		logParts = append(logParts, "effective-directive="+details.EffectiveDirective)
	}

	requestID := GetRequestID(r.Context())
	if requestID != "" {
		logParts = append(logParts, "request-id="+requestID)
	}

	realIP := GetRealIP(r.Context())
	if realIP != "" {
		logParts = append(logParts, "client-ip="+realIP)
	}

	logger.Println(strings.Join(logParts, " "))
}
