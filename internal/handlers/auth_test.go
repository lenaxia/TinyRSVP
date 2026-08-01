package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestShowLogin_ValidReturnURL(t *testing.T) {
	tests := []struct {
		name           string
		returnURL      string
		wantStatusCode int
		wantInBody     string
	}{
		{
			name:           "with valid return URL",
			returnURL:      "/dashboard",
			wantStatusCode: http.StatusOK,
			wantInBody:     "return=%2fdashboard",
		},
		{
			name:           "with root return URL",
			returnURL:      "/",
			wantStatusCode: http.StatusOK,
			wantInBody:     "return=%2f",
		},
		{
			name:           "with empty return URL defaults to root",
			returnURL:      "",
			wantStatusCode: http.StatusOK,
			wantInBody:     "return=%2f",
		},
		{
			name:           "with nested path",
			returnURL:      "/events/123",
			wantStatusCode: http.StatusOK,
			wantInBody:     "return=%2fevents%2f123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AuthHandlers{}

			reqURL := "/login"
			if tt.returnURL != "" {
				reqURL += "?return=" + url.QueryEscape(tt.returnURL)
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			h.ShowLogin(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("ShowLogin() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.wantInBody) {
				t.Errorf("ShowLogin() body should contain %q, got %q", tt.wantInBody, body)
			}
		})
	}
}

func TestShowLogin_InvalidReturnURL(t *testing.T) {
	tests := []struct {
		name           string
		returnURL      string
		wantStatusCode int
	}{
		{
			name:           "external URL rejected",
			returnURL:      "https://evil.com/phishing",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "protocol relative URL rejected",
			returnURL:      "//evil.com/phishing",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "javascript protocol rejected",
			returnURL:      "javascript:alert(1)",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "data URL rejected",
			returnURL:      "data:text/html,<script>alert(1)</script>",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AuthHandlers{}

			reqURL := "/login?return=" + url.QueryEscape(tt.returnURL)
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			h.ShowLogin(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("ShowLogin() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestOIDCLogin_RedirectsToProvider(t *testing.T) {
	tests := []struct {
		name           string
		returnURL      string
		wantStatusCode int
	}{
		{
			name:           "redirects with return URL",
			returnURL:      "/dashboard",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "redirects without return URL",
			returnURL:      "",
			wantStatusCode: http.StatusFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &mockAuthenticator{
				handleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
					http.Redirect(w, r, "https://oidc.example.com/authorize", http.StatusFound)
					return nil
				},
			}

			h := &AuthHandlers{
				authenticator: mockAuth,
			}

			reqURL := "/auth/oidc/login"
			if tt.returnURL != "" {
				reqURL += "?return=" + url.QueryEscape(tt.returnURL)
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			h.OIDCLogin(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("OIDCLogin() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			location := w.Header().Get("Location")
			if location == "" {
				t.Error("OIDCLogin() should set Location header")
			}
		})
	}
}

func TestOIDCLogin_InvalidReturnURL(t *testing.T) {
	h := &AuthHandlers{
		authenticator: &mockAuthenticator{},
	}

	reqURL := "/auth/oidc/login?return=" + url.QueryEscape("https://evil.com")
	req := httptest.NewRequest(http.MethodGet, reqURL, nil)
	w := httptest.NewRecorder()

	h.OIDCLogin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("OIDCLogin() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestOIDCLogin_AuthenticatorError(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			return http.ErrAbortHandler
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/login", nil)
	w := httptest.NewRecorder()

	h.OIDCLogin(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("OIDCLogin() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestOIDCCallback_Success(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{
				Email:       "user@example.com",
				Name:        "Test User",
				OIDCSubject: strPtr("subject-123"),
			}, nil
		},
	}

	userService := &mockAuthUserService{
		getOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return &models.User{ID: 7, Email: email, Role: models.RoleEventManager}, nil
		},
		updateLastLoginFunc: func(ctx context.Context, userID int64) error {
			return nil
		},
	}

	sessionMgr := &mockAuthSessionManager{
		createSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
			return &models.Session{ID: "sess-1"}, nil
		},
		setSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
			return nil
		},
		clearSessionCookieFunc: func(w http.ResponseWriter) error {
			return nil
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
		userService:   userService,
		sessionMgr:    sessionMgr,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc123&state=xyz", nil)
	w := httptest.NewRecorder()

	h.OIDCCallback(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("OIDCCallback() status = %v, want %v", w.Code, http.StatusFound)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("OIDCCallback() location = %v, want / (default when no return URL)", location)
	}
}

func TestOIDCCallback_Error(t *testing.T) {
	tests := []struct {
		name           string
		callbackError  error
		wantStatusCode int
	}{
		{
			name:           "authentication error",
			callbackError:  http.ErrAbortHandler,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &mockAuthenticator{
				handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
					return nil, tt.callbackError
				},
			}

			h := &AuthHandlers{
				authenticator: mockAuth,
				userService:   &mockAuthUserService{},
				sessionMgr:    &mockAuthSessionManager{},
			}

			req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc123&state=xyz", nil)
			w := httptest.NewRecorder()

			h.OIDCCallback(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("OIDCCallback() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestLogout_Success(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
		sessionMgr:    &mockAuthSessionManager{},
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Logout() status = %v, want %v", w.Code, http.StatusFound)
	}

	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Logout() location = %v, want /login", location)
	}
}

func TestLogout_MethodNotAllowed(t *testing.T) {
	h := &AuthHandlers{
		authenticator: &mockAuthenticator{},
	}

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Logout() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestLogout_Error(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return http.ErrAbortHandler
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Logout() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestValidateReturnURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid root path",
			url:     "/",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			url:     "/dashboard",
			wantErr: false,
		},
		{
			name:    "valid nested path",
			url:     "/events/123/edit",
			wantErr: false,
		},
		{
			name:    "valid path with query",
			url:     "/dashboard?tab=events",
			wantErr: false,
		},
		{
			name:    "empty string defaults to root",
			url:     "",
			wantErr: false,
		},
		{
			name:    "external URL rejected",
			url:     "https://evil.com",
			wantErr: true,
		},
		{
			name:    "protocol relative URL rejected",
			url:     "//evil.com",
			wantErr: true,
		},
		{
			name:    "javascript protocol rejected",
			url:     "javascript:alert(1)",
			wantErr: true,
		},
		{
			name:    "data URL rejected",
			url:     "data:text/html,<script>",
			wantErr: true,
		},
		{
			name:    "relative path without leading slash rejected",
			url:     "dashboard",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := auth.ValidateReturnURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("auth.ValidateReturnURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockAuthenticator struct {
	handleLoginFunc    func(w http.ResponseWriter, r *http.Request) error
	handleCallbackFunc func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error)
	handleLogoutFunc   func(w http.ResponseWriter, r *http.Request) error
}

func (m *mockAuthenticator) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	if m.handleLoginFunc != nil {
		return m.handleLoginFunc(w, r)
	}
	return nil
}

func (m *mockAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
	if m.handleCallbackFunc != nil {
		return m.handleCallbackFunc(w, r)
	}
	return nil, nil
}

func (m *mockAuthenticator) HandleLogout(w http.ResponseWriter, r *http.Request) error {
	if m.handleLogoutFunc != nil {
		return m.handleLogoutFunc(w, r)
	}
	return nil
}

type mockAuthUserService struct {
	createUserFunc        func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
	getOrCreateUserFunc   func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
	getUserByIDFunc       func(ctx context.Context, id int64) (*models.User, error)
	getUserByEmailFunc    func(ctx context.Context, email string) (*models.User, error)
	updateUserFunc        func(ctx context.Context, user *models.User) error
	updateUserRoleFunc    func(ctx context.Context, userID int64, role models.UserRole) error
	updateLastLoginFunc   func(ctx context.Context, userID int64) error
	deleteUserFunc        func(ctx context.Context, id int64) error
	listUsersFunc         func(ctx context.Context, limit, offset int) ([]*models.User, error)
	countUsersFunc        func(ctx context.Context) (int, error)
	countAdminsFunc       func(ctx context.Context) (int, error)
}

func (m *mockAuthUserService) CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, email, name, oidcSubject)
	}
	return &models.User{Email: email}, nil
}
func (m *mockAuthUserService) GetOrCreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	if m.getOrCreateUserFunc != nil {
		return m.getOrCreateUserFunc(ctx, email, name, oidcSubject)
	}
	return &models.User{Email: email}, nil
}
func (m *mockAuthUserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return &models.User{ID: id}, nil
}
func (m *mockAuthUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return &models.User{Email: email}, nil
}
func (m *mockAuthUserService) UpdateUser(ctx context.Context, user *models.User) error {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(ctx, user)
	}
	return nil
}
func (m *mockAuthUserService) UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error {
	if m.updateUserRoleFunc != nil {
		return m.updateUserRoleFunc(ctx, userID, role)
	}
	return nil
}
func (m *mockAuthUserService) UpdateLastLogin(ctx context.Context, userID int64) error {
	if m.updateLastLoginFunc != nil {
		return m.updateLastLoginFunc(ctx, userID)
	}
	return nil
}
func (m *mockAuthUserService) DeleteUser(ctx context.Context, id int64) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(ctx, id)
	}
	return nil
}
func (m *mockAuthUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	if m.listUsersFunc != nil {
		return m.listUsersFunc(ctx, limit, offset)
	}
	return []*models.User{}, nil
}
func (m *mockAuthUserService) CountUsers(ctx context.Context) (int, error) {
	if m.countUsersFunc != nil {
		return m.countUsersFunc(ctx)
	}
	return 0, nil
}
func (m *mockAuthUserService) CountAdmins(ctx context.Context) (int, error) {
	if m.countAdminsFunc != nil {
		return m.countAdminsFunc(ctx)
	}
	return 0, nil
}

