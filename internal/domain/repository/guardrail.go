package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type GuardrailRepository interface {
	SaveTrigger(
		ctx context.Context,
		trigger *models.GuardrailTrigger) error
	Get(
		ctx context.Context,
		experimentID uuid.UUID,
		) (models.GuardrailTrigger, error)
}
