package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mocksvcs "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
	"go.uber.org/mock/gomock"
)

func TestCleanupExpiredTokensHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockSvc.EXPECT().CleanupExpiredTokens(gomock.Any()).Return(int64(5), nil)

	handler := NewCleanupHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/invites/cleanup", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if deleted, ok := response["deleted"].(float64); !ok || int64(deleted) != 5 {
		t.Errorf("expected deleted=5, got %v", response["deleted"])
	}

	if message, ok := response["message"].(string); !ok || message == "" {
		t.Errorf("expected non-empty message, got %v", response["message"])
	}
}

func TestCleanupExpiredTokensHandler_NoExpiredTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockSvc.EXPECT().CleanupExpiredTokens(gomock.Any()).Return(int64(0), nil)

	handler := NewCleanupHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/invites/cleanup", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if deleted, ok := response["deleted"].(float64); !ok || int64(deleted) != 0 {
		t.Errorf("expected deleted=0, got %v", response["deleted"])
	}
}

func TestCleanupExpiredTokensHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockSvc.EXPECT().CleanupExpiredTokens(gomock.Any()).Return(int64(0), errors.New("database error"))

	handler := NewCleanupHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/invites/cleanup", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if errorMsg, ok := response["error"].(string); !ok || errorMsg == "" {
		t.Errorf("expected non-empty error message, got %v", response["error"])
	}
}

func TestCleanupExpiredTokensHandler_InvalidMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	handler := NewCleanupHandler(mockSvc)

	tests := []struct {
		name   string
		method string
	}{
		{"GET not allowed", http.MethodGet},
		{"PUT not allowed", http.MethodPut},
		{"DELETE not allowed", http.MethodDelete},
		{"PATCH not allowed", http.MethodPatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/invites/cleanup", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
			}
		})
	}
}
