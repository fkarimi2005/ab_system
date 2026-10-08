package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExperimentApprovalRepository struct {
	db *gorm.DB
}

func NewExperimentApprovalRepository(db *gorm.DB) *ExperimentApprovalRepository {
	return &ExperimentApprovalRepository{db: db}
}

// RecordDecision сохраняет решение согласующего по конкретной версии эксперимента.
// Если этот человек уже решал по этой версии, решение перезаписывается.
func (r *ExperimentApprovalRepository) RecordDecision(
	ctx context.Context,
	experimentID uuid.UUID,
	approverID uuid.UUID,
	version int,
	status models.ApprovalStatus,
	comment string,
) error {
	const op = "RecordDecision"
	now := time.Now()

	approval := models.ExperimentApproval{
		ExperimentID: experimentID,
		ApproverID:   approverID,
		Version:      version,
		Status:       status,
		Comment:      comment,
		DecidedAt:    &now,
	}

	err := r.db.WithContext(ctx).
		Omit(clause.Associations).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "experiment_id"},
				{Name: "approver_id"},
				{Name: "version"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"status", "comment", "decided_at", "updated_at",
			}),
		}).
		Create(&approval).Error

	return repository.CheckError(ctx, op, err)
}

func (r *ExperimentApprovalRepository) GetApprovalsByExperimentID(
	ctx context.Context,
	experimentID uuid.UUID,
) ([]models.ExperimentApproval, error) {
	const op = "GetApprovalsByExperimentID"
	var approvals []models.ExperimentApproval

	err := r.db.WithContext(ctx).
		Where("experiment_id = ?", experimentID).
		Order("version ASC, created_at ASC").
		Find(&approvals).Error

	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return approvals, nil
}

// CountApproved считает одобрения данной версии.
// approverIDs == nil — считаем от любых согласующих (fallback без группы),
// иначе только от перечисленных (участники группы).
func (r *ExperimentApprovalRepository) CountApproved(
	ctx context.Context,
	experimentID uuid.UUID,
	version int,
	approverIDs []uuid.UUID,
) (int, error) {
	const op = "CountApproved"

	q := r.db.WithContext(ctx).
		Model(&models.ExperimentApproval{}).
		Where("experiment_id = ? AND version = ? AND status = ?",
			experimentID, version, models.ApprovalApproved)
	if approverIDs != nil {
		q = q.Where("approver_id IN ?", approverIDs)
	}

	var n int64
	if err := repository.CheckError(ctx, op, q.Count(&n).Error); err != nil {
		return 0, err
	}
	return int(n), nil
}
