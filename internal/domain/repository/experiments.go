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
	// TransitionStatus атомарно меняет статус from -> to;
	// errs.ErrInvalidTransition, если текущий статус уже не from.
	TransitionStatus(
		ctx context.Context,
		experimentID uuid.UUID,
		from, to models.ExperimentStatus,
	) error
}

type ExperimentVersions interface {
	ListVersions(ctx context.Context, experimentID uuid.UUID) ([]models.ExperimentVersion, error)
	GetVersion(ctx context.Context, experimentID uuid.UUID, version int) (models.ExperimentVersion, error)
}

type ExperimentRepository interface {
	ExperimentReader
	ExperimentWriter
	ExperimentLifecycle
	ExperimentVersions
}
