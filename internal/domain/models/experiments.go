package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

const FullAudienceBP = 10000 // 100% аудитории

type Experiment struct {
	ID            uuid.UUID        `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FeatureFlagID uuid.UUID        `json:"feature_id" gorm:"type:uuid;not null;index"`
	FeatureFlag   FeatureFlag      `json:"-" gorm:"foreignKey:FeatureFlagID;references:ID"`
	Name          string           `json:"name" gorm:"not null;type:varchar(255)"`
	Status        ExperimentStatus `json:"status" gorm:"not null;type:varchar(35);default:'draft'"`

	AudienceBP int `json:"audience_bp" gorm:"not null;check:chk_experiments_audience,audience_bp > 0 AND audience_bp <= 10000"`

	Version int       `json:"version" gorm:"not null;default:1"`
	OwnerID uuid.UUID `json:"owner_id" gorm:"type:uuid;not null;index"`
	User    User      `json:"-" gorm:"foreignKey:OwnerID;references:ID"`

	Variants []ExperimentVariant `json:"variants" gorm:"foreignKey:ExperimentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ExperimentVariant struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ExperimentID uuid.UUID `json:"experiment_id" gorm:"type:uuid;not null;uniqueIndex:idx_variant_name"`
	Name         string    `json:"name" gorm:"not null;type:varchar(255);uniqueIndex:idx_variant_name"`
	Value        string    `json:"value" gorm:"not null;type:text"`

	Weight    int  `json:"weight" gorm:"not null;check:chk_variants_weight,weight > 0"`
	IsControl bool `json:"is_control" gorm:"not null;default:false"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
type ExperimentStatus string

const (
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusReview    ExperimentStatus = "review"
	ExperimentStatusApproved  ExperimentStatus = "approved"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusArchived  ExperimentStatus = "archived"
	ExperimentStatusRejected  ExperimentStatus = "rejected"
)
