package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type ExperimentReader interface {
	GetAllExperiments(
		ctx context.Context,
	) ([]models.Experiment, error)

	GetExperimentByID(
		ctx context.Context,
		experimentID uuid.UUID,
	) (models.Experiment, error)

	ExperimentExists(
		ctx context.Context,
		experimentID uuid.UUID,
	) (bool, error)
}

type ExperimentWriter interface {
	CreateExperiment(
		ctx context.Context,
		experiment *models.Experiment,
	) (*models.Experiment, error)

	UpdateExperiment(
		ctx context.Context,
		experiment *models.Experiment,
		expectedVersion int,
		snapshot []byte,
	) error
}

type ExperimentLifecycle interface {
	TransitionExperiment(
		ctx context.Context,
		experimentID uuid.UUID,
		status models.ExperimentStatus,
	) error

	CompleteExperiment(
		ctx context.Context,
		experimentID uuid.UUID,
	) error
}

type ExperimentRepository interface {
	ExperimentReader
	ExperimentWriter
	ExperimentLifecycle
}
