package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/tinyrsvp/internal/models"
)

type mockUserService struct {
	ListUsersFunc      func(ctx context.Context, limit, offset int) ([]*models.User, error)
	GetUserByIDFunc    func(ctx context.Context, id int64) (*models.User, error)
	UpdateUserRoleFunc func(ctx context.Context, userID int64, role models.UserRole) error
	DeleteUserFunc     func(ctx context.Context, id int64) error
	CountUsersFunc     func(ctx context.Context) (int, error)
	CountAdminsFunc    func(ctx context.Context) (int, error)
}

func (m *mockUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, limit, offset)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserService) UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error {
	if m.UpdateUserRoleFunc != nil {
		return m.UpdateUserRoleFunc(ctx, userID, role)
	}
	return errors.New("not implemented")
}

func (m *mockUserService) DeleteUser(ctx context.Context, id int64) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *mockUserService) CountUsers(ctx context.Context) (int, error) {
	if m.CountUsersFunc != nil {
		return m.CountUsersFunc(ctx)
	}
	return 0, errors.New("not implemented")
}

func (m *mockUserService) CountAdmins(ctx context.Context) (int, error) {
	if m.CountAdminsFunc != nil {
		return m.CountAdminsFunc(ctx)
	}
	return 0, errors.New("not implemented")
}

func TestListUsers(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   string
		mockUsers     []*models.User
		mockCount     int
		mockListErr   error
		mockCountErr  error
		wantStatus    int
		wantUserCount int
		wantTotal     int
	}{
		{
			name:        "list users successfully",
			queryParams: "",
			mockUsers: []*models.User{
				{ID: 1, Email: "user1@example.com", Name: "User 1", Role: models.RoleAdmin, CreatedAt: time.Now()},
				{ID: 2, Email: "user2@example.com", Name: "User 2", Role: models.RoleEventManager, CreatedAt: time.Now()},
			},
			mockCount:     2,
			wantStatus:    http.StatusOK,
			wantUserCount: 2,
			wantTotal:     2,
		},
		{
			name:        "list users with pagination",
			queryParams: "?limit=1&offset=1",
			mockUsers: []*models.User{
				{ID: 2, Email: "user2@example.com", Name: "User 2", Role: models.RoleEventManager, CreatedAt: time.Now()},
			},
			mockCount:     5,
			wantStatus:    http.StatusOK,
			wantUserCount: 1,
			wantTotal:     5,
		},
		{
			name:        "list users with invalid limit",
			queryParams: "?limit=invalid",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "list users with invalid offset",
			queryParams: "?offset=invalid",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "list users service error",
			queryParams: "",
			mockListErr: errors.New("database error"),
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name:         "list users count error",
			queryParams:  "",
			mockUsers:    []*models.User{},
			mockCountErr: errors.New("count error"),
			wantStatus:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{
				ListUsersFunc: func(ctx context.Context, limit, offset int) ([]*models.User, error) {
					if tt.mockListErr != nil {
						return nil, tt.mockListErr
					}
					return tt.mockUsers, nil
				},
				CountUsersFunc: func(ctx context.Context) (int, error) {
					if tt.mockCountErr != nil {
						return 0, tt.mockCountErr
					}
					return tt.mockCount, nil
				},
			}

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ListUsers(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var response ListUsersResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(response.Users) != tt.wantUserCount {
					t.Errorf("Expected %d users, got %d", tt.wantUserCount, len(response.Users))
				}

				if response.Total != tt.wantTotal {
					t.Errorf("Expected total %d, got %d", tt.wantTotal, response.Total)
				}
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mockUser   *models.User
		mockErr    error
		wantStatus int
	}{
		{
			name:   "get user successfully",
			userID: "1",
			mockUser: &models.User{
				ID:        1,
				Email:     "user@example.com",
				Name:      "Test User",
				Role:      models.RoleEventManager,
				CreatedAt: time.Now(),
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "get user invalid ID",
			userID:     "invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "get user not found",
			userID:     "999",
			mockErr:    &models.NotFoundError{Resource: "User", ID: int64(999)},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "get user service error",
			userID:     "1",
			mockErr:    errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{
				GetUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockUser, nil
				},
			}

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			handler.GetUser(w, req, tt.userID)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var response UserDTO
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.ID != tt.mockUser.ID {
					t.Errorf("Expected user ID %d, got %d", tt.mockUser.ID, response.ID)
				}
			}
		})
	}
}

