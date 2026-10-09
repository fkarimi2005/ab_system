package models

import (
	"time"

	"github.com/google/uuid"
)

type MetricKind string

const (
	MetricKindCount             MetricKind = "count"              // число событий
	MetricKindConversion        MetricKind = "conversion"         // доля показов с целевым событием
	MetricKindErrorRate         MetricKind = "error_rate"         // доля показов с событием-ошибкой
	MetricKindLatencyPercentile MetricKind = "latency_percentile" // перцентиль числового поля события
)

// Metric — запись каталога метрик.
type Metric struct {
	ID   uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Key  string     `json:"key" gorm:"type:varchar(64);not null;uniqueIndex"`
	Name string     `json:"name" gorm:"type:varchar(255);not null"`
	Kind MetricKind `json:"kind" gorm:"type:varchar(32);not null"`

	// EventTypeID — событие-источник метрики.
	EventTypeID uuid.UUID `json:"event_type_id" gorm:"type:uuid;not null;index"`
	// DenominatorEventTypeID — знаменатель для conversion и error_rate;
	// если пусто, в отчёте берётся тип показа (exposure).
	DenominatorEventTypeID *uuid.UUID `json:"denominator_event_type_id,omitempty" gorm:"type:uuid"`

	// Только для latency_percentile: какой перцентиль и из какого поля properties события.
	Percentile *int   `json:"percentile,omitempty"`
	ValueField string `json:"value_field,omitempty" gorm:"type:varchar(64)"`

	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (m Metric) IsArchived() bool { return m.ArchivedAt != nil }
