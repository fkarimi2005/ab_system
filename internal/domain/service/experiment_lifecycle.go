package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service/validation"
	"AB_system/pkg/errs"
	"context"
	"errors"

	"github.com/google/uuid"
)

type Action string

const (
	ActionSubmit  Action = "submit"
	ActionRework  Action = "rework"
	ActionStart   Action = "start"
	ActionPause   Action = "pause"
	ActionResume  Action = "resume"
	ActionArchive Action = "archive"
)

type transitionRule struct {
	from, to models.ExperimentStatus
}

// Единственное место, где описаны допустимые переходы.
var transitionRules = map[Action]transitionRule{
	ActionSubmit:  {from: models.ExperimentStatusDraft, to: models.ExperimentStatusReview},
	ActionRework:  {from: models.ExperimentStatusRejected, to: models.ExperimentStatusDraft},
	ActionStart:   {from: models.ExperimentStatusApproved, to: models.ExperimentStatusRunning},
	ActionPause:   {from: models.ExperimentStatusRunning, to: models.ExperimentStatusPaused},
	ActionResume:  {from: models.ExperimentStatusPaused, to: models.ExperimentStatusRunning},
	ActionArchive: {from: models.ExperimentStatusCompleted, to: models.ExperimentStatusArchived},
}

func (s *ExperimentService) Transition(
	ctx context.Context, id uuid.UUID, actor models.Actor, action Action,
) (models.Experiment, error) {
	rule, ok := transitionRules[action]
	if !ok {
		return models.Experiment{}, errs.ErrInvalidTransition
	}

	e, err := s.experimentRepo.GetExperimentByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return models.Experiment{}, errs.ErrExperimentNotFound
		}
		return models.Experiment{}, err
	}
	if e.OwnerID != actor.ID && !actor.IsAdmin() {
		return models.Experiment{}, errs.ErrPermissionDenied
	}
	if e.Status != rule.from {
		return models.Experiment{}, errs.ErrInvalidTransition
	}

	// на ревью уходит только корректная конфигурация
	if action == ActionSubmit {
		if err := s.validateForReview(ctx, e); err != nil {
			return models.Experiment{}, err
		}
	}

	if err := s.experimentRepo.TransitionStatus(ctx, e.ID, rule.from, rule.to); err != nil {
		// сработал частичный уникальный индекс: на флаге уже есть активный эксперимент
		if action == ActionStart && errors.Is(err, errs.ErrDuplicateEntry) {
			return models.Experiment{}, errs.ErrActiveExperimentExists
		}
		return models.Experiment{}, err
	}

	e.Status = rule.to
	return e, nil
}

func (s *ExperimentService) validateForReview(ctx context.Context, e models.Experiment) error {
	flag, err := s.flagRepo.GetFeatureFlagByID(ctx, e.FeatureFlagID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return errs.ErrFeatureFlagNotFound
		}
		return err
	}
	if err := validation.ValidateConfig(e.Name, e.AudienceBP, e.Variants); err != nil {
		return err
	}
	return validation.ValidateVariantValues(flag.ValueType, e.Variants)
}
