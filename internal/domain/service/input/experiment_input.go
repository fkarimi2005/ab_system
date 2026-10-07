package input

import (
	"AB_system/internal/domain/models"
	"github.com/google/uuid"
)

// internal/domain/service/experiment_input.go
type VariantInput struct {
	Name      string
	Value     string
	WeightBP  int
	IsControl bool
}

type CreateExperimentInput struct {
	FeatureFlagID uuid.UUID
	Name          string
	AudienceBP    int
	Variants      []VariantInput
}
type UpdateExperimentInput struct {
	Name       string
	AudienceBP int
	Variants   []VariantInput
}

func ToVariantModels(in []VariantInput) []models.ExperimentVariant {
	res := make([]models.ExperimentVariant, 0, len(in))
	for _, v := range in {
		res = append(res, models.ExperimentVariant{
			Name: v.Name, Value: v.Value, Weight: v.WeightBP, IsControl: v.IsControl,
		})
	}
	return res
}
