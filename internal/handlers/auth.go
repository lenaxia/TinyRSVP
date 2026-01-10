package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

type Authenticator interface {
	HandleLogin(w http.ResponseWriter, r *http.Request) error
	HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error)
	HandleLogout(w http.ResponseWriter, r *http.Request) error
}

type AuthResult struct {
	Email       string
	Name        string
	OIDCSubject string
}

type AuthHandlers struct {
	authenticator Authenticator
}

func NewAuthHandlers(authenticator Authenticator) *AuthHandlers {
	return &AuthHandlers{
		authenticator: authenticator,
	}
}

func (h *AuthHandlers) ShowLogin(w http.ResponseWriter, r *http.Request) {
	returnURL := r.URL.Query().Get("return")
	
	validatedURL, err := validateReturnURL(returnURL)
	if err != nil {
		HandleError(w, r, NewBadRequestError("Invalid return URL"))
		return
	}

	data := struct {
		ReturnURL string
	}{
		ReturnURL: validatedURL,
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - TinyRSVP</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            max-width: 400px;
            margin: 100px auto;
            padding: 20px;
            text-align: center;
        }
        h1 {
            font-size: 32px;
            margin-bottom: 40px;
            color: #333;
        }
        .login-button {
            display: inline-block;
            padding: 12px 24px;
            background-color: #007bff;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            font-size: 16px;
        }
        .login-button:hover {
            background-color: #0056b3;
        }
    </style>
</head>
<body>
    <h1>TinyRSVP</h1>
    <p>Please log in to continue</p>
    <a href="/auth/oidc/login?return={{.ReturnURL}}" class="login-button">Log In with OIDC</a>
</body>
</html>`

	t := template.Must(template.New("login").Parse(tmpl))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	t.Execute(w, data)
}

func (h *AuthHandlers) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	returnURL := r.URL.Query().Get("return")
	
	if _, err := validateReturnURL(returnURL); err != nil {
		HandleError(w, r, NewBadRequestError("Invalid return URL"))
		return
	}

	if err := h.authenticator.HandleLogin(w, r); err != nil {
		HandleError(w, r, NewInternalError())
		return
	}
}

func (h *AuthHandlers) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	_, err := h.authenticator.HandleCallback(w, r)
	if err != nil {
		HandleError(w, r, NewUnauthorizedError("Authentication failed"))
		return
	}
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HandleError(w, r, &APIError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Method not allowed",
		})
		return
	}

	if err := h.authenticator.HandleLogout(w, r); err != nil {
		HandleError(w, r, NewInternalError())
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func validateReturnURL(returnURL string) (string, error) {
	if returnURL == "" {
		return "/", nil
	}

	parsedURL, err := url.Parse(returnURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Scheme != "" && parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", NewBadRequestError("Invalid URL scheme")
	}

	if parsedURL.Host != "" {
		return "", NewBadRequestError("External URLs not allowed")
	}

	if !strings.HasPrefix(returnURL, "/") {
		return "", NewBadRequestError("URL must be absolute path")
	}

	if strings.HasPrefix(returnURL, "//") {
		return "", NewBadRequestError("Protocol-relative URLs not allowed")
	}

	return returnURL, nil
}
