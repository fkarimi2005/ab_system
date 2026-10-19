package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"github.com/google/uuid"
	"time"
)

type EventService struct {
	eventRepo repository.EventRepository
}

func NewEventService(e repository.EventRepository) *EventService {
	return &EventService{eventRepo: e}
}

func (s *EventService) Create(ctx context.Context, event models.Event) (models.Event, error) {
	if event.EventID == uuid.Nil {
		return models.Event{}, errs.ErrIdIsEmpty
	}
	if event.UserID == uuid.Nil || event.ExperimentID == uuid.Nil {
		return models.Event{}, errs.ErrIdIsEmpty
	}
	if event.EventType == "" {
		return models.Event{}, errs.ErrEventTypeNotFound
	}
	if !models.EventTypeFound[event.EventType] {
		return models.Event{}, errs.ErrEventTypeNotFound
	}
	if event.EventTS.IsZero() {
		event.EventTS = time.Now()
	}
	return s.eventRepo.CreateEvent(ctx, event)
}
func (s *EventService) Get(ctx context.Context, eventID uuid.UUID) (models.Event, error) {
	if eventID == uuid.Nil {
		return models.Event{}, errs.ErrIdIsEmpty
	}
	return s.eventRepo.GetEventByID(ctx, eventID)
}
func (s *EventService) List(ctx context.Context) ([]models.Event, error) {
	return s.eventRepo.ListEvents(ctx)
}
