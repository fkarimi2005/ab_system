package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
)

// Пока групп согласующих нет: достаточно одного одобрения (fallback из ТЗ).

type ApprovalService struct {
	experimentRepo repository.ExperimentRepository
	approvalRepo   repository.ApprovalRepository
	groupRepo      repository.ApproverGroupRepository
}

func NewApprovalService(
	experimentRepo repository.ExperimentRepository,
	approvalRepo repository.ApprovalRepository,
	groupRepo repository.ApproverGroupRepository,
) *ApprovalService {
	return &ApprovalService{experimentRepo: experimentRepo, approvalRepo: approvalRepo, groupRepo: groupRepo}
}

// approvalRule: сколько одобрений нужно и чьи они считаются.
// members == nil — группы нет, подходит любой согласующий (fallback).
type approvalRule struct {
	min     int
	members []uuid.UUID
}

func (s *ApprovalService) ruleFor(ctx context.Context, ownerID uuid.UUID) (approvalRule, error) {
	g, err := s.groupRepo.GetApproverGroup(ctx, ownerID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return approvalRule{min: defaultMinApprovals}, nil
		}
		return approvalRule{}, err
	}
	return approvalRule{min: g.MinApprovals, members: g.MemberIDs()}, nil
}

const defaultMinApprovals = 1

// decide — общая часть всех трёх решений: проверки и запись решения.
func (s *ApprovalService) decide(
	ctx context.Context,
	experimentID uuid.UUID,
	actor models.Actor,
	status models.ApprovalStatus,
	comment string,
) (models.Experiment, approvalRule, error) {
	e, err := s.experimentRepo.GetExperimentByID(ctx, experimentID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return models.Experiment{}, approvalRule{}, errs.ErrExperimentNotFound
		}
		return models.Experiment{}, approvalRule{}, err
	}
	if e.Status != models.ExperimentStatusReview {
		return models.Experiment{}, approvalRule{}, errs.ErrInvalidTransition
	}
	if e.OwnerID == actor.ID {
		return models.Experiment{}, approvalRule{}, errs.ErrPermissionDenied
	}

	rule, err := s.ruleFor(ctx, e.OwnerID)
	if err != nil {
		return models.Experiment{}, approvalRule{}, err
	}
	// есть группа: решать могут только её участники
	if rule.members != nil && !slices.Contains(rule.members, actor.ID) {
		return models.Experiment{}, approvalRule{}, errs.ErrNotInApproverGroup
	}

	if err := s.approvalRepo.RecordDecision(ctx, e.ID, actor.ID, e.Version, status, comment); err != nil {
		return models.Experiment{}, approvalRule{}, err
	}
	return e, rule, nil
}

// Approve — одобрение. Когда одобрений набралось достаточно,
// эксперимент переходит review -> approved.
func (s *ApprovalService) Approve(
	ctx context.Context,
	experimentID uuid.UUID,
	actor models.Actor,
	comment string,
) error {

	e, rule, err := s.decide(ctx, experimentID, actor, models.ApprovalApproved, comment)
	if err != nil {
		return err
	}

	n, err := s.approvalRepo.CountApproved(ctx, e.ID, e.Version, rule.members)
	if err != nil {
		return err
	}
	if n < rule.min {
		return nil
	}

	err = s.experimentRepo.TransitionStatus(
		ctx, e.ID, models.ExperimentStatusReview, models.ExperimentStatusApproved,
	)
	if err != nil && !errors.Is(err, errs.ErrInvalidTransition) {
		return err
	}
	return nil
}

// Reject — отклонение. Эксперимент сразу переходит review -> rejected.
func (s *ApprovalService) Reject(
	ctx context.Context,
	experimentID uuid.UUID,
	actor models.Actor,
	comment string,
) error {
	if comment == "" {
		return errs.ErrCommentRequired
	}
	e, _, err := s.decide(ctx, experimentID, actor, models.ApprovalRejected, comment)
	if err != nil {
		return err
	}
	return s.moveFromReview(ctx, e.ID, models.ExperimentStatusRejected)
}

// RequestChanges — запрос правок. Эксперимент возвращается review -> draft.
func (s *ApprovalService) RequestChanges(
	ctx context.Context,
	experimentID uuid.UUID,
	actor models.Actor,
	comment string,
) error {
	if comment == "" {
		return errs.ErrCommentRequired
	}
	e, _, err := s.decide(ctx, experimentID, actor, models.ApprovalChangesNeeded, comment)
	if err != nil {
		return err
	}
	return s.moveFromReview(ctx, e.ID, models.ExperimentStatusDraft)
}

// moveFromReview меняет статус из review. Если другой согласующий уже успел
// сменить статус (гонка), это не считаем ошибкой.
func (s *ApprovalService) moveFromReview(
	ctx context.Context,
	experimentID uuid.UUID,
	to models.ExperimentStatus,
) error {
	err := s.experimentRepo.TransitionStatus(
		ctx, experimentID, models.ExperimentStatusReview, to,
	)
	if err != nil && !errors.Is(err, errs.ErrInvalidTransition) {
		return err
	}
	return nil
}
func (s *ApprovalService) List(
	ctx context.Context, experimentID uuid.UUID,
) ([]models.ExperimentApproval, error) {
	exists, err := s.experimentRepo.ExperimentExists(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrExperimentNotFound
	}
	return s.approvalRepo.GetApprovalsByExperimentID(ctx, experimentID)
}
