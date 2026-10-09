package dto

import (
	"AB_system/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

// ---------- типы событий ----------

type CreateEventTypeRequest struct {
	Key              string `json:"key" binding:"required,max=64"`
	Name             string `json:"name" binding:"required,max=255"`
	Description      string `json:"description" binding:"max=2000"`
	RequiresExposure bool   `json:"requires_exposure"`
	IsExposure       bool   `json:"is_exposure"`
}

type UpdateEventTypeRequest struct {
	Name             string `json:"name" binding:"required,max=255"`
	Description      string `json:"description" binding:"max=2000"`
	RequiresExposure bool   `json:"requires_exposure"`
}

type EventTypeResponse struct {
	ID               uuid.UUID  `json:"id"`
	Key              string     `json:"key"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	RequiresExposure bool       `json:"requires_exposure"`
	IsExposure       bool       `json:"is_exposure"`
	ArchivedAt       *time.Time `json:"archived_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func NewEventTypeResponse(e models.EventType) EventTypeResponse {
	return EventTypeResponse{
		ID: e.ID, Key: e.Key, Name: e.Name, Description: e.Description,
		RequiresExposure: e.RequiresExposure, IsExposure: e.IsExposure,
		ArchivedAt: e.ArchivedAt, CreatedAt: e.CreatedAt,
	}
}

// ---------- метрики ----------

type CreateMetricRequest struct {
	Key                    string     `json:"key" binding:"required,max=64"`
	Name                   string     `json:"name" binding:"required,max=255"`
	Kind                   string     `json:"kind" binding:"required"`
	EventTypeID            uuid.UUID  `json:"event_type_id" binding:"required"`
	DenominatorEventTypeID *uuid.UUID `json:"denominator_event_type_id"`
	Percentile             *int       `json:"percentile"`
	ValueField             string     `json:"value_field" binding:"max=64"`
}

type RenameMetricRequest struct {
	Name string `json:"name" binding:"required,max=255"`
}

type MetricResponse struct {
	ID                     uuid.UUID  `json:"id"`
	Key                    string     `json:"key"`
	Name                   string     `json:"name"`
	Kind                   string     `json:"kind"`
	EventTypeID            uuid.UUID  `json:"event_type_id"`
	DenominatorEventTypeID *uuid.UUID `json:"denominator_event_type_id,omitempty"`
	Percentile             *int       `json:"percentile,omitempty"`
	ValueField             string     `json:"value_field,omitempty"`
	ArchivedAt             *time.Time `json:"archived_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
}

func NewMetricResponse(m models.Metric) MetricResponse {
	return MetricResponse{
		ID: m.ID, Key: m.Key, Name: m.Name, Kind: string(m.Kind), EventTypeID: m.EventTypeID,
		DenominatorEventTypeID: m.DenominatorEventTypeID, Percentile: m.Percentile, ValueField: m.ValueField,
		ArchivedAt: m.ArchivedAt, CreatedAt: m.CreatedAt,
	}
}
