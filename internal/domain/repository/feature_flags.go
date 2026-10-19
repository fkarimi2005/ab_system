package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type FeatureFlagsReader interface {
	GetAllFeatureFlags(
		ctx context.Context,
	) ([]models.FeatureFlag, error)
	GetFeatureFlagByID(
		ctx context.Context,
		featureFlagID uuid.UUID,
	) (models.FeatureFlag, error)
	ExistsFeatureFlag(
		ctx context.Context,
		featureFlagID uuid.UUID,
	) (bool, error)
	GetFeatureFlagByKey(
		ctx context.Context,
		key string,
	) (models.FeatureFlag, error)
}
type FeatureFlagsWriter interface {
	CreateFeatureFlags(ctx context.Context, featureFlag *models.FeatureFlag) (*models.FeatureFlag, error)
	DeleteFeatureFlag(ctx context.Context, featureFlagID uuid.UUID) error
	UpdateFeatureFlagsDefaultValue(ctx context.Context, featureFlagID uuid.UUID, defaultValue string) error
}
type FeatureFlagsRepository interface {
	FeatureFlagsReader
	FeatureFlagsWriter
}
