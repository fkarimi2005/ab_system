package models

import (
	"github.com/google/uuid"
	"time"
)

type ExperimentDecision struct {
	ID                uuid.UUID         `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	ExperimentID      uuid.UUID         `json:"experiment_id" gorm:"index;type:uuid; not null"`
	UserID            uuid.UUID         `json:"user_id" gorm:"index;type:uuid; not null"`
	VariantID         uuid.UUID         `json:"variant_id" gorm:"index;type:uuid; not null"`
	Experiment        Experiment        `json:"-" gorm:"foreignkey:ExperimentID;references:ID"`
	User              User              `json:"-" gorm:"foreignkey:UserID;references:ID"`
	ExperimentVariant ExperimentVariant `json:"-" gorm:"foreignkey:VariantID;references:ID"`
	CreatedAt         time.Time         `json:"created_at" gorm:"autoCreateTime"`
}
