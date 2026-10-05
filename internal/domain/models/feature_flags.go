package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type FeatureFlag struct {
	ID           uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Key          string         `json:"key" gorm:"type:varchar(255);not null;uniqueIndex:idx_feature_flags_key,where:deleted_at IS NULL"`
	ValueType    string         `json:"value_type" gorm:"type:varchar(20);not null"`
	DefaultValue string         `json:"default_value" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
