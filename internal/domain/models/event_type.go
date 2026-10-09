package models

import (
	"time"

	"github.com/google/uuid"
)

// EventType — запись каталога событий. Типы событий настраиваются данными, а не кодом.
type EventType struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Key         string    `json:"key" gorm:"type:varchar(64);not null;uniqueIndex"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`

	// RequiresExposure: событие засчитывается, только если у его decision_id был показ (exposure).
	RequiresExposure bool `json:"requires_exposure" gorm:"not null;default:false"`
	// IsExposure: само событие является показом варианта. Активным может быть только один такой тип.
	IsExposure bool `json:"is_exposure" gorm:"not null;default:false"`

	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (e EventType) IsArchived() bool { return e.ArchivedAt != nil }
