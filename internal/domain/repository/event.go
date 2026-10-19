package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
	"time"
)

type EventWriter interface {
	CreateEvent(
		ctx context.Context,
		e models.Event,
	) (models.Event, error)
}
type EventReader interface {
	GetEventByID(
		ctx context.Context,
		id uuid.UUID,
	) (models.Event, error)
	ListEvents(ctx context.Context) ([]models.Event, error)
	CountEventsByEventType(
		ctx context.Context,
		eventType string,
		experimentID uuid.UUID,
		variantID uuid.UUID,
	) (int, error)
	CountEventsByEventTypeForGuard(
		ctx context.Context,
		eventType string,
		experimentID uuid.UUID,
		variantID uuid.UUID,
		from, to time.Time,
	) (int, error)
}
type EventRepository interface {
	EventWriter
	EventReader
}
