package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestSessionManager_CreateSession(t *testing.T) {
	mockRepo := &MockSessionRepository{
		CreateFunc: func(ctx context.Context, session *models.Session) error {
			if session.ID == "" {
				return fmt.Errorf("session ID is empty")
			}
			if session.UserID == 0 {
				return fmt.Errorf("user ID is 0")
			}
			return nil
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.1")
	r.Header.Set("User-Agent", "Mozilla/5.0")

	session, err := mgr.CreateSession(context.Background(), 123, r)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session == nil {
		t.Fatal("Expected session, got nil")
	}

	if session.ID == "" {
		t.Error("Expected non-empty session ID")
	}

	if session.UserID != 123 {
		t.Errorf("Expected user ID 123, got %d", session.UserID)
	}

	if session.IPAddress == nil || *session.IPAddress != "203.0.113.1" {
		t.Errorf("Expected IP address 203.0.113.1, got %v", session.IPAddress)
	}

	if session.UserAgent == nil || *session.UserAgent != "Mozilla/5.0" {
		t.Errorf("Expected user agent Mozilla/5.0, got %v", session.UserAgent)
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Error("Expected future expiration time")
	}

	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	if session.ExpiresAt.Sub(expectedExpiry) > time.Minute {
		t.Errorf("Expected expiry around %v, got %v", expectedExpiry, session.ExpiresAt)
	}
}

func TestSessionManager_CreateSession_UniqueIDs(t *testing.T) {
	mockRepo := &MockSessionRepository{
		CreateFunc: func(ctx context.Context, session *models.Session) error {
			return nil
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	r := httptest.NewRequest("GET", "/", nil)

	session1, err := mgr.CreateSession(context.Background(), 123, r)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	session2, err := mgr.CreateSession(context.Background(), 123, r)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session1.ID == session2.ID {
		t.Error("Expected different session IDs")
	}
}

func TestSessionManager_GetSession_Valid(t *testing.T) {
	sessionID := "test-session-id"
	mockRepo := &MockSessionRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*models.Session, error) {
			if id != sessionID {
				return nil, &models.NotFoundError{Resource: "Session", ID: id}
			}
			return &models.Session{
				ID:        sessionID,
				UserID:    123,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	session, err := mgr.GetSession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if session == nil {
		t.Fatal("Expected session, got nil")
	}

	if session.ID != sessionID {
		t.Errorf("Expected ID %s, got %s", sessionID, session.ID)
	}
}

func TestSessionManager_GetSession_Expired(t *testing.T) {
	sessionID := "expired-session"
	deleteSessionCalled := false

	mockRepo := &MockSessionRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*models.Session, error) {
			return &models.Session{
				ID:        sessionID,
				UserID:    123,
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id string) error {
			if id == sessionID {
				deleteSessionCalled = true
			}
			return nil
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	session, err := mgr.GetSession(context.Background(), sessionID)
	if err == nil {
		t.Fatal("Expected error for expired session, got nil")
	}

	if session != nil {
		t.Error("Expected nil session for expired")
	}

	if !deleteSessionCalled {
		t.Error("Expected expired session to be deleted")
	}
}

func TestSessionManager_RefreshSession(t *testing.T) {
	sessionID := "test-session-id"
	updateLastAccessedCalled := false

	mockRepo := &MockSessionRepository{
		UpdateLastAccessedFunc: func(ctx context.Context, id string) error {
			if id == sessionID {
				updateLastAccessedCalled = true
			}
			return nil
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	err := mgr.RefreshSession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("RefreshSession failed: %v", err)
	}

	if !updateLastAccessedCalled {
		t.Error("Expected UpdateLastAccessed to be called")
	}
}

func TestSessionManager_RefreshSession_NotFound(t *testing.T) {
	mockRepo := &MockSessionRepository{
		UpdateLastAccessedFunc: func(ctx context.Context, id string) error {
			return &models.NotFoundError{Resource: "Session", ID: id}
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	err := mgr.RefreshSession(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent session")
	}
}

func TestSessionManager_DeleteSession(t *testing.T) {
	sessionID := "test-session-id"
	deleteCalled := false

	mockRepo := &MockSessionRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			if id == sessionID {
				deleteCalled = true
			}
			return nil
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	err := mgr.DeleteSession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	if !deleteCalled {
		t.Error("Expected Delete to be called")
	}
}

func TestSessionManager_DeleteSession_NotFound(t *testing.T) {
	mockRepo := &MockSessionRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			return &models.NotFoundError{Resource: "Session", ID: id}
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	err := mgr.DeleteSession(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent session")
	}
}

func TestSessionManager_DeleteUserSessions(t *testing.T) {
	userID := int64(123)
	deleteCalled := false

	mockRepo := &MockSessionRepository{
		DeleteByUserIDFunc: func(ctx context.Context, uid int64) error {
			if uid == userID {
				deleteCalled = true
			}
			return nil
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	err := mgr.DeleteUserSessions(context.Background(), userID)
	if err != nil {
		t.Fatalf("DeleteUserSessions failed: %v", err)
	}

	if !deleteCalled {
		t.Error("Expected DeleteByUserID to be called")
	}
}

func TestSessionManager_CleanupExpired(t *testing.T) {
	mockRepo := &MockSessionRepository{
		DeleteExpiredFunc: func(ctx context.Context) (int64, error) {
			return 5, nil
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	count, err := mgr.CleanupExpired(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpired failed: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5 deleted sessions, got %d", count)
	}
}

func TestSessionManager_CleanupExpired_Error(t *testing.T) {
	mockRepo := &MockSessionRepository{
		DeleteExpiredFunc: func(ctx context.Context) (int64, error) {
			return 0, fmt.Errorf("database error")
		},
	}

	mgr := NewSessionManager(mockRepo, true)

	_, err := mgr.CleanupExpired(context.Background())
	if err == nil {
		t.Fatal("Expected error from CleanupExpired")
	}
}

func TestSessionManager_SetSessionCookie(t *testing.T) {
	mockRepo := &MockSessionRepository{}
	mgr := NewSessionManager(mockRepo, true)

	w := httptest.NewRecorder()
	sessionID := "test-session-id-123"

	err := mgr.SetSessionCookie(w, sessionID)
	if err != nil {
		t.Fatalf("SetSessionCookie failed: %v", err)
	}

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie, got none")
	}

	if sessionCookie.Value != sessionID {
		t.Errorf("Expected value %s, got %s", sessionID, sessionCookie.Value)
	}

	if !sessionCookie.HttpOnly {
		t.Error("Expected HttpOnly cookie")
	}

	if !sessionCookie.Secure {
		t.Error("Expected Secure cookie")
	}

	if sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Expected SameSite Lax, got %v", sessionCookie.SameSite)
	}

	if sessionCookie.Path != "/" {
		t.Errorf("Expected path /, got %s", sessionCookie.Path)
	}

	expectedMaxAge := 7 * 24 * 60 * 60
	if sessionCookie.MaxAge != expectedMaxAge {
		t.Errorf("Expected MaxAge %d, got %d", expectedMaxAge, sessionCookie.MaxAge)
	}
}

func TestSessionManager_SetSessionCookie_NonSecure(t *testing.T) {
	mockRepo := &MockSessionRepository{}
	mgr := NewSessionManager(mockRepo, false)

	w := httptest.NewRecorder()
	sessionID := "test-session-id"

	err := mgr.SetSessionCookie(w, sessionID)
	if err != nil {
		t.Fatalf("SetSessionCookie failed: %v", err)
	}

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie, got none")
	}

	if sessionCookie.Secure {
		t.Error("Expected non-secure cookie in non-secure mode")
	}

	if !sessionCookie.HttpOnly {
		t.Error("Expected HttpOnly cookie even in non-secure mode")
	}
}

func TestSessionManager_ClearSessionCookie(t *testing.T) {
	mockRepo := &MockSessionRepository{}
	mgr := NewSessionManager(mockRepo, true)

	w := httptest.NewRecorder()

	err := mgr.ClearSessionCookie(w)
	if err != nil {
		t.Fatalf("ClearSessionCookie failed: %v", err)
	}

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie, got none")
	}

	if sessionCookie.Value != "" {
		t.Errorf("Expected empty value, got %s", sessionCookie.Value)
	}

	if sessionCookie.MaxAge != -1 {
		t.Errorf("Expected MaxAge -1, got %d", sessionCookie.MaxAge)
	}
}

func TestSessionManager_GetSessionFromRequest(t *testing.T) {
	mockRepo := &MockSessionRepository{}
	mgr := NewSessionManager(mockRepo, true)

	sessionID := "test-session-id"
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: sessionID,
	})

	retrievedID, err := mgr.GetSessionFromRequest(r)
	if err != nil {
		t.Fatalf("GetSessionFromRequest failed: %v", err)
	}

	if retrievedID != sessionID {
		t.Errorf("Expected %s, got %s", sessionID, retrievedID)
	}
}

func TestSessionManager_GetSessionFromRequest_NoCookie(t *testing.T) {
	mockRepo := &MockSessionRepository{}
	mgr := NewSessionManager(mockRepo, true)

	r := httptest.NewRequest("GET", "/", nil)

	_, err := mgr.GetSessionFromRequest(r)
	if err == nil {
		t.Fatal("Expected error for missing cookie, got nil")
	}
}

func TestGenerateSessionID(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		id, err := generateSessionID()
		if err != nil {
			t.Fatalf("generateSessionID() error = %v", err)
		}

		if id == "" {
			t.Error("Generated empty session ID")
		}

		if len(id) < 40 {
			t.Errorf("Session ID too short: %d bytes", len(id))
		}

		if seen[id] {
			t.Errorf("Collision detected: %q", id)
		}
		seen[id] = true
	}
}

func TestGenerateSessionID_Length(t *testing.T) {
	id, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID() error = %v", err)
	}

	if len(id) != 44 {
		t.Errorf("Expected session ID length 44 (32 bytes base64url encoded), got %d", len(id))
	}
}

func TestGetClientIP_DirectConnection(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.100:12345"

	ip := getClientIP(r)

	if ip != "192.168.1.100" {
		t.Errorf("Expected IP 192.168.1.100, got %s", ip)
	}
}

func TestGetClientIP_XForwardedFor_Single(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "192.168.1.100")

	ip := getClientIP(r)

	if ip != "192.168.1.100" {
		t.Errorf("Expected IP 192.168.1.100, got %s", ip)
	}
}

func TestGetClientIP_XForwardedFor_Multiple(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.2, 10.0.0.1")

	ip := getClientIP(r)

	if ip != "192.168.1.100" {
		t.Errorf("Expected first IP 192.168.1.100, got %s", ip)
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Real-IP", "192.168.1.100")

	ip := getClientIP(r)

	if ip != "192.168.1.100" {
		t.Errorf("Expected IP 192.168.1.100, got %s", ip)
	}
}

func TestGetClientIP_XRealIP_Precedence(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Real-IP", "192.168.1.100")
	r.Header.Set("X-Forwarded-For", "192.168.1.200")

	ip := getClientIP(r)

	if ip != "192.168.1.100" {
		t.Errorf("Expected X-Real-IP to take precedence, got %s", ip)
	}
}

func TestGetClientIP_IPv6(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "[2001:db8::1]:12345"

	ip := getClientIP(r)

	if ip != "2001:db8::1" {
		t.Errorf("Expected IPv6 2001:db8::1, got %s", ip)
	}
}

func TestGetClientIP_IPv6_NoPort(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "[2001:db8::1]"

	ip := getClientIP(r)

	if ip != "2001:db8::1" {
		t.Errorf("Expected IPv6 2001:db8::1, got %s", ip)
	}
}

func TestGetClientIP_WithWhitespace(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "  192.168.1.100  , 10.0.0.2")

	ip := getClientIP(r)

	if ip != "192.168.1.100" {
		t.Errorf("Expected trimmed IP 192.168.1.100, got %s", ip)
	}
}

type MockSessionRepository struct {
	CreateFunc             func(ctx context.Context, session *models.Session) error
	GetByIDFunc            func(ctx context.Context, id string) (*models.Session, error)
	GetByUserIDFunc        func(ctx context.Context, userID int64) ([]*models.Session, error)
	UpdateFunc             func(ctx context.Context, session *models.Session) error
	DeleteFunc             func(ctx context.Context, id string) error
	DeleteByUserIDFunc     func(ctx context.Context, userID int64) error
	DeleteExpiredFunc      func(ctx context.Context) (int64, error)
	UpdateLastAccessedFunc func(ctx context.Context, id string) error
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, session)
	}
	return nil
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &models.Session{ID: id, UserID: 1}, nil
}

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Session, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return []*models.Session{}, nil
}

func (m *MockSessionRepository) Update(ctx context.Context, session *models.Session) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, session)
	}
	return nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	if m.DeleteByUserIDFunc != nil {
		return m.DeleteByUserIDFunc(ctx, userID)
	}
	return nil
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	if m.DeleteExpiredFunc != nil {
		return m.DeleteExpiredFunc(ctx)
	}
	return 0, nil
}

func (m *MockSessionRepository) UpdateLastAccessed(ctx context.Context, id string) error {
	if m.UpdateLastAccessedFunc != nil {
		return m.UpdateLastAccessedFunc(ctx, id)
	}
	return nil
}
