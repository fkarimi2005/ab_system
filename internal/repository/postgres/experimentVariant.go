package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExperimentVariantRepository struct {
	db *gorm.DB
}

func NewExperimentVariantRepository(db *gorm.DB) ExperimentVariantRepository {
	return ExperimentVariantRepository{
		db: db,
	}
}
func (r *ExperimentVariantRepository) GetVariant(
	ctx context.Context,
	variantID uuid.UUID,
) (models.ExperimentVariant, error) {
	const op = "GetVariant"
	var result models.ExperimentVariant
	err := r.db.Where(&models.ExperimentVariant{
		ID: variantID,
	}).First(&result).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.ExperimentVariant{}, err
	}
	return result, nil
}
func (r *ExperimentVariantRepository) ListExperimentVariants(
	ctx context.Context,
	experimentId uuid.UUID,
) ([]models.ExperimentVariant, error) {
	const op = "ListExperimentVariants"
	var result []models.ExperimentVariant
	err := r.db.Where(&models.ExperimentVariant{},
		experimentId).Find(&result).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return []models.ExperimentVariant{}, err

	}
	return result, nil
}
