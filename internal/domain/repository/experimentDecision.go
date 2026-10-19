package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type ExperimentDecisionRepository interface {
	Save(
		ctx context.Context,
		decision models.ExperimentDecision,
	) error
	Exists(ctx context.Context,
		decisionID uuid.UUID) (bool, error)
	GetByID(ctx context.Context, decisionID uuid.UUID) (models.ExperimentDecision, error)
}
