package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type ExperimentVariantRepository interface {
	GetVariant(
		ctx context.Context,
		variantID uuid.UUID,
	) (models.ExperimentVariant, error)
	ListExperimentVariants(
		ctx context.Context,
		experimentId uuid.UUID,
	) ([]models.ExperimentVariant, error)
}
