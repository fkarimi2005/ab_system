package repository

import (
	"AB_system/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type ApproverGroupRepository interface {
	GetApproverGroup(ctx context.Context, experimenterID uuid.UUID) (models.ApproverGroup, error)
	SaveApproverGroup(ctx context.Context, g *models.ApproverGroup) error
}
