package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/internal/domain/service/validation"
	"AB_system/pkg/errs"
	"context"
	"errors"
	"github.com/google/uuid"
)

type FeatureFlagService struct {
	featureFlagRepo repository.FeatureFlagsRepository
}

func NewFeatureFlagService(featureFlagRepo repository.FeatureFlagsRepository) *FeatureFlagService {
	return &FeatureFlagService{featureFlagRepo: featureFlagRepo}
}
func (s *FeatureFlagService) CreateFeatureFlag(
	ctx context.Context,
	f *models.FeatureFlag,
) (*models.FeatureFlag, error) {
	if f.Key == "" {
		return nil, errs.ErrKeyIsEmpty
	}
	if f.ValueType == "" {
		return nil, errs.ErrValueTypeIsEmpty
	}
	if f.DefaultValue == "" {
		return nil, errs.ErrDefaultValueIsEmpty
	}
	// проверяет и сам тип, и то, что значение ему соответствует
	if err := validation.ValidateValue(f.ValueType, f.DefaultValue); err != nil {
		return nil, err
	}
	return s.featureFlagRepo.CreateFeatureFlags(ctx, f)
}
func (s *FeatureFlagService) DeleteFeatureFlag(ctx context.Context,
	featureFlagID uuid.UUID) error {
	if featureFlagID == uuid.Nil {
		return errs.ErrIdIsEmpty
	}
	exist, err := s.featureFlagRepo.ExistsFeatureFlag(ctx, featureFlagID)
	if err != nil {
		return err
	}
	if !exist {
		return errs.ErrFeatureFlagNotFound
	}
	return s.featureFlagRepo.DeleteFeatureFlag(ctx, featureFlagID)
}
func (s *FeatureFlagService) GetFeatureFlag(ctx context.Context,
	featureFlagID uuid.UUID,
) (models.FeatureFlag, error) {
	if featureFlagID == uuid.Nil {
		return models.FeatureFlag{}, errs.ErrIdIsEmpty
	}
	exist, err := s.featureFlagRepo.ExistsFeatureFlag(ctx, featureFlagID)
	if err != nil {
		return models.FeatureFlag{}, err
	}
	if !exist {
		return models.FeatureFlag{}, errs.ErrFeatureFlagNotFound
	}
	return s.featureFlagRepo.GetFeatureFlagByID(ctx, featureFlagID)
}
func (s *FeatureFlagService) GetFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error) {
	return s.featureFlagRepo.GetAllFeatureFlags(ctx)
}

func (s *FeatureFlagService) UpdateDefaultValue(ctx context.Context, id uuid.UUID, v string) error {
	if v == "" {
		return errs.ErrDefaultValueIsEmpty
	}
	flag, err := s.featureFlagRepo.GetFeatureFlagByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return errs.ErrFeatureFlagNotFound
		}
		return err
	}
	if err := validation.ValidateValue(flag.ValueType, v); err != nil {
		return err
	}
	return s.featureFlagRepo.UpdateFeatureFlagsDefaultValue(ctx, id, v)
}
