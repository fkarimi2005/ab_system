package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/internal/domain/service/input"
	"AB_system/internal/domain/service/validation"
	"AB_system/pkg/errs"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
)

type ExperimentService struct {
	experimentRepo repository.ExperimentRepository
	flagRepo       repository.FeatureFlagsRepository
}

func NewExperimentService(experimentRepo repository.ExperimentRepository, flagRepo repository.FeatureFlagsRepository) *ExperimentService {
	return &ExperimentService{experimentRepo: experimentRepo,
		flagRepo: flagRepo}
}
func (s *ExperimentService) CreateExperiment(
	ctx context.Context, ownerID uuid.UUID, in input.CreateExperimentInput,
) (*models.Experiment, error) {

	flag, err := s.flagRepo.GetFeatureFlagByID(ctx, in.FeatureFlagID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return nil, errs.ErrFeatureFlagNotFound
		}
		return nil, err
	}

	variants := input.ToVariantModels(in.Variants)

	if err := validation.ValidateConfig(in.Name, in.AudienceBP, variants); err != nil {
		return nil, err
	}
	if err := validation.ValidateVariantValues(flag.ValueType, variants); err != nil {
		return nil, err
	}

	return s.experimentRepo.CreateExperiment(ctx, &models.Experiment{
		FeatureFlagID: in.FeatureFlagID,
		Name:          in.Name,
		Status:        models.ExperimentStatusDraft, // задаёт сервер
		AudienceBP:    in.AudienceBP,
		Version:       1,       // задаёт сервер
		OwnerID:       ownerID, // из авторизации
		Variants:      variants,
	})
}
func (s *ExperimentService) UpdateExperiment(
	ctx context.Context, id uuid.UUID, actor models.Actor, in input.UpdateExperimentInput,
) error {
	cur, err := s.experimentRepo.GetExperimentByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return errs.ErrExperimentNotFound
		}
		return err
	}
	if cur.OwnerID != actor.ID && actor.Role != models.RoleAdmin {
		return errs.ErrPermissionDenied
	}
	if cur.Status != models.ExperimentStatusDraft {
		return errs.ErrExperimentNotEditable
	}

	flag, err := s.flagRepo.GetFeatureFlagByID(ctx, cur.FeatureFlagID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return errs.ErrFeatureFlagNotFound
		}
		return err
	}

	variants := input.ToVariantModels(in.Variants)
	if err := validation.ValidateConfig(in.Name, in.AudienceBP, variants); err != nil {
		return err
	}
	if err := validation.ValidateVariantValues(flag.ValueType, variants); err != nil {
		return err
	}

	oldVersion := cur.Version
	cur.Name = in.Name
	cur.AudienceBP = in.AudienceBP
	cur.Variants = variants
	cur.Version = oldVersion + 1

	snapshot, err := json.Marshal(cur)
	if err != nil {
		return err
	}
	return s.experimentRepo.UpdateExperiment(ctx, &cur, oldVersion, snapshot)
}

func (s *ExperimentService) GetExperimentByID(ctx context.Context, id uuid.UUID) (models.Experiment, error) {

	return s.experimentRepo.GetExperimentByID(ctx, id)
}
func (s *ExperimentService) GetExperiments(ctx context.Context) ([]models.Experiment, error) {
	return s.experimentRepo.GetAllExperiments(ctx)

}
