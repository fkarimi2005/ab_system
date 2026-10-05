package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// manualTransitions — переходы, которые владелец/админ делает вручную.
// Переходы в approved/rejected и review -> draft (по запросу правок)
// выполняются только через решения согласующих.
var manualTransitions = map[models.ExperimentStatus][]models.ExperimentStatus{
	models.ExperimentStatusDraft:     {models.ExperimentStatusReview, models.ExperimentStatusArchived},
	models.ExperimentStatusReview:    {models.ExperimentStatusDraft},
	models.ExperimentStatusApproved:  {models.ExperimentStatusRunning},
	models.ExperimentStatusRunning:   {models.ExperimentStatusPaused, models.ExperimentStatusCompleted},
	models.ExperimentStatusPaused:    {models.ExperimentStatusRunning, models.ExperimentStatusCompleted},
	models.ExperimentStatusCompleted: {models.ExperimentStatusArchived},
	models.ExperimentStatusRejected:  {models.ExperimentStatusArchived},
}

type Decision string

const (
	DecisionApprove        Decision = "approve"
	DecisionReject         Decision = "reject"
	DecisionRequestChanges Decision = "request_changes"
)

// DefaultMinApprovals — порог, когда у владельца нет группы согласующих (fallback):
// достаточно одного одобрения любого approver/admin, кроме самого владельца.
const DefaultMinApprovals = 1

type LifecycleService struct {
	experiments repository.ExperimentRepository
	approvals   repository.ApprovalRepository
	groups      repository.ApproverGroupRepository
}

func NewLifecycleService(
	e repository.ExperimentRepository,
	a repository.ApprovalRepository,
	g repository.ApproverGroupRepository,
) *LifecycleService {
	return &LifecycleService{experiments: e, approvals: a, groups: g}
}

func (s *LifecycleService) load(ctx context.Context, id uuid.UUID) (models.Experiment, error) {
	e, err := s.experiments.GetExperimentByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return models.Experiment{}, errs.ErrExperimentNotFound
		}
		return models.Experiment{}, err
	}
	return e, nil
}

func canTransition(from, to models.ExperimentStatus) bool {
	for _, t := range manualTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// Transition — ручной переход статуса (владелец или admin).
func (s *LifecycleService) Transition(
	ctx context.Context, id uuid.UUID, actor models.Actor, to models.ExperimentStatus,
) (models.Experiment, error) {
	e, err := s.load(ctx, id)
	if err != nil {
		return models.Experiment{}, err
	}
	if e.OwnerID != actor.ID && !actor.IsAdmin() {
		return models.Experiment{}, errs.ErrPermissionDenied
	}
	if !canTransition(e.Status, to) {
		return models.Experiment{}, errs.ErrInvalidTransition
	}
	if err := s.experiments.TransitionStatus(ctx, id, e.Status, to); err != nil {
		if errors.Is(err, errs.ErrDuplicateEntry) && to == models.ExperimentStatusRunning {
			return models.Experiment{}, errs.ErrActiveExperimentExists
		}
		return models.Experiment{}, err
	}
	return s.load(ctx, id)
}

// Decide — решение согласующего по эксперименту в статусе review.
func (s *LifecycleService) Decide(
	ctx context.Context, id uuid.UUID, actor models.Actor, d Decision, comment string,
) (models.Experiment, error) {
	e, err := s.load(ctx, id)
	if err != nil {
		return models.Experiment{}, err
	}
	if e.Status != models.ExperimentStatusReview {
		return models.Experiment{}, errs.ErrInvalidTransition
	}
	if e.OwnerID == actor.ID {
		return models.Experiment{}, errs.ErrSelfApproval
	}
	if d != DecisionApprove && comment == "" {
		return models.Experiment{}, errs.ErrCommentRequired
	}

	required, members, err := s.threshold(ctx, e.OwnerID)
	if err != nil {
		return models.Experiment{}, err
	}
	if members != nil && !containsID(members, actor.ID) {
		return models.Experiment{}, errs.ErrNotInApproverGroup
	}

	status := map[Decision]models.ApprovalStatus{
		DecisionApprove:        models.ApprovalApproved,
		DecisionReject:         models.ApprovalRejected,
		DecisionRequestChanges: models.ApprovalChangesNeeded,
	}[d]
	if status == "" {
		return models.Experiment{}, errs.ErrInvalidField
	}
	now := time.Now()
	if err := s.approvals.RecordDecision(ctx, &models.ExperimentApproval{
		ExperimentID: e.ID,
		ApproverID:   actor.ID,
		Version:      e.Version,
		Status:       status,
		Comment:      comment,
		DecidedAt:    &now,
	}); err != nil {
		return models.Experiment{}, err
	}

	var to models.ExperimentStatus
	switch d {
	case DecisionReject:
		to = models.ExperimentStatusRejected
	case DecisionRequestChanges:
		to = models.ExperimentStatusDraft
	case DecisionApprove:
		n, err := s.approvals.CountApproved(ctx, e.ID, e.Version, members)
		if err != nil {
			return models.Experiment{}, err
		}
		if n >= required {
			to = models.ExperimentStatusApproved
		}
	}
	if to != "" {
		err := s.experiments.TransitionStatus(ctx, e.ID, models.ExperimentStatusReview, to)
		// гонка с другим согласующим: статус уже сменился — не ошибка
		if err != nil && !errors.Is(err, errs.ErrInvalidTransition) {
			return models.Experiment{}, err
		}
	}
	return s.load(ctx, e.ID)
}

// threshold возвращает порог одобрений и состав группы владельца
// (members == nil — группы нет, работает fallback).
func (s *LifecycleService) threshold(ctx context.Context, ownerID uuid.UUID) (int, []uuid.UUID, error) {
	g, err := s.groups.GetApproverGroup(ctx, ownerID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return DefaultMinApprovals, nil, nil
		}
		return 0, nil, err
	}
	return g.MinApprovals, g.MemberIDs(), nil
}

func (s *LifecycleService) Approvals(ctx context.Context, id uuid.UUID) ([]models.ExperimentApproval, error) {
	if _, err := s.load(ctx, id); err != nil {
		return nil, err
	}
	return s.approvals.ListApprovals(ctx, id)
}

func containsID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
