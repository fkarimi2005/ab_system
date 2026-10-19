package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"time"
)

type Event struct {
	EventID      uuid.UUID         `json:"event_id" gorm:"type:uuid;not null;uniqueIndex"`
	ExperimentID uuid.UUID         `json:"experiment_id" gorm:"type:uuid;not null"`
	EventType    EventTyp          `json:"event_type" gorm:"type:varchar(55);not null"`
	UserID       uuid.UUID         `json:"user_id" gorm:"type:uuid;not null"`
	DecisionID   uuid.UUID         `json:"decision_id" gorm:"type:uuid;not null"`
	Properties   datatypes.JSONMap `json:"properties" gorm:"type:text;not null"`
	CreatedAt    time.Time         `json:"created_at" gorm:"autoCreateTime;not null;"`
	EventTS      time.Time         `json:"event_ts" gorm:"not null;index"`
}
