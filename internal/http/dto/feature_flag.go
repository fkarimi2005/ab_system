package dto

import (
	"AB_system/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type CreateFeatureFlagRequest struct {
	Key          string `json:"key" binding:"required,max=255"`
	ValueType    string `json:"value_type" binding:"required,oneof=string number bool"`
	DefaultValue string `json:"default_value" binding:"required"`
}

type UpdateDefaultValueRequest struct {
	Value string `json:"value" binding:"required"`
}

type FeatureFlagResponse struct {
	ID           uuid.UUID `json:"id"`
	Key          string    `json:"key"`
	ValueType    string    `json:"value_type"`
	DefaultValue string    `json:"default_value"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewFeatureFlagResponse(f *models.FeatureFlag) FeatureFlagResponse {
	return FeatureFlagResponse{
		ID:           f.ID,
		Key:          f.Key,
		ValueType:    f.ValueType,
		DefaultValue: f.DefaultValue,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}
