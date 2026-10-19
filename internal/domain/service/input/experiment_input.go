package input

import (
	"AB_system/internal/domain/models"
	"github.com/google/uuid"
)

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
	Targeting     []byte
	Metrics       []MetricRefInput
	Guardrails    []GuardrailInput
}
type UpdateExperimentInput struct {
	Name       string
	AudienceBP int
	Targeting  []byte
	Metrics    []MetricRefInput
	Guardrails []GuardrailInput
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

type MetricRefInput struct {
	MetricID uuid.UUID
	Role     string
}

type GuardrailInput struct {
	MetricID      uuid.UUID
	Threshold     float64
	Comparison    string
	WindowSeconds int
	Action        string
}
