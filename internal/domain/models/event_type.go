package models

import (
	"time"

	"github.com/google/uuid"
)

type EventType struct {
	ID               uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Key              string     `json:"key" gorm:"type:varchar(64);not null;uniqueIndex"`
	Name             string     `json:"name" gorm:"type:varchar(255);not null"`
	Description      string     `json:"description" gorm:"type:text"`
	RequiresExposure bool       `json:"requires_exposure" gorm:"not null;default:false"`
	IsExposure       bool       `json:"is_exposure" gorm:"not null;default:false"`
	ArchivedAt       *time.Time `json:"archived_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (e EventType) IsArchived() bool { return e.ArchivedAt != nil }

type EventTyp string

var EventTypeFound = map[EventTyp]bool{
	"shown":              true,
	"added_to_favorites": true,
	"click":              true,
	"purchase":           true,
	"error":              true,
	"api_request":        true,
}
