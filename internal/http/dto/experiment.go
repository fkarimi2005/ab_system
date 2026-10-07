package dto

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service/input"
	"github.com/google/uuid"
	"time"
)

type VariantRequest struct {
	Name      string `json:"name" binding:"required,max=255"`
	Value     string `json:"value" binding:"required"`
	WeightBP  int    `json:"weight_bp" binding:"required,gt=0,lte=10000"`
	IsControl bool   `json:"is_control"`
}

type CreateExperimentRequest struct {
	FeatureFlagID uuid.UUID        `json:"feature_flag_id" binding:"required"`
	Name          string           `json:"name" binding:"required,max=255"`
	AudienceBP    int              `json:"audience_bp" binding:"required,gt=0,lte=10000"`
	Variants      []VariantRequest `json:"variants" binding:"required,min=1,max=20,dive"`
}

func (r CreateExperimentRequest) ToInput() input.CreateExperimentInput {
	in := input.CreateExperimentInput{
		FeatureFlagID: r.FeatureFlagID,
		Name:          r.Name,
		AudienceBP:    r.AudienceBP,
	}
	for _, v := range r.Variants {
		in.Variants = append(in.Variants, input.VariantInput{
			Name: v.Name, Value: v.Value, WeightBP: v.WeightBP, IsControl: v.IsControl,
		})
	}
	return in
}

type VariantResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	WeightBP  int       `json:"weight_bp"`
	IsControl bool      `json:"is_control"`
}

type ExperimentResponse struct {
	ID            uuid.UUID         `json:"id"`
	FeatureFlagID uuid.UUID         `json:"feature_flag_id"`
	Name          string            `json:"name"`
	Status        string            `json:"status"`
	AudienceBP    int               `json:"audience_bp"`
	Version       int               `json:"version"`
	OwnerID       uuid.UUID         `json:"owner_id"`
	Variants      []VariantResponse `json:"variants"`
	CreatedAt     time.Time         `json:"created_at"`
}

func NewExperimentResponse(e *models.Experiment) ExperimentResponse {
	res := ExperimentResponse{
		ID: e.ID, FeatureFlagID: e.FeatureFlagID, Name: e.Name,
		Status: string(e.Status), AudienceBP: e.AudienceBP, Version: e.Version,
		OwnerID: e.OwnerID, CreatedAt: e.CreatedAt,
		Variants: make([]VariantResponse, 0, len(e.Variants)),
	}
	for _, v := range e.Variants {
		res.Variants = append(res.Variants, VariantResponse{
			ID: v.ID, Name: v.Name, Value: v.Value, WeightBP: v.Weight, IsControl: v.IsControl,
		})
	}
	return res
}
func (r UpdateExperimentRequest) ToInput() input.UpdateExperimentInput {
	in := input.UpdateExperimentInput{Name: r.Name, AudienceBP: r.AudienceBP}
	for _, v := range r.Variants {
		in.Variants = append(in.Variants, input.VariantInput{
			Name: v.Name, Value: v.Value, WeightBP: v.WeightBP, IsControl: v.IsControl,
		})
	}
	return in
}

type UpdateExperimentRequest struct {
	Name       string           `json:"name" binding:"required,max=255"`
	AudienceBP int              `json:"audience_bp" binding:"required,gt=0,lte=10000"`
	Variants   []VariantRequest `json:"variants" binding:"required,min=1,max=20,dive"`
}
