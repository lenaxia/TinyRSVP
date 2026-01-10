package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
)

const (
	CSRFCookieName = "csrf_token"
	CSRFHeaderName = "X-CSRF-Token"
	CSRFFieldName  = "csrf_token"
)

type csrfContextKey string

const csrfTokenKey csrfContextKey = "csrf_token"

var CSRFTokenKey = csrfTokenKey

func CSRF(tokenLength int) Middleware {
	if tokenLength <= 0 {
		panic("CSRF token length must be positive")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := getOrGenerateToken(r, tokenLength)
			if err != nil {
				http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), csrfTokenKey, token)
			r = r.WithContext(ctx)

			if !isSafeMethod(r.Method) {
				if !validateCSRFToken(r, token) {
					http.Error(w, "Invalid or missing CSRF token", http.StatusForbidden)
					return
				}

				newToken, err := generateToken(tokenLength)
				if err != nil {
					http.Error(w, "Failed to rotate CSRF token", http.StatusInternalServerError)
					return
				}

				ctx = context.WithValue(r.Context(), csrfTokenKey, newToken)
				r = r.WithContext(ctx)
				setCSRFCookie(w, newToken)
			} else {
				setCSRFCookie(w, token)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetCSRFToken(ctx context.Context) string {
	token, ok := ctx.Value(csrfTokenKey).(string)
	if !ok {
		return ""
	}
	return token
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet ||
		method == http.MethodHead ||
		method == http.MethodOptions ||
		method == http.MethodTrace
}

func getOrGenerateToken(r *http.Request, tokenLength int) (string, error) {
	cookie, err := r.Cookie(CSRFCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return generateToken(tokenLength)
}

func generateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func setCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		HttpOnly: false,
		Secure:   false,
	})
}

func validateCSRFToken(r *http.Request, expectedToken string) bool {
	submittedToken := getSubmittedToken(r)
	if submittedToken == "" {
		return false
	}

	cookieToken, err := r.Cookie(CSRFCookieName)
	if err != nil || cookieToken.Value == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(submittedToken), []byte(expectedToken)) == 1 &&
		subtle.ConstantTimeCompare([]byte(cookieToken.Value), []byte(expectedToken)) == 1
}

func getSubmittedToken(r *http.Request) string {
	token := r.Header.Get(CSRFHeaderName)
	if token != "" {
		return token
	}

	if err := r.ParseForm(); err != nil {
		return ""
	}

	return r.FormValue(CSRFFieldName)
}
