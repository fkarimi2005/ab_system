package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExperimentApprovalRepository struct {
	db *gorm.DB
}

func NewExperimentApprovalRepository(db *gorm.DB) *ExperimentApprovalRepository {
	return &ExperimentApprovalRepository{
		db: db,
	}
}

func (r *ExperimentApprovalRepository) GetApprovalsByExperimentID(
	ctx context.Context,
	experimentID uuid.UUID,
) ([]models.ExperimentApproval, error) {
	const op = "GetApprovalsByExperimentID"
	var approvals []models.ExperimentApproval

	err := r.db.WithContext(ctx).
		Where("experiment_id = ?", experimentID).
		Order("created_at ASC").
		Find(&approvals).Error

	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}

	return approvals, nil
}

func (r *ExperimentApprovalRepository) ApproveExperiment(
	ctx context.Context,
	experimentID uuid.UUID,
	approverID uuid.UUID,
) error {
	const op = "ApproveExperiment"
	now := time.Now()

	approval := models.ExperimentApproval{
		ExperimentID: experimentID,
		ApproverID:   approverID,
		Status:       models.ApprovalApproved,
		DecidedAt:    &now,
	}

	result := r.db.WithContext(ctx).
		Create(&approval).Error
	if err := repository.CheckError(ctx, op, result); err != nil {
		return err
	}
	return nil
}

func (r *ExperimentApprovalRepository) RequestExperimentChanges(
	ctx context.Context,
	experimentID uuid.UUID,
	approverID uuid.UUID,
	comment string,
) error {
	const op = "RequestExperimentChanges"
	now := time.Now()

	approval := models.ExperimentApproval{
		ExperimentID: experimentID,
		ApproverID:   approverID,
		Status:       models.ApprovalChangesNeeded,
		Comment:      comment,
		DecidedAt:    &now,
	}

	result := r.db.WithContext(ctx).
		Create(&approval).Error
	if err := repository.CheckError(ctx, op, result); err != nil {
		return err
	}
	return nil
}

func (r *ExperimentApprovalRepository) RejectExperiment(
	ctx context.Context,
	experimentID uuid.UUID,
	approverID uuid.UUID,
	comment string,
) error {
	const op = "RejectExperiment"
	now := time.Now()

	approval := models.ExperimentApproval{
		ExperimentID: experimentID,
		ApproverID:   approverID,
		Status:       models.ApprovalRejected,
		Comment:      comment,
		DecidedAt:    &now,
	}

	result := r.db.WithContext(ctx).
		Create(&approval).Error
	if err := repository.CheckError(ctx, op, result); err != nil {
		return err
	}
	return nil
}