var _ auth.UserService = (*mockAuthUserService)(nil)

type mockAuthSessionManager struct {
	createSessionFunc       func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error)
	getSessionFunc          func(ctx context.Context, sessionID string) (*models.Session, error)
	refreshSessionFunc      func(ctx context.Context, sessionID string) error
	deleteSessionFunc       func(ctx context.Context, sessionID string) error
	deleteUserSessionsFunc  func(ctx context.Context, userID int64) error
	cleanupExpiredFunc      func(ctx context.Context) (int64, error)
	setSessionCookieFunc    func(w http.ResponseWriter, sessionID string) error
	clearSessionCookieFunc  func(w http.ResponseWriter) error
	getSessionFromReqFunc   func(r *http.Request) (string, error)
}

func (m *mockAuthSessionManager) CreateSession(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, userID, r)
	}
	return &models.Session{ID: "sess"}, nil
}
func (m *mockAuthSessionManager) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	if m.getSessionFunc != nil {
		return m.getSessionFunc(ctx, sessionID)
	}
	return &models.Session{ID: sessionID}, nil
}
func (m *mockAuthSessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	if m.refreshSessionFunc != nil {
		return m.refreshSessionFunc(ctx, sessionID)
	}
	return nil
}
func (m *mockAuthSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(ctx, sessionID)
	}
	return nil
}
func (m *mockAuthSessionManager) DeleteUserSessions(ctx context.Context, userID int64) error {
	if m.deleteUserSessionsFunc != nil {
		return m.deleteUserSessionsFunc(ctx, userID)
	}
	return nil
}
func (m *mockAuthSessionManager) CleanupExpired(ctx context.Context) (int64, error) {
	if m.cleanupExpiredFunc != nil {
		return m.cleanupExpiredFunc(ctx)
	}
	return 0, nil
}
func (m *mockAuthSessionManager) SetSessionCookie(w http.ResponseWriter, sessionID string) error {
	if m.setSessionCookieFunc != nil {
		return m.setSessionCookieFunc(w, sessionID)
	}
	return nil
}
func (m *mockAuthSessionManager) ClearSessionCookie(w http.ResponseWriter) error {
	if m.clearSessionCookieFunc != nil {
		return m.clearSessionCookieFunc(w)
	}
	return nil
}
func (m *mockAuthSessionManager) GetSessionFromRequest(r *http.Request) (string, error) {
	if m.getSessionFromReqFunc != nil {
		return m.getSessionFromReqFunc(r)
	}
	return "", nil
}

