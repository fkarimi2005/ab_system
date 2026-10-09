package repository

import (
	"AB_system/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type DecisionRepository interface {
	// RecordDecisions сохраняет решения; уже существующие (тот же decision_id) пропускает.
	RecordDecisions(ctx context.Context, decisions []models.Decision) error
	// GetDecisionsByIDs возвращает найденные решения, ненайденные пропускает.
	GetDecisionsByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Decision, error)
	// GetDecision возвращает errs.ErrRecordNotFound, если решения нет.
	GetDecision(ctx context.Context, id uuid.UUID) (models.Decision, error)
}

type EventRepository interface {
	// InsertEventsIdempotent вставляет события в одной транзакции; для каждого события
	// возвращает true, если оно новое, и false, если event_id уже был (дубль).
	InsertEventsIdempotent(ctx context.Context, events []models.Event) ([]bool, error)
	// ListDecisionEvents — события решения по времени вместе с настройками типов.
	ListDecisionEvents(ctx context.Context, decisionID uuid.UUID) ([]models.EventRow, error)
}
