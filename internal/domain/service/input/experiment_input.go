package input

import "github.com/google/uuid"

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
