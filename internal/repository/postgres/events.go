package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DecisionRepository struct {
	db *gorm.DB
}

func NewDecisionRepository(db *gorm.DB) *DecisionRepository {
	return &DecisionRepository{db: db}
}

func (r *DecisionRepository) RecordDecisions(ctx context.Context, decisions []models.Decision) error {
	const op = "RecordDecisions"
	if len(decisions) == 0 {
		return nil
	}
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&decisions).Error
	return repository.CheckError(ctx, op, err)
}

func (r *DecisionRepository) GetDecisionsByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Decision, error) {
	const op = "GetDecisionsByIDs"
	var res []models.Decision
	if len(ids) == 0 {
		return res, nil
	}
	err := r.db.WithContext(ctx).Where("decision_id IN ?", ids).Find(&res).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *DecisionRepository) GetDecision(ctx context.Context, id uuid.UUID) (models.Decision, error) {
	const op = "GetDecision"
	var d models.Decision
	err := r.db.WithContext(ctx).First(&d, "decision_id = ?", id).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.Decision{}, err
	}
	return d, nil
}

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) InsertEventsIdempotent(ctx context.Context, events []models.Event) ([]bool, error) {
	const op = "InsertEventsIdempotent"
	inserted := make([]bool, len(events))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range events {
			res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&events[i])
			if res.Error != nil {
				return res.Error
			}
			inserted[i] = res.RowsAffected == 1
		}
		return nil
	})
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return inserted, nil
}

func (r *EventRepository) ListDecisionEvents(ctx context.Context, decisionID uuid.UUID) ([]models.EventRow, error) {
	const op = "ListDecisionEvents"
	var res []models.EventRow
	err := r.db.WithContext(ctx).
		Table("events AS e").
		Select("e.*, t.key AS type_key, t.requires_exposure, t.is_exposure").
		Joins("JOIN event_types t ON t.id = e.event_type_id").
		Where("e.decision_id = ?", decisionID).
		Order("e.event_ts ASC, e.event_id ASC").
		Scan(&res).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return res, nil
}
