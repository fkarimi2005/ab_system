package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GuardrailTriggerRepository struct {
	db *gorm.DB
}

func NewGuardrailTriggerRepository(db *gorm.DB) *GuardrailTriggerRepository {
	return &GuardrailTriggerRepository{
		db: db,
	}
}
func (r *GuardrailTriggerRepository) SaveTrigger(
	ctx context.Context,
	trigger *models.GuardrailTrigger) error {
	const op = "SaveTrigger"
	err := r.db.WithContext(ctx).Create(trigger).Error
	if err = repository.CheckError(ctx, op, err); err != nil {
		return err
	}
	return nil
}
func (r *GuardrailTriggerRepository) Get(
	ctx context.Context,
	experimentID uuid.UUID,
) (models.GuardrailTrigger, error) {
	const op = "GetTrigger"
	var trigger models.GuardrailTrigger
	err := r.db.WithContext(ctx).
		Model(&models.GuardrailTrigger{}).
		Where("experiment_id = ?", experimentID).
		First(&trigger).Error
	if err = repository.CheckError(ctx, op, err); err != nil {
		return trigger, err
	}
	return trigger, nil
}
