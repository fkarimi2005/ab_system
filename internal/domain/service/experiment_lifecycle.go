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
	ActionSubmit   Action = "submit"
	ActionRework   Action = "rework"
	ActionStart    Action = "start"
	ActionPause    Action = "pause"
	ActionResume   Action = "resume"
	ActionComplete Action = "complete"

	ActionArchive Action = "archive"
)

type transitionRule struct {
	from []models.ExperimentStatus
	to   models.ExperimentStatus
}

func (r transitionRule) allows(s models.ExperimentStatus) bool {
	for _, f := range r.from {
		if f == s {
			return true
		}
	}
	return false
}

var transitionRules = map[Action]transitionRule{
	ActionSubmit:   {from: []models.ExperimentStatus{models.ExperimentStatusDraft}, to: models.ExperimentStatusReview},
	ActionRework:   {from: []models.ExperimentStatus{models.ExperimentStatusRejected}, to: models.ExperimentStatusDraft},
	ActionStart:    {from: []models.ExperimentStatus{models.ExperimentStatusApproved}, to: models.ExperimentStatusRunning},
	ActionPause:    {from: []models.ExperimentStatus{models.ExperimentStatusRunning}, to: models.ExperimentStatusPaused},
	ActionResume:   {from: []models.ExperimentStatus{models.ExperimentStatusPaused}, to: models.ExperimentStatusRunning},
	ActionComplete: {from: []models.ExperimentStatus{models.ExperimentStatusRunning, models.ExperimentStatusPaused}, to: models.ExperimentStatusCompleted},
	ActionArchive:  {from: []models.ExperimentStatus{models.ExperimentStatusCompleted, models.ExperimentStatusRejected}, to: models.ExperimentStatusArchived},
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
	if !rule.allows(e.Status) {
		return models.Experiment{}, errs.ErrInvalidTransition
	}

	if action == ActionSubmit {
		if err := s.validateForReview(ctx, e); err != nil {
			return models.Experiment{}, err
		}
	}

	if err := s.experimentRepo.TransitionStatus(ctx, e.ID, e.Status, rule.to); err != nil {
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
	if err := validation.ValidateVariantValues(flag.ValueType, e.Variants); err != nil {
		return err
	}
	if _, err := normalizeTargeting(e.Targeting); err != nil {
		return err
	}
	return s.checkConfigForReview(ctx, e)
}
