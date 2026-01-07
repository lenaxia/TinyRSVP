package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/tinyrsvp/internal/models"
)

type mockUserService struct {
	getOrCreateUserFunc func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
}

func (m *mockUserService) GetOrCreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	if m.getOrCreateUserFunc != nil {
		return m.getOrCreateUserFunc(ctx, email, name, oidcSubject)
	}
	return &models.User{ID: 1, Email: email, Name: name}, nil
}

func (m *mockUserService) CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserService) UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error {
	return nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id int64) error {
	return nil
}

func (m *mockUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return nil, nil
}

type mockSessionManager struct {
	createSessionFunc      func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error)
	setSessionCookieFunc   func(w http.ResponseWriter, sessionID string) error
	clearSessionCookieFunc func(w http.ResponseWriter) error
	getSessionFromRequest  func(r *http.Request) (string, error)
	deleteSessionFunc      func(ctx context.Context, sessionID string) error
}

func (m *mockSessionManager) CreateSession(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, userID, r)
	}
	return &models.Session{ID: "session123", UserID: userID}, nil
}

func (m *mockSessionManager) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	return nil, nil
}

func (m *mockSessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *mockSessionManager) DeleteUserSessions(ctx context.Context, userID int64) error {
	return nil
}

func (m *mockSessionManager) CleanupExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockSessionManager) SetSessionCookie(w http.ResponseWriter, sessionID string) error {
	if m.setSessionCookieFunc != nil {
		return m.setSessionCookieFunc(w, sessionID)
	}
	return nil
}

func (m *mockSessionManager) ClearSessionCookie(w http.ResponseWriter) error {
	if m.clearSessionCookieFunc != nil {
		return m.clearSessionCookieFunc(w)
	}
	return nil
}

func (m *mockSessionManager) GetSessionFromRequest(r *http.Request) (string, error) {
	if m.getSessionFromRequest != nil {
		return m.getSessionFromRequest(r)
	}
	return "", http.ErrNoCookie
}

func TestForwardAuthenticator_HandleCallback(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		trustedIPs []string
		wantErr    bool
		wantEmail  string
		wantName   string
	}{
		{
			name: "valid authelia headers",
			headers: map[string]string{
				"Remote-User":  "testuser",
				"Remote-Email": "test@example.com",
				"Remote-Name":  "Test User",
			},
			remoteAddr: "127.0.0.1:12345",
			trustedIPs: []string{"127.0.0.1"},
			wantErr:    false,
			wantEmail:  "test@example.com",
			wantName:   "Test User",
		},
		{
			name: "missing email header",
			headers: map[string]string{
				"Remote-User": "testuser",
			},
			remoteAddr: "127.0.0.1:12345",
			trustedIPs: []string{"127.0.0.1"},
			wantErr:    true,
		},
		{
			name: "missing user header",
			headers: map[string]string{
				"Remote-Email": "test@example.com",
			},
			remoteAddr: "127.0.0.1:12345",
			trustedIPs: []string{"127.0.0.1"},
			wantErr:    true,
		},
		{
			name: "invalid email format",
			headers: map[string]string{
				"Remote-User":  "testuser",
				"Remote-Email": "not-an-email",
			},
			remoteAddr: "127.0.0.1:12345",
			trustedIPs: []string{"127.0.0.1"},
			wantErr:    true,
		},
		{
			name: "empty email",
			headers: map[string]string{
				"Remote-User":  "testuser",
				"Remote-Email": "",
			},
			remoteAddr: "127.0.0.1:12345",
			trustedIPs: []string{"127.0.0.1"},
			wantErr:    true,
		},
		{
			name: "empty user",
			headers: map[string]string{
				"Remote-User":  "",
				"Remote-Email": "test@example.com",
			},
			remoteAddr: "127.0.0.1:12345",
			trustedIPs: []string{"127.0.0.1"},
			wantErr:    true,
		},
		{
			name: "untrusted IP",
			headers: map[string]string{
				"Remote-User":  "testuser",
				"Remote-Email": "test@example.com",
			},
			remoteAddr: "192.168.1.100:12345",
			trustedIPs: []string{"127.0.0.1"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ForwardAuthConfig{
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  tt.trustedIPs,
			}

			userService := &mockUserService{}
			sessionMgr := &mockSessionManager{}

			auth := NewForwardAuthenticator(cfg, userService, sessionMgr)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr

			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}

			result, err := auth.HandleCallback(w, r)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Email != tt.wantEmail {
					t.Errorf("Expected email %q, got %q", tt.wantEmail, result.Email)
				}

				if result.Name != tt.wantName {
					t.Errorf("Expected name %q, got %q", tt.wantName, result.Name)
				}

				if result.OIDCSubject != nil {
					t.Error("Expected nil OIDC subject for forward auth")
				}
			}
		})
	}
}

