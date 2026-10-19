package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}
func (r *EventRepository) CreateEvent(
	ctx context.Context,
	e models.Event,
) (models.Event, error) {
	const op = "CreateEvent"
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&e)

	if err := repository.CheckError(ctx, op, res.Error); err != nil {
		return models.Event{}, err
	}
	if res.RowsAffected == 1 {
		return e, nil
	}
	var existingEvent models.Event
	err := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("event_id=?", e.EventID).
		First(&existingEvent).Error
	if err = repository.CheckError(ctx, op, err); err != nil {
		return models.Event{}, err
	}
	return existingEvent, nil

}
func (r *EventRepository) GetEventByID(
	ctx context.Context,
	id uuid.UUID) (models.Event, error) {
	const op = "GetEventByID"
	var e models.Event
	err := r.db.WithContext(ctx).
		First(&e, "id = ?", id).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.Event{}, err
	}
	return e, nil
}
func (r *EventRepository) ListEvents(
	ctx context.Context) ([]models.Event, error) {
	const op = "ListEvents"
	var events []models.Event
	err := r.db.WithContext(ctx).Find(&events).Error
	if err := repository.
		CheckError(ctx, op, err); err != nil {
		return []models.Event{}, err

	}
	return events, nil
}
func (r *EventRepository) CountEventsByEventType(
	ctx context.Context,
	eventType string,
	experimentID uuid.UUID,
	variantID uuid.UUID,
) (int, error) {
	const op = "CountEventsByEventType"
	var count int64
	err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*)
     FROM events e
     JOIN experiment_decisions d ON d.id = e.decision_id
     WHERE e.event_type = ?
       AND e.experiment_id = ?
       AND d.variant_id = ?`,
		eventType,
		experimentID,
		variantID,
	).Scan(&count).Error

	if err := repository.CheckError(ctx, op, err); err != nil {
		return 0, err
	}
	return int(count), nil
}
func (r *EventRepository) CountEventsByEventTypeForGuard(
	ctx context.Context,
	eventType string,
	experimentID uuid.UUID,
	variantID uuid.UUID,
	from, to time.Time,
) (int, error) {
	const op = "CountEventsByEventTypeForGuard"

	var count int64

	err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*)
FROM events e
JOIN experiment_decisions d ON d.id = e.decision_id
WHERE e.event_type = ?
AND e.experiment_id = ?
AND d.variant_id = ?
AND e.created_at >= ?
AND e.created_at < ?`,
		eventType,
		experimentID,
		variantID,
		from,
		to,
	).Scan(&count).Error

	if err := repository.CheckError(ctx, op, err); err != nil {
		return 0, err
	}

	return int(count), nil
}
