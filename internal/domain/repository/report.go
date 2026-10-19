package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type ReportRepository interface {
	Save(
		ctx context.Context,
		report *models.Report,
	) error
	Get(
		ctx context.Context,
		experimentID uuid.UUID,
	) ([]models.Report, error)
	GetAll(
		ctx context.Context,
	) ([]models.Report, error)
}
