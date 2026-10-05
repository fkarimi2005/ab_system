package repository

import (
	"AB_system/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type ApprovalRepository interface {
	// RecordDecision сохраняет решение согласующего по версии эксперимента;
	// повторное решение того же согласующего перезаписывает прежнее.
	RecordDecision(ctx context.Context, a *models.ExperimentApproval) error
	ListApprovals(ctx context.Context, experimentID uuid.UUID) ([]models.ExperimentApproval, error)
	// CountApproved считает одобрения версии; approverIDs == nil — от любых согласующих.
	CountApproved(ctx context.Context, experimentID uuid.UUID, version int, approverIDs []uuid.UUID) (int, error)
}

type ApproverGroupRepository interface {
	// GetApproverGroup возвращает errs.ErrRecordNotFound, если группы нет.
	GetApproverGroup(ctx context.Context, experimenterID uuid.UUID) (models.ApproverGroup, error)
	SaveApproverGroup(ctx context.Context, g *models.ApproverGroup) error
}
