package handlers

import (
	"context"
	"strconv"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/eventid"
)

type eventIDResolver interface {
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	GetByPublicID(ctx context.Context, publicID string) (*models.Event, error)
	GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error)
}

func resolveEventID(ctx context.Context, repo eventIDResolver, idParam string) (*models.Event, error) {
	if idParam == "" {
		return nil, NewBadRequestError("event ID cannot be empty")
	}

	numericID, err := strconv.ParseInt(idParam, 10, 64)
	if err == nil && numericID > 0 {
		event, err := repo.GetByID(ctx, numericID)
		if err == nil {
			return event, nil
		}
		if _, ok := err.(*models.NotFoundError); !ok {
			return nil, err
		}
	}

	if err := eventid.ValidateEventID(idParam); err == nil {
		event, err := repo.GetByPublicID(ctx, idParam)
		if err == nil {
			return event, nil
		}
		if _, ok := err.(*models.NotFoundError); !ok {
			return nil, err
		}
	}

	event, err := repo.GetByFriendlyName(ctx, idParam)
	if err == nil {
		return event, nil
	}

	return nil, &models.NotFoundError{
		Resource: "Event",
		ID:       idParam,
	}
}

func resolveEventIDFromRepo(ctx context.Context, repo repositories.EventRepository, idParam string) (*models.Event, error) {
	return resolveEventID(ctx, repo, idParam)
}
