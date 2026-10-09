package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"AB_system/pkg/errs"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventTypeRepository struct {
	db *gorm.DB
}

func NewEventTypeRepository(db *gorm.DB) *EventTypeRepository {
	return &EventTypeRepository{db: db}
}

func (r *EventTypeRepository) CreateEventType(ctx context.Context, e *models.EventType) (*models.EventType, error) {
	const op = "CreateEventType"
	if err := repository.CheckError(ctx, op, r.db.WithContext(ctx).Create(e).Error); err != nil {
		return nil, err
	}
	return e, nil
}

func (r *EventTypeRepository) GetEventTypeByID(ctx context.Context, id uuid.UUID) (models.EventType, error) {
	const op = "GetEventTypeByID"
	var e models.EventType
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.EventType{}, err
	}
	return e, nil
}

func (r *EventTypeRepository) ListEventTypes(ctx context.Context, includeArchived bool) ([]models.EventType, error) {
	const op = "ListEventTypes"
	var res []models.EventType
	q := r.db.WithContext(ctx).Order("created_at ASC")
	if !includeArchived {
		q = q.Where("archived_at IS NULL")
	}
	if err := repository.CheckError(ctx, op, q.Find(&res).Error); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *EventTypeRepository) UpdateEventType(ctx context.Context, e *models.EventType) error {
	const op = "UpdateEventType"
	res := r.db.WithContext(ctx).Model(&models.EventType{}).
		Where("id = ?", e.ID).
		Select("Name", "Description", "RequiresExposure").
		Updates(e)
	if err := repository.CheckError(ctx, op, res.Error); err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errs.ErrEventTypeNotFound
	}
	return nil
}

func (r *EventTypeRepository) ArchiveEventType(ctx context.Context, id uuid.UUID) error {
	const op = "ArchiveEventType"
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&models.EventType{}).
		Where("id = ? AND archived_at IS NULL", id).
		Update("archived_at", now)
	return repository.CheckError(ctx, op, res.Error)
}

func (r *EventTypeRepository) ActiveExposureTypeExists(ctx context.Context) (bool, error) {
	const op = "ActiveExposureTypeExists"
	var n int64
	err := r.db.WithContext(ctx).Model(&models.EventType{}).
		Where("is_exposure = ? AND archived_at IS NULL", true).Count(&n).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return false, err
	}
	return n > 0, nil
}

type MetricRepository struct {
	db *gorm.DB
}

func NewMetricRepository(db *gorm.DB) *MetricRepository {
	return &MetricRepository{db: db}
}

func (r *MetricRepository) CreateMetric(ctx context.Context, m *models.Metric) (*models.Metric, error) {
	const op = "CreateMetric"
	if err := repository.CheckError(ctx, op, r.db.WithContext(ctx).Create(m).Error); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MetricRepository) GetMetricByID(ctx context.Context, id uuid.UUID) (models.Metric, error) {
	const op = "GetMetricByID"
	var m models.Metric
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.Metric{}, err
	}
	return m, nil
}

func (r *MetricRepository) GetMetricsByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Metric, error) {
	const op = "GetMetricsByIDs"
	var res []models.Metric
	if len(ids) == 0 {
		return res, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *MetricRepository) ListMetrics(ctx context.Context, includeArchived bool) ([]models.Metric, error) {
	const op = "ListMetrics"
	var res []models.Metric
	q := r.db.WithContext(ctx).Order("created_at ASC")
	if !includeArchived {
		q = q.Where("archived_at IS NULL")
	}
	if err := repository.CheckError(ctx, op, q.Find(&res).Error); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *MetricRepository) UpdateMetricName(ctx context.Context, id uuid.UUID, name string) error {
	const op = "UpdateMetricName"
	res := r.db.WithContext(ctx).Model(&models.Metric{}).Where("id = ?", id).Update("name", name)
	if err := repository.CheckError(ctx, op, res.Error); err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errs.ErrMetricNotFound
	}
	return nil
}

func (r *MetricRepository) ArchiveMetric(ctx context.Context, id uuid.UUID) error {
	const op = "ArchiveMetric"
	res := r.db.WithContext(ctx).Model(&models.Metric{}).
		Where("id = ? AND archived_at IS NULL", id).
		Update("archived_at", time.Now())
	return repository.CheckError(ctx, op, res.Error)
}

func (r *EventTypeRepository) GetEventTypesByKeys(ctx context.Context, keys []string) ([]models.EventType, error) {
	const op = "GetEventTypesByKeys"
	var res []models.EventType
	if len(keys) == 0 {
		return res, nil
	}
	err := r.db.WithContext(ctx).Where("key IN ?", keys).Find(&res).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return res, nil
}
