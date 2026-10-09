package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeatureFlagRepository struct {
	db *gorm.DB
}

func NewFeatureFlagRepository(db *gorm.DB) *FeatureFlagRepository {
	return &FeatureFlagRepository{
		db: db,
	}
}
func (r *FeatureFlagRepository) GetAllFeatureFlags(
	ctx context.Context,
) ([]models.FeatureFlag, error) {
	const op = "GetAllFeatureFlags"
	var flags []models.FeatureFlag
	err := r.db.WithContext(ctx).Find(&flags).Error
	if err = repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return flags, err

}
func (r *FeatureFlagRepository) GetFeatureFlagByID(
	ctx context.Context,
	featureFlagID uuid.UUID,
) (models.FeatureFlag, error) {
	const op = "GetFeatureFlagByID"
	var flag models.FeatureFlag
	err := r.db.WithContext(ctx).Where("id=?", featureFlagID).First(&flag).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return flag, err
	}
	return flag, nil
}
func (r *FeatureFlagRepository) CreateFeatureFlags(
	ctx context.Context,
	featureFlag *models.FeatureFlag,
) (*models.FeatureFlag, error) {
	const op = "CreateFeatureFlag"
	result := r.db.WithContext(ctx).Create(featureFlag)
	if err := repository.CheckError(ctx, op, result.Error); err != nil {
		return nil, err
	}
	return featureFlag, nil
}

func (r *FeatureFlagRepository) ExistsFeatureFlag(
	ctx context.Context,
	featureFlagID uuid.UUID,
) (bool, error) {
	const op = "ExistsFeatureFlag"
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.FeatureFlag{}).
		Where("id=?", featureFlagID).
		Count(&count).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return false, err
	}
	return count > 0, nil
}
func (r *FeatureFlagRepository) DeleteFeatureFlag(
	ctx context.Context,
	featureFlagID uuid.UUID,
) error {
	const op = "DeleteFeatureFlag"
	err := r.db.WithContext(ctx).Where("id=?", featureFlagID).Delete(&models.FeatureFlag{}).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return err
	}
	return nil
}
func (r *FeatureFlagRepository) UpdateFeatureFlagsDefaultValue(
	ctx context.Context,
	featureFlagID uuid.UUID,
	defaultValue string,
) error {
	const op = "UpdateFeatureFlagsDefaultValue"
	var flag models.FeatureFlag
	err := r.db.WithContext(ctx).Model(&flag).Where("id=?", featureFlagID).Update("default_value", defaultValue).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return err
	}
	return nil

}
func (r *FeatureFlagRepository) GetFeatureFlagByKey(
	ctx context.Context,
	key string,
) (models.FeatureFlag, error) {
	const op = "GetFeatureFlagByKey"
	var flag models.FeatureFlag
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&flag).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.FeatureFlag{}, err
	}
	return flag, nil
}