var _ auth.SessionManager = (*mockAuthSessionManager)(nil)

func TestOIDCCallback_GetOrCreateUserError(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "x@example.com"}, nil
		},
	}
	userSvc := &mockAuthUserService{
		getOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h := &AuthHandlers{authenticator: mockAuth, userService: userSvc, sessionMgr: &mockAuthSessionManager{}}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc", nil)
	w := httptest.NewRecorder()
	h.OIDCCallback(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestOIDCCallback_CreateSessionError(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "x@example.com"}, nil
		},
	}
	userSvc := &mockAuthUserService{}
	sessMgr := &mockAuthSessionManager{
		createSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
			return nil, fmt.Errorf("session error")
		},
	}
	h := &AuthHandlers{authenticator: mockAuth, userService: userSvc, sessionMgr: sessMgr}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc", nil)
	w := httptest.NewRecorder()
	h.OIDCCallback(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestOIDCCallback_SetSessionCookieError(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "x@example.com"}, nil
		},
	}
	userSvc := &mockAuthUserService{}
	sessMgr := &mockAuthSessionManager{
		setSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
			return fmt.Errorf("cookie error")
		},
	}
	h := &AuthHandlers{authenticator: mockAuth, userService: userSvc, sessionMgr: sessMgr}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc", nil)
	w := httptest.NewRecorder()
	h.OIDCCallback(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestOIDCCallback_ReturnURLCookieFallback(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "x@example.com"}, nil
		},
	}
	h := &AuthHandlers{
		authenticator: mockAuth,
		userService:   &mockAuthUserService{},
		sessionMgr:    &mockAuthSessionManager{},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc", nil)
	req.AddCookie(&http.Cookie{Name: auth.ReturnURLCookieName, Value: "/events/42"})
	w := httptest.NewRecorder()

	h.OIDCCallback(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w.Code)
	}
	location := w.Header().Get("Location")
	if location != "/events/42" {
		t.Errorf("expected redirect to /events/42 (from cookie), got %v", location)
	}
}

