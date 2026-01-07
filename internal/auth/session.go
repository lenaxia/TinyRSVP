package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/tinyrsvp/internal/db/repositories"
	"github.com/yourusername/tinyrsvp/internal/models"
)

const (
	SessionCookieName = "tinyrsvp_session"
	SessionDuration   = 7 * 24 * time.Hour
)

type sessionManager struct {
	repo   repositories.SessionRepository
	secure bool
}

func NewSessionManager(repo repositories.SessionRepository, secure bool) SessionManager {
	return &sessionManager{
		repo:   repo,
		secure: secure,
	}
}

func (m *sessionManager) CreateSession(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	now := time.Now()
	session := &models.Session{
		ID:             sessionID,
		UserID:         userID,
		CreatedAt:      now,
		ExpiresAt:      now.Add(SessionDuration),
		LastAccessedAt: now,
		IPAddress:      &ipAddress,
		UserAgent:      &userAgent,
	}

	if err := m.repo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (m *sessionManager) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := m.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsExpired() {
		m.repo.Delete(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

func (m *sessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	return m.repo.UpdateLastAccessed(ctx, sessionID)
}

func (m *sessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	return m.repo.Delete(ctx, sessionID)
}

func (m *sessionManager) DeleteUserSessions(ctx context.Context, userID int64) error {
	return m.repo.DeleteByUserID(ctx, userID)
}

func (m *sessionManager) CleanupExpired(ctx context.Context) (int64, error) {
	return m.repo.DeleteExpired(ctx)
}

func (m *sessionManager) SetSessionCookie(w http.ResponseWriter, sessionID string) error {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(SessionDuration.Seconds()),
		Secure:   m.secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *sessionManager) ClearSessionCookie(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   m.secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *sessionManager) GetSessionFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", fmt.Errorf("session cookie not found: %w", err)
	}
	return cookie.Value, nil
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		if strings.Contains(host, "[") {
			if idx > strings.LastIndex(host, "]") {
				host = host[:idx]
			}
		} else {
			host = host[:idx]
		}
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return host
}
