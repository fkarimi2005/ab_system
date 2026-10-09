package models

import "github.com/google/uuid"

type MetricRole string

const (
	MetricRoleTarget     MetricRole = "target"
	MetricRoleDiagnostic MetricRole = "diagnostic"
)

// ExperimentMetric — метрика, привязанная к эксперименту.
type ExperimentMetric struct {
	ExperimentID uuid.UUID  `json:"-" gorm:"type:uuid;primaryKey"`
	MetricID     uuid.UUID  `json:"metric_id" gorm:"type:uuid;primaryKey"`
	Role         MetricRole `json:"role" gorm:"type:varchar(16);not null"`
}

type GuardrailComparison string

const (
	GuardrailAbove GuardrailComparison = "gt" // срабатывает, когда значение больше порога
	GuardrailBelow GuardrailComparison = "lt" // срабатывает, когда значение меньше порога
)

type GuardrailAction string

const (
	GuardrailPause           GuardrailAction = "pause"
	GuardrailRevertToControl GuardrailAction = "revert_to_control"
)

// ExperimentGuardrail — автоматический «предохранитель» эксперимента.
type ExperimentGuardrail struct {
	ID            uuid.UUID           `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ExperimentID  uuid.UUID           `json:"-" gorm:"type:uuid;not null;uniqueIndex:idx_guardrail_metric"`
	MetricID      uuid.UUID           `json:"metric_id" gorm:"type:uuid;not null;uniqueIndex:idx_guardrail_metric"`
	Threshold     float64             `json:"threshold" gorm:"not null"`
	Comparison    GuardrailComparison `json:"comparison" gorm:"type:varchar(4);not null"`
	WindowSeconds int                 `json:"window_seconds" gorm:"not null"`
	Action        GuardrailAction     `json:"action" gorm:"type:varchar(32);not null"`
}
