package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

func (r *ApproverGroupRepository) SaveApproverGroup(
	ctx context.Context, g *models.ApproverGroup,
) error {
	const op = "SaveApproverGroup"
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cur models.ApproverGroup
		err := tx.First(&cur, "experimenter_id = ?", g.ExperimenterID).Error
		switch {
		case err == nil: // группа есть: меняем порог, старый состав удаляем
			g.ID = cur.ID
			if err := tx.Model(&cur).Update("min_approvals", g.MinApprovals).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.ApproverGroupMember{}, "group_id = ?", cur.ID).Error; err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound): // группы нет: создаём
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
