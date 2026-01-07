package auth

import (
	"context"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionContextKey contextKey = "session"
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}

func WithSession(ctx context.Context, session *models.Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

func SessionFromContext(ctx context.Context) (*models.Session, bool) {
	session, ok := ctx.Value(sessionContextKey).(*models.Session)
	return session, ok
}
