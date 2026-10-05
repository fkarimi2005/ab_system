package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"

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

func (r *ExperimentApprovalRepository) RecordDecision(
	ctx context.Context, a *models.ExperimentApproval,
) error {
	const op = "RecordDecision"
	err := r.db.WithContext(ctx).
		Omit(clause.Associations).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "experiment_id"}, {Name: "approver_id"}, {Name: "version"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"status", "comment", "decided_at", "updated_at"}),
		}).
		Create(a).Error
	return repository.CheckError(ctx, op, err)
}

func (r *ExperimentApprovalRepository) ListApprovals(
	ctx context.Context, experimentID uuid.UUID,
) ([]models.ExperimentApproval, error) {
	const op = "ListApprovals"
	var res []models.ExperimentApproval
	err := r.db.WithContext(ctx).
		Where("experiment_id = ?", experimentID).
		Order("version ASC, created_at ASC").
		Find(&res).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ExperimentApprovalRepository) CountApproved(
	ctx context.Context, experimentID uuid.UUID, version int, approverIDs []uuid.UUID,
) (int, error) {
	const op = "CountApproved"
	q := r.db.WithContext(ctx).Model(&models.ExperimentApproval{}).
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

type ApproverGroupRepository struct {
	db *gorm.DB
}

func NewApproverGroupRepository(db *gorm.DB) *ApproverGroupRepository {
	return &ApproverGroupRepository{db: db}
}

func (r *ApproverGroupRepository) GetApproverGroup(
	ctx context.Context, experimenterID uuid.UUID,
) (models.ApproverGroup, error) {
	const op = "GetApproverGroup"
	var g models.ApproverGroup
	err := r.db.WithContext(ctx).
		Preload("Members").
		First(&g, "experimenter_id = ?", experimenterID).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.ApproverGroup{}, err
	}
	return g, nil
}

// SaveApproverGroup создаёт группу или полностью заменяет состав и порог.
func (r *ApproverGroupRepository) SaveApproverGroup(ctx context.Context, g *models.ApproverGroup) error {
	const op = "SaveApproverGroup"
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cur models.ApproverGroup
		err := tx.First(&cur, "experimenter_id = ?", g.ExperimenterID).Error
		switch {
		case err == nil:
			g.ID = cur.ID
			if err := tx.Model(&cur).Update("min_approvals", g.MinApprovals).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.ApproverGroupMember{}, "group_id = ?", cur.ID).Error; err != nil {
				return err
			}
		case err == gorm.ErrRecordNotFound:
			if err := tx.Omit("Members").Create(g).Error; err != nil {
				return err
			}
		default:
			return err
		}
		members := make([]models.ApproverGroupMember, 0, len(g.Members))
		for _, m := range g.Members {
			members = append(members, models.ApproverGroupMember{GroupID: g.ID, ApproverID: m.ApproverID})
		}
		g.Members = members
		return tx.Create(&members).Error
	})
	return repository.CheckError(ctx, op, err)
}
