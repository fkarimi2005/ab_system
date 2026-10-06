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
	RecordDecision(
		ctx context.Context,
		experimentID uuid.UUID,
		approverID uuid.UUID,
		version int,
		status models.ApprovalStatus,
		comment string,
	) error
}

type ApprovalReader interface {
	GetApprovalsByExperimentID(
		ctx context.Context,
		experimentID uuid.UUID,
	) ([]models.ExperimentApproval, error)

	CountApproved(
		ctx context.Context,
		experimentID uuid.UUID,
		version int,
		approverIDs []uuid.UUID,
	) (int, error)
}
