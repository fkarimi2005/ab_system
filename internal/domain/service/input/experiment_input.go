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
	Targeting     []byte // JSON-правило таргетинга; пусто или null — подходят все
	Metrics       []MetricRefInput
	Guardrails    []GuardrailInput
}
type UpdateExperimentInput struct {
	Name       string
	AudienceBP int
	Targeting  []byte // JSON-правило таргетинга; пусто или null — подходят все
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

// MetricRefInput — метрика эксперимента: из каталога и с ролью (target | diagnostic).
type MetricRefInput struct {
	MetricID uuid.UUID
	Role     string
}

// GuardrailInput — автоматический предохранитель эксперимента.
type GuardrailInput struct {
	MetricID      uuid.UUID
	Threshold     float64
	Comparison    string // gt | lt
	WindowSeconds int
	Action        string // pause | revert_to_control
}
