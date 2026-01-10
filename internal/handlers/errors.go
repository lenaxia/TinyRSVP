package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Fields     map[string]string
	Err        error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    string            `json:"code,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func NewValidationError(fields map[string]string) *APIError {
	return &APIError{
		StatusCode: http.StatusBadRequest,
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		Fields:     fields,
	}
}

func NewNotFoundError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    message,
	}
}

func NewPermissionDeniedError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusForbidden,
		Code:       "PERMISSION_DENIED",
		Message:    message,
	}
}

func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

func NewConflictError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    message,
	}
}

func NewInternalError() *APIError {
	return &APIError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    "An internal error occurred",
	}
}

func NewBadRequestError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
	}
}

func NewTimeoutError() *APIError {
	return &APIError{
		StatusCode: http.StatusGatewayTimeout,
		Code:       "TIMEOUT",
		Message:    "Request timeout",
	}
}

func toAPIError(err error) *APIError {
	if errors.Is(err, context.DeadlineExceeded) {
		return &APIError{
			StatusCode: http.StatusGatewayTimeout,
			Code:       "TIMEOUT",
			Message:    "Request timeout",
			Err:        err,
		}
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	var notFoundErr *models.NotFoundError
	if errors.As(err, &notFoundErr) {
		return &APIError{
			StatusCode: http.StatusNotFound,
			Code:       "NOT_FOUND",
			Message:    notFoundErr.Error(),
			Err:        err,
		}
	}

	var validationErr *models.ValidationError
	if errors.As(err, &validationErr) {
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "VALIDATION_ERROR",
			Message:    validationErr.Error(),
			Err:        err,
		}
	}

	var conflictErr *models.ConflictError
	if errors.As(err, &conflictErr) {
		return &APIError{
			StatusCode: http.StatusConflict,
			Code:       "CONFLICT",
			Message:    conflictErr.Error(),
			Err:        err,
		}
	}

	var permissionErr *models.PermissionDeniedError
	if errors.As(err, &permissionErr) {
		return &APIError{
			StatusCode: http.StatusForbidden,
			Code:       "PERMISSION_DENIED",
			Message:    permissionErr.Error(),
			Err:        err,
		}
	}

	var unauthorizedErr *models.UnauthorizedError
	if errors.As(err, &unauthorizedErr) {
		return &APIError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    unauthorizedErr.Error(),
			Err:        err,
		}
	}

	var forbiddenErr *models.ForbiddenError
	if errors.As(err, &forbiddenErr) {
		return &APIError{
			StatusCode: http.StatusForbidden,
			Code:       "PERMISSION_DENIED",
			Message:    forbiddenErr.Error(),
			Err:        err,
		}
	}

	var assetsValidationErr *assets.ValidationError
	if errors.As(err, &assetsValidationErr) {
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "VALIDATION_ERROR",
			Message:    assetsValidationErr.Error(),
			Err:        err,
		}
	}

	var versionConflictErr *models.VersionConflictError
	if errors.As(err, &versionConflictErr) {
		return &APIError{
			StatusCode: http.StatusConflict,
			Code:       "VERSION_CONFLICT",
			Message:    "version conflict",
			Err:        err,
		}
	}

	var optimisticLockErr *models.OptimisticLockError
	if errors.As(err, &optimisticLockErr) {
		return &APIError{
			StatusCode: http.StatusConflict,
			Code:       "VERSION_CONFLICT",
			Message:    "version conflict",
			Err:        err,
		}
	}

	if err.Error() == "invalid state transition" {
		return &APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "BAD_REQUEST",
			Message:    "invalid state transition",
			Err:        err,
		}
	}

	return &APIError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    "An internal error occurred",
		Err:        err,
	}
}

func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}

func writeJSONError(w http.ResponseWriter, apiErr *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)

	resp := ErrorResponse{
		Error:   http.StatusText(apiErr.StatusCode),
		Message: apiErr.Message,
		Code:    apiErr.Code,
		Fields:  apiErr.Fields,
	}

	json.NewEncoder(w).Encode(resp)
}

func writeHTMLError(w http.ResponseWriter, r *http.Request, apiErr *APIError) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(apiErr.StatusCode)

	requestID := middleware.GetRequestID(r.Context())

	data := struct {
		StatusCode int
		Error      string
		Message    string
		RequestID  string
	}{
		StatusCode: apiErr.StatusCode,
		Error:      http.StatusText(apiErr.StatusCode),
		Message:    apiErr.Message,
		RequestID:  requestID,
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.StatusCode}} {{.Error}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            max-width: 600px;
            margin: 100px auto;
            padding: 20px;
            text-align: center;
            background-color: #f9fafb;
        }
        h1 {
            font-size: 72px;
            margin: 0;
            color: #dc3545;
        }
        h2 {
            font-size: 24px;
            margin: 20px 0;
            color: #333;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .request-id {
            margin-top: 40px;
            font-size: 12px;
            color: #999;
        }
        a {
            color: #007bff;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <h1>{{.StatusCode}}</h1>
    <h2>{{.Error}}</h2>
    <p>{{.Message}}</p>
    <p><a href="/">Return to Home</a></p>
    {{if .RequestID}}
    <div class="request-id">Request ID: {{.RequestID}}</div>
    {{end}}
</body>
</html>`

	t := template.Must(template.New("error").Parse(tmpl))
	t.Execute(w, data)
}

func logError(ctx context.Context, apiErr *APIError) {
	requestID := middleware.GetRequestID(ctx)

	log.Printf("[%s] Error %d: %s (code=%s)",
		requestID,
		apiErr.StatusCode,
		apiErr.Message,
		apiErr.Code,
	)

	if apiErr.StatusCode >= 500 && apiErr.Err != nil {
		log.Printf("[%s] Underlying error: %v", requestID, apiErr.Err)
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := toAPIError(err)

	logError(r.Context(), apiErr)

	if wantsJSON(r) {
		writeJSONError(w, apiErr)
	} else {
		writeHTMLError(w, r, apiErr)
	}
}

