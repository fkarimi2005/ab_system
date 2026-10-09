package repository

import (
	"AB_system/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type EventTypeRepository interface {
	CreateEventType(ctx context.Context, e *models.EventType) (*models.EventType, error)
	// GetEventTypeByID возвращает errs.ErrRecordNotFound, если записи нет.
	GetEventTypeByID(ctx context.Context, id uuid.UUID) (models.EventType, error)
	ListEventTypes(ctx context.Context, includeArchived bool) ([]models.EventType, error)
	UpdateEventType(ctx context.Context, e *models.EventType) error
	ArchiveEventType(ctx context.Context, id uuid.UUID) error
	// ActiveExposureTypeExists сообщает, есть ли неархивный тип показа.
	ActiveExposureTypeExists(ctx context.Context) (bool, error)
	// GetEventTypesByKeys возвращает найденные типы (в том числе архивные), ненайденные пропускает.
	GetEventTypesByKeys(ctx context.Context, keys []string) ([]models.EventType, error)
}

type MetricRepository interface {
	CreateMetric(ctx context.Context, m *models.Metric) (*models.Metric, error)
	GetMetricByID(ctx context.Context, id uuid.UUID) (models.Metric, error)
	GetMetricsByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Metric, error)
	ListMetrics(ctx context.Context, includeArchived bool) ([]models.Metric, error)
	UpdateMetricName(ctx context.Context, id uuid.UUID, name string) error
	ArchiveMetric(ctx context.Context, id uuid.UUID) error
}
