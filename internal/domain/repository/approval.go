package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type ApprovalRepository interface {
	ApprovalReader
	ApprovalWriter
}
type ApprovalWriter interface {
	ApproveExperiment(
		ctx context.Context,
		experimentID uuid.UUID,
		approverID uuid.UUID,
	) error

	RequestExperimentChanges(
		ctx context.Context,
		experimentID uuid.UUID,
		approverID uuid.UUID,
		comment string,
	) error

	RejectExperiment(
		ctx context.Context,
		experimentID uuid.UUID,
		approverID uuid.UUID,
		comment string,
	) error
}
type ApprovalReader interface {
	GetApprovalsByExperimentID(
		ctx context.Context,
		experimentID uuid.UUID,
	) ([]models.ExperimentApproval, error)
}