func TestUpdateUserRole(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    UpdateRoleRequest
		mockUser       *models.User
		mockGetErr     error
		mockUpdateErr  error
		mockAdminCount int
		mockCountErr   error
		wantStatus     int
	}{
		{
			name:   "promote user to admin",
			userID: "1",
			requestBody: UpdateRoleRequest{
				Role: "admin",
			},
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			mockAdminCount: 2,
			wantStatus:     http.StatusOK,
		},
		{
			name:   "demote admin to event manager",
			userID: "1",
			requestBody: UpdateRoleRequest{
				Role: "event_manager",
			},
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			mockAdminCount: 2,
			wantStatus:     http.StatusOK,
		},
		{
			name:   "cannot demote last admin",
			userID: "1",
			requestBody: UpdateRoleRequest{
				Role: "event_manager",
			},
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			mockAdminCount: 1,
			wantStatus:     http.StatusConflict,
		},
		{
			name:   "invalid user ID",
			userID: "invalid",
			requestBody: UpdateRoleRequest{
				Role: "admin",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid role",
			userID: "1",
			requestBody: UpdateRoleRequest{
				Role: "invalid_role",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			requestBody: UpdateRoleRequest{
				Role: "admin",
			},
			mockGetErr: &models.NotFoundError{Resource: "User", ID: int64(999)},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "count admins error",
			userID: "1",
			requestBody: UpdateRoleRequest{
				Role: "event_manager",
			},
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			mockCountErr: errors.New("count error"),
			wantStatus:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{
				GetUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
					if tt.mockGetErr != nil {
						return nil, tt.mockGetErr
					}
					return tt.mockUser, nil
				},
				UpdateUserRoleFunc: func(ctx context.Context, userID int64, role models.UserRole) error {
					if tt.mockUpdateErr != nil {
						return tt.mockUpdateErr
					}
					return nil
				},
				CountAdminsFunc: func(ctx context.Context) (int, error) {
					if tt.mockCountErr != nil {
						return 0, tt.mockCountErr
					}
					return tt.mockAdminCount, nil
				},
			}

			handler := NewUserHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/users/"+tt.userID+"/role", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.UpdateUserRole(w, req, tt.userID)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockUser       *models.User
		mockGetErr     error
		mockDeleteErr  error
		mockAdminCount int
		mockCountErr   error
		wantStatus     int
	}{
		{
			name:   "delete event manager successfully",
			userID: "1",
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "delete admin successfully",
			userID: "1",
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			mockAdminCount: 2,
			wantStatus:     http.StatusNoContent,
		},
		{
			name:   "cannot delete last admin",
			userID: "1",
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			mockAdminCount: 1,
			wantStatus:     http.StatusConflict,
		},
		{
			name:       "invalid user ID",
			userID:     "invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "user not found",
			userID:     "999",
			mockGetErr: &models.NotFoundError{Resource: "User", ID: int64(999)},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "count admins error",
			userID: "1",
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			mockCountErr: errors.New("count error"),
			wantStatus:   http.StatusInternalServerError,
		},
		{
			name:   "delete error",
			userID: "1",
			mockUser: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			mockDeleteErr: errors.New("delete error"),
			wantStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{
				GetUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
					if tt.mockGetErr != nil {
						return nil, tt.mockGetErr
					}
					return tt.mockUser, nil
				},
				DeleteUserFunc: func(ctx context.Context, id int64) error {
					if tt.mockDeleteErr != nil {
						return tt.mockDeleteErr
					}
					return nil
				},
				CountAdminsFunc: func(ctx context.Context) (int, error) {
					if tt.mockCountErr != nil {
						return 0, tt.mockCountErr
					}
					return tt.mockAdminCount, nil
				},
			}

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			handler.DeleteUser(w, req, tt.userID)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
			}
		})
	}
}

func TestParseUserID(t *testing.T) {
	tests := []struct {
		name    string
		idStr   string
		wantID  int64
		wantErr bool
	}{
		{
			name:    "valid ID",
			idStr:   "123",
			wantID:  123,
			wantErr: false,
		},
		{
			name:    "invalid ID",
			idStr:   "invalid",
			wantErr: true,
		},
		{
			name:    "negative ID",
			idStr:   "-1",
			wantErr: true,
		},
		{
			name:    "zero ID",
			idStr:   "0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := parseUserID(tt.idStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseUserID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && id != tt.wantID {
				t.Errorf("Expected ID %d, got %d", tt.wantID, id)
			}
		})
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name     string
		roleStr  string
		wantRole models.UserRole
		wantErr  bool
	}{
		{
			name:     "valid admin role",
			roleStr:  "admin",
			wantRole: models.RoleAdmin,
			wantErr:  false,
		},
		{
			name:     "valid event_manager role",
			roleStr:  "event_manager",
			wantRole: models.RoleEventManager,
			wantErr:  false,
		},
		{
			name:    "invalid role",
			roleStr: "invalid_role",
			wantErr: true,
		},
		{
			name:    "empty role",
			roleStr: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := validateRole(tt.roleStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateRole() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && role != tt.wantRole {
				t.Errorf("Expected role %s, got %s", tt.wantRole, role)
			}
		})
	}
}
