package dto

import (
	"AB_system/internal/domain/models"
	"github.com/google/uuid"
	"time"
)

type EventRequest struct {
	EventID      uuid.UUID       `json:"event_id"`
	ExperimentID uuid.UUID       `json:"experiment_id"`
	UserID       uuid.UUID       `json:"user_id"`
	EventType    models.EventTyp `json:"event_type"`
	DecisionID   uuid.UUID       `json:"decision_id"`
	Properties   map[string]any  `json:"properties"`
}
type EventResponse struct {
	ExperimenterID uuid.UUID       `json:"experimenter_id"`
	UserID         uuid.UUID       `json:"user_id"`
	EventType      models.EventTyp `json:"event_type"`
	Properties     map[string]any  `json:"properties"`
	DecisionID     uuid.UUID       `json:"decision_id"`
	CreatedAt      time.Time       `json:"created_at"`
}

func NewEventResponse(event *models.Event) EventResponse {
	return EventResponse{
		ExperimenterID: event.ExperimentID,
		UserID:         event.UserID,
		EventType:      event.EventType,
		DecisionID:     event.DecisionID,
		Properties:     event.Properties,

		CreatedAt: event.CreatedAt,
	}
}
