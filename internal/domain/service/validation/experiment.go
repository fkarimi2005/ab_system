package validation

import (
	"AB_system/internal/domain/models"
	"AB_system/pkg/errs"
)

func ValidateConfig(name string, audienceBP int, variants []models.ExperimentVariant) error {
	if name == "" {
		return errs.ErrExperimentNameIsEmpty
	}
	if audienceBP <= 0 || audienceBP > models.FullAudienceBP {
		return errs.ErrInvalidField
	}
	if len(variants) == 0 {
		return errs.ErrExperimentVariantsEmpty
	}

	sum, controls := 0, 0
	seen := map[string]bool{}
	for _, v := range variants {
		if v.Name == "" || v.Value == "" || v.Weight <= 0 || seen[v.Name] {
			return errs.ErrInvalidField
		}
		seen[v.Name] = true
		sum += v.Weight
		if v.IsControl {
			controls++
		}
	}
	if sum != audienceBP {
		return errs.ErrWeightsSumMismatch
	}
	if controls != 1 {
		return errs.ErrControlVariantCount
	}
	return nil
}
