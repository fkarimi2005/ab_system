package models

import (
	"github.com/google/uuid"
	"time"
)

type ExperimentVersion struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExperimentID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_exp_version"`
	Version      int       `gorm:"not null;uniqueIndex:idx_exp_version"`
	Snapshot     []byte    `gorm:"type:jsonb;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
