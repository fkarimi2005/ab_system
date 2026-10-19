package models

import (
	"github.com/google/uuid"
	"time"
)

type MetricRole string

const (
	MetricRoleTarget     MetricRole = "target"
	MetricRoleDiagnostic MetricRole = "diagnostic"
)

type ExperimentMetric struct {
	ExperimentID uuid.UUID  `json:"-" gorm:"type:uuid;primaryKey"`
	MetricID     uuid.UUID  `json:"metric_id" gorm:"type:uuid;primaryKey"`
	Role         MetricRole `json:"role" gorm:"type:varchar(16);not null"`
}

type GuardrailComparison string

const (
	GuardrailAbove GuardrailComparison = "gt"
	GuardrailBelow GuardrailComparison = "lt"
)

type GuardrailAction string

const (
	GuardrailPause           GuardrailAction = "pause"
	GuardrailRevertToControl GuardrailAction = "revert_to_control"
)

type ExperimentGuardrail struct {
	ID            uuid.UUID           `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ExperimentID  uuid.UUID           `json:"-" gorm:"type:uuid;not null;uniqueIndex:idx_guardrail_metric"`
	MetricID      uuid.UUID           `json:"metric_id" gorm:"type:uuid;not null;uniqueIndex:idx_guardrail_metric"`
	Threshold     float64             `json:"threshold" gorm:"not null"`
	Comparison    GuardrailComparison `json:"comparison" gorm:"type:varchar(4);not null"`
	WindowSeconds int                 `json:"window_seconds" gorm:"not null"`
	Action        GuardrailAction     `json:"action" gorm:"type:varchar(32);not null"`
}
type GuardrailTrigger struct {
	ID            uuid.UUID           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ExperimentID  uuid.UUID           `gorm:"type:uuid;not null;index"`
	VariantID     uuid.UUID           `gorm:"type:uuid;not null"`
	MetricID      uuid.UUID           `gorm:"type:uuid;not null"`
	Value         float64             `gorm:"not null"`
	Threshold     float64             `gorm:"not null"`
	Comparison    GuardrailComparison `gorm:"type:varchar(4);not null"`
	WindowSeconds int                 `gorm:"not null"`
	Action        GuardrailAction     `gorm:"type:varchar(32);not null"`
	TriggeredAt   time.Time           `gorm:"not null"`
}
