package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Decision — записанное решение Decide: какой вариант получил пользователь.
// decision_id детерминирован, поэтому повторный запрос не создаёт дубль.
type Decision struct {
	DecisionID        uuid.UUID       `json:"decision_id" gorm:"primaryKey;type:uuid"`
	ExperimentID      uuid.UUID       `json:"experiment_id" gorm:"type:uuid;not null;index"`
	ExperimentVersion int             `json:"experiment_version" gorm:"not null"`
	FlagKey           string          `json:"flag_key" gorm:"type:varchar(255);not null"`
	SubjectID         string          `json:"subject_id" gorm:"type:varchar(255);not null;index"`
	Variant           string          `json:"variant" gorm:"type:varchar(255);not null"`
	Attributes        json.RawMessage `json:"attributes,omitempty" gorm:"type:jsonb"`
	DecidedAt         time.Time       `json:"decided_at" gorm:"not null"`
}

// Event — принятое событие. event_id — ключ дедупликации.
type Event struct {
	EventID     string          `json:"event_id" gorm:"primaryKey;type:varchar(128)"`
	EventTypeID uuid.UUID       `json:"event_type_id" gorm:"type:uuid;not null;index"`
	DecisionID  *uuid.UUID      `json:"decision_id,omitempty" gorm:"type:uuid;index:idx_events_decision_ts,priority:1"`
	SubjectID   string          `json:"subject_id" gorm:"type:varchar(255);not null"`
	EventTS     time.Time       `json:"event_ts" gorm:"not null;index:idx_events_decision_ts,priority:2;index"`
	ReceivedAt  time.Time       `json:"received_at" gorm:"not null"`
	Properties  json.RawMessage `json:"properties,omitempty" gorm:"type:jsonb"`
}

// EventRow — событие вместе с настройками его типа (для атрибуции).
type EventRow struct {
	Event
	TypeKey          string
	RequiresExposure bool
	IsExposure       bool
}
