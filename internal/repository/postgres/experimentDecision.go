package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExperimentDecisionRepository struct {
	db *gorm.DB
}

func NewExperimentDecisionRepository(db *gorm.DB) *ExperimentDecisionRepository {
	return &ExperimentDecisionRepository{
		db: db,
	}
}
func (r *ExperimentDecisionRepository) Save(
	ctx context.Context,
	decision models.ExperimentDecision,
) error {
	const op = "ExperimentDecisionRepository.Save"
	err := r.db.WithContext(ctx).Create(&decision).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return err
	}
	return nil

}
func (r *ExperimentDecisionRepository) Exists(ctx context.Context, decisionID uuid.UUID) (bool, error) {
	const op = "ExperimentDecisionRepository.Exists"
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ExperimentDecision{}).
		Where(&models.ExperimentDecision{
			ID: decisionID,
		}).Create(&count).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return false, err
	}
	return count > 0, nil
}
func (r *ExperimentDecisionRepository) GetByID(ctx context.Context, decisionID uuid.UUID) (models.ExperimentDecision, error) {
	const op = "ExperimentDecisionRepository.GetByID"
	var decision models.ExperimentDecision
	err := r.db.WithContext(ctx).First(&decision, models.ExperimentDecision{
		ID: decisionID,
	}).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.ExperimentDecision{}, err
	}
	return decision, nil
}
