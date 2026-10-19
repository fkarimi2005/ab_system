package models

import (
	"github.com/google/uuid"
	"time"
)

type Report struct {
	ExperimentID uuid.UUID `json:"experiment_id" gorm:"primary_key;type:uuid;not null"`
	MetricID     uuid.UUID `json:"metric_id" gorm:"primary_key;type:uuid;not null"`
	VariantID    uuid.UUID `json:"variant_id" gorm:"primary_key;type:uuid;not null"`
	Count        float64   `json:"count" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
}
type ReportResult struct {
	ExperimentID  uuid.UUID       `json:"experiment_id"`
	VariantReport []VariantReport `json:"variant_report"`
}
type VariantReport struct {
	VariantID uuid.UUID         `json:"variant_id"`
	Metrics   []MetricWithCount `json:"metrics"`
}
type MetricWithCount struct {
	MetricKey string  `json:"metric_key"`
	Count     float64 `json:"count"`
}
