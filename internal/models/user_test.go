package models

import (
	"testing"
	"time"
)

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{
			name: "admin role returns true",
			role: RoleAdmin,
			want: true,
		},
		{
			name: "event_manager role returns false",
			role: RoleEventManager,
			want: false,
		},
		{
			name: "empty role returns false",
			role: "",
			want: false,
		},
		{
			name: "invalid role returns false",
			role: "invalid_role",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			if got := u.IsAdmin(); got != tt.want {
				t.Errorf("User.IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsEventManager(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{
			name: "admin role returns true",
			role: RoleAdmin,
			want: true,
		},
		{
			name: "event_manager role returns true",
			role: RoleEventManager,
			want: true,
		},
		{
			name: "empty role returns false",
			role: "",
			want: false,
		},
		{
			name: "invalid role returns false",
			role: "invalid_role",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			if got := u.IsEventManager(); got != tt.want {
				t.Errorf("User.IsEventManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserRole_Constants(t *testing.T) {
	t.Run("role constants have correct values", func(t *testing.T) {
		if RoleAdmin != "admin" {
			t.Errorf("RoleAdmin = %v, want 'admin'", RoleAdmin)
		}
		if RoleEventManager != "event_manager" {
			t.Errorf("RoleEventManager = %v, want 'event_manager'", RoleEventManager)
		}
	})
}

func TestUser_StructFields(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(-24 * time.Hour)
	oidcSubject := "google-oauth2|123456"

	user := &User{
		ID:          123,
		Email:       "test@example.com",
		Name:        "Test User",
		Role:        RoleEventManager,
		OIDCSubject: &oidcSubject,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: &lastLogin,
	}

	t.Run("all fields are accessible", func(t *testing.T) {
		if user.ID != 123 {
			t.Errorf("ID = %v, want 123", user.ID)
		}
		if user.Email != "test@example.com" {
			t.Errorf("Email = %v, want 'test@example.com'", user.Email)
		}
		if user.Name != "Test User" {
			t.Errorf("Name = %v, want 'Test User'", user.Name)
		}
		if user.Role != RoleEventManager {
			t.Errorf("Role = %v, want RoleEventManager", user.Role)
		}
		if user.OIDCSubject == nil || *user.OIDCSubject != oidcSubject {
			t.Errorf("OIDCSubject = %v, want %v", user.OIDCSubject, oidcSubject)
		}
		if user.CreatedAt.IsZero() {
			t.Error("CreatedAt should not be zero")
		}
		if user.UpdatedAt.IsZero() {
			t.Error("UpdatedAt should not be zero")
		}
		if user.LastLoginAt == nil {
			t.Error("LastLoginAt should not be nil")
		}
	})

	t.Run("optional fields can be nil", func(t *testing.T) {
		minimalUser := &User{
			ID:        1,
			Email:     "minimal@example.com",
			Name:      "Minimal",
			Role:      RoleAdmin,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if minimalUser.OIDCSubject != nil {
			t.Error("OIDCSubject should be nil for minimal user")
		}
		if minimalUser.LastLoginAt != nil {
			t.Error("LastLoginAt should be nil for minimal user")
		}
	})
}

func TestUser_RolePermissions(t *testing.T) {
	t.Run("admin has all permissions", func(t *testing.T) {
		admin := &User{Role: RoleAdmin}
		if !admin.IsAdmin() {
			t.Error("Admin should have IsAdmin() = true")
		}
		if !admin.IsEventManager() {
			t.Error("Admin should have IsEventManager() = true")
		}
	})

	t.Run("event manager has limited permissions", func(t *testing.T) {
		manager := &User{Role: RoleEventManager}
		if manager.IsAdmin() {
			t.Error("Event manager should have IsAdmin() = false")
		}
		if !manager.IsEventManager() {
			t.Error("Event manager should have IsEventManager() = true")
		}
	})
}