func TestForwardAuthenticator_IPValidation(t *testing.T) {
	tests := []struct {
		name       string
		trustedIPs []string
		remoteAddr string
		headers    map[string]string
		wantErr    bool
	}{
		{
			name:       "trusted direct connection",
			trustedIPs: []string{"127.0.0.1"},
			remoteAddr: "127.0.0.1:12345",
			wantErr:    false,
		},
		{
			name:       "untrusted direct connection",
			trustedIPs: []string{"127.0.0.1"},
			remoteAddr: "192.168.1.100:12345",
			wantErr:    true,
		},
		{
			name:       "trusted via X-Forwarded-For",
			trustedIPs: []string{"10.0.0.1"},
			remoteAddr: "192.168.1.100:12345",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1, 192.168.1.100",
			},
			wantErr: false,
		},
		{
			name:       "trusted via X-Real-IP",
			trustedIPs: []string{"10.0.0.1"},
			remoteAddr: "192.168.1.100:12345",
			headers: map[string]string{
				"X-Real-IP": "10.0.0.1",
			},
			wantErr: false,
		},
		{
			name:       "untrusted via X-Forwarded-For",
			trustedIPs: []string{"10.0.0.1"},
			remoteAddr: "192.168.1.100:12345",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.50, 192.168.1.100",
			},
			wantErr: true,
		},
		{
			name:       "IPv6 trusted",
			trustedIPs: []string{"::1"},
			remoteAddr: "[::1]:12345",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ForwardAuthConfig{
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  tt.trustedIPs,
			}

			userService := &mockUserService{}
			sessionMgr := &mockSessionManager{}

			auth := NewForwardAuthenticator(cfg, userService, sessionMgr)

			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			r.Header.Set("Remote-User", "testuser")
			r.Header.Set("Remote-Email", "test@example.com")

			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			_, err := auth.HandleCallback(w, r)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCallback() with IP validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestForwardAuthenticator_HandleLogin(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		wantErr    bool
	}{
		{
			name: "valid headers creates session",
			headers: map[string]string{
				"Remote-User":  "testuser",
				"Remote-Email": "test@example.com",
			},
			remoteAddr: "127.0.0.1:12345",
			wantErr:    false,
		},
		{
			name: "invalid headers returns error",
			headers: map[string]string{
				"Remote-User": "testuser",
			},
			remoteAddr: "127.0.0.1:12345",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ForwardAuthConfig{
				UserHeader:  "Remote-User",
				EmailHeader: "Remote-Email",
				TrustedIPs:  []string{"127.0.0.1"},
			}

			userService := &mockUserService{}
			sessionMgr := &mockSessionManager{}

			auth := NewForwardAuthenticator(cfg, userService, sessionMgr)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr

			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}

			err := auth.HandleLogin(w, r)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestForwardAuthenticator_HandleLogout(t *testing.T) {
	cfg := &ForwardAuthConfig{
		UserHeader:  "Remote-User",
		EmailHeader: "Remote-Email",
		TrustedIPs:  []string{"127.0.0.1"},
	}

	sessionDeleted := false
	cookieCleared := false

	userService := &mockUserService{}
	sessionMgr := &mockSessionManager{
		getSessionFromRequest: func(r *http.Request) (string, error) {
			return "session123", nil
		},
		deleteSessionFunc: func(ctx context.Context, sessionID string) error {
			sessionDeleted = true
			return nil
		},
		clearSessionCookieFunc: func(w http.ResponseWriter) error {
			cookieCleared = true
			return nil
		},
	}

	auth := NewForwardAuthenticator(cfg, userService, sessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/logout", nil)

	err := auth.HandleLogout(w, r)
	if err != nil {
		t.Errorf("HandleLogout() error = %v", err)
	}

	if !sessionDeleted {
		t.Error("Expected session to be deleted")
	}

	if !cookieCleared {
		t.Error("Expected session cookie to be cleared")
	}
}
