package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/tinyrsvp/internal/auth"
	"github.com/yourusername/tinyrsvp/internal/models"
)

type UserService interface {
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error
	DeleteUser(ctx context.Context, id int64) error
	CountUsers(ctx context.Context) (int, error)
	CountAdmins(ctx context.Context) (int, error)
}

type UserHandler struct {
	userService UserService
	authChecker auth.AuthorizationChecker
}

func NewUserHandler(userService UserService, authChecker auth.AuthorizationChecker) *UserHandler {
	return &UserHandler{
		userService: userService,
		authChecker: authChecker,
	}
}

type ListUsersResponse struct {
	Users  []UserDTO `json:"users"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

type UserDTO struct {
	ID          int64      `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        string     `json:"role"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	if !h.authChecker.CanManageUsers(r.Context(), user) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			respondError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		limit = parsedLimit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			respondError(w, http.StatusBadRequest, "invalid offset parameter")
			return
		}
		offset = parsedOffset
	}

	users, err := h.userService.ListUsers(r.Context(), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	total, err := h.userService.CountUsers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to count users")
		return
	}

	userDTOs := make([]UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = toUserDTO(user)
	}

	response := ListUsersResponse{
		Users:  userDTOs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request, userIDStr string) {
	currentUser, _ := auth.UserFromContext(r.Context())
	if !h.authChecker.CanManageUsers(r.Context(), currentUser) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	userID, err := parseUserID(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, toUserDTO(user))
}

func (h *UserHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request, userIDStr string) {
	currentUser, _ := auth.UserFromContext(r.Context())
	if !h.authChecker.CanManageUsers(r.Context(), currentUser) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	userID, err := parseUserID(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newRole, err := validateRole(req.Role)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	if user.Role == models.RoleAdmin && newRole != models.RoleAdmin {
		adminCount, err := h.userService.CountAdmins(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to check admin count")
			return
		}

		if adminCount <= 1 {
			respondError(w, http.StatusConflict, "cannot demote last admin")
			return
		}
	}

	if err := h.userService.UpdateUserRole(r.Context(), userID, newRole); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update user role")
		return
	}

	user.Role = newRole
	respondJSON(w, http.StatusOK, toUserDTO(user))
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, userIDStr string) {
	currentUser, _ := auth.UserFromContext(r.Context())
	if !h.authChecker.CanManageUsers(r.Context(), currentUser) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	userID, err := parseUserID(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	if user.Role == models.RoleAdmin {
		adminCount, err := h.userService.CountAdmins(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to check admin count")
			return
		}

		if adminCount <= 1 {
			respondError(w, http.StatusConflict, "cannot delete last admin")
			return
		}
	}

	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseUserID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format")
	}

	if id <= 0 {
		return 0, fmt.Errorf("user ID must be positive")
	}

	return id, nil
}

func validateRole(roleStr string) (models.UserRole, error) {
	switch roleStr {
	case "admin":
		return models.RoleAdmin, nil
	case "event_manager":
		return models.RoleEventManager, nil
	default:
		return "", fmt.Errorf("invalid role: must be 'admin' or 'event_manager'")
	}
}

func toUserDTO(user *models.User) UserDTO {
	return UserDTO{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Role:        string(user.Role),
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