func TestOIDCCallback_UpdateLastLoginError(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "x@example.com"}, nil
		},
	}
	userSvc := &mockAuthUserService{
		updateLastLoginFunc: func(ctx context.Context, userID int64) error {
			return fmt.Errorf("db error")
		},
	}
	h := &AuthHandlers{
		authenticator: mockAuth,
		userService:   userSvc,
		sessionMgr:    &mockAuthSessionManager{},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc", nil)
	w := httptest.NewRecorder()
	h.OIDCCallback(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 (flow should continue despite UpdateLastLogin error), got %d", w.Code)
	}
}

func TestOIDCCallback_InvalidReturnURLFallback(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "x@example.com"}, nil
		},
	}
	h := &AuthHandlers{
		authenticator: mockAuth,
		userService:   &mockAuthUserService{},
		sessionMgr:    &mockAuthSessionManager{},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc&return=javascript:alert(1)", nil)
	w := httptest.NewRecorder()
	h.OIDCCallback(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w.Code)
	}
	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("expected redirect to / for invalid return URL, got %v", location)
	}
}

func TestOIDCCallback_ReturnURLCookieCleared(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "x@example.com"}, nil
		},
	}
	h := &AuthHandlers{
		authenticator: mockAuth,
		userService:   &mockAuthUserService{},
		sessionMgr:    &mockAuthSessionManager{},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc", nil)
	req.AddCookie(&http.Cookie{Name: auth.ReturnURLCookieName, Value: "/dashboard"})
	w := httptest.NewRecorder()
	h.OIDCCallback(w, req)

	var clearedCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == auth.ReturnURLCookieName {
			clearedCookie = c
			break
		}
	}
	if clearedCookie == nil {
		t.Fatal("expected return URL cookie to be cleared")
	}
	if clearedCookie.MaxAge != -1 {
		t.Errorf("expected MaxAge=-1, got %d", clearedCookie.MaxAge)
	}
	if clearedCookie.Value != "" {
		t.Errorf("expected Value to be empty, got %q", clearedCookie.Value)
	}
}

func TestOIDCCallback_NilResult(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return nil, nil
		},
	}
	h := &AuthHandlers{
		authenticator: mockAuth,
		userService:   &mockAuthUserService{},
		sessionMgr:    &mockAuthSessionManager{},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc", nil)
	w := httptest.NewRecorder()
	h.OIDCCallback(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for nil result, got %d", w.Code)
	}
}

func TestLogout_SessionCookieCleared(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}
	sessMgr := &mockAuthSessionManager{
		clearSessionCookieFunc: func(w http.ResponseWriter) error {
			http.SetCookie(w, &http.Cookie{
				Name:   auth.SessionCookieName,
				Value:  "",
				MaxAge: -1,
			})
			return nil
		},
	}
	h := &AuthHandlers{authenticator: mockAuth, sessionMgr: sessMgr}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	h.Logout(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w.Code)
	}

	cleared := false
	for _, c := range w.Result().Cookies() {
		if c.Name == auth.SessionCookieName && c.MaxAge == -1 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("expected session cookie to be cleared (MaxAge=-1)")
	}
}
