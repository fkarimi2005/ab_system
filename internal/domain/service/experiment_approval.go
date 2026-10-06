package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"errors"

	"github.com/google/uuid"
)

// Пока групп согласующих нет: достаточно одного одобрения (fallback из ТЗ).
const defaultMinApprovals = 1

type ApprovalService struct {
	experimentRepo repository.ExperimentRepository
	approvalRepo   repository.ApprovalRepository
}

func NewApprovalService(
	experimentRepo repository.ExperimentRepository,
	approvalRepo repository.ApprovalRepository,
) *ApprovalService {
	return &ApprovalService{experimentRepo: experimentRepo, approvalRepo: approvalRepo}
}

// decide — общая часть всех трёх решений: проверки и запись решения.
func (s *ApprovalService) decide(
	ctx context.Context,
	experimentID uuid.UUID,
	actor models.Actor,
	status models.ApprovalStatus,
	comment string,
) (models.Experiment, error) {
	e, err := s.experimentRepo.GetExperimentByID(ctx, experimentID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return models.Experiment{}, errs.ErrExperimentNotFound
		}
		return models.Experiment{}, err
	}
	if e.Status != models.ExperimentStatusReview {
		return models.Experiment{}, errs.ErrInvalidTransition
	}
	if e.OwnerID == actor.ID {
		return models.Experiment{}, errs.ErrPermissionDenied // нельзя согласовывать своё
	}
	if err := s.approvalRepo.RecordDecision(
		ctx, e.ID, actor.ID, e.Version, status, comment,
	); err != nil {
		return models.Experiment{}, err
	}
	return e, nil
}

// Approve — одобрение. Когда одобрений набралось достаточно,
// эксперимент переходит review -> approved.
func (s *ApprovalService) Approve(
	ctx context.Context,
	experimentID uuid.UUID,
	actor models.Actor,
	comment string,
) error {
	e, err := s.decide(ctx, experimentID, actor, models.ApprovalApproved, comment)
	if err != nil {
		return err
	}
	if comment == "" {
		return errs.ErrCommentRequired
	}

	n, err := s.approvalRepo.CountApproved(ctx, e.ID, e.Version, nil)
	if err != nil {
		return err
	}
	if n < defaultMinApprovals {
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
	e, err := s.decide(ctx, experimentID, actor, models.ApprovalRejected, comment)
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
	e, err := s.decide(ctx, experimentID, actor, models.ApprovalChangesNeeded, comment)
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
