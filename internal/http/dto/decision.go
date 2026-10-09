package dto

import (
	"AB_system/internal/domain/service"

	"github.com/google/uuid"
)

type DecideRequest struct {
	SubjectID  string            `json:"subject_id" binding:"required,max=255"`
	Attributes map[string]string `json:"attributes"`
	FlagKeys   []string          `json:"flag_keys" binding:"required,min=1,max=100,dive,required,max=255"`
}

type DecideItemResponse struct {
	FlagKey      string     `json:"flag_key"`
	Value        *string    `json:"value"`
	ValueType    string     `json:"value_type,omitempty"`
	Source       string     `json:"source"`
	DecisionID   *uuid.UUID `json:"decision_id,omitempty"`
	ExperimentID *uuid.UUID `json:"experiment_id,omitempty"`
	Variant      string     `json:"variant,omitempty"`
}

type DecideResponse struct {
	Decisions []DecideItemResponse `json:"decisions"`
}

func NewDecideResponse(items []service.DecideItem) DecideResponse {
	res := DecideResponse{Decisions: make([]DecideItemResponse, 0, len(items))}
	for _, it := range items {
		res.Decisions = append(res.Decisions, DecideItemResponse{
			FlagKey: it.FlagKey, Value: it.Value, ValueType: it.ValueType, Source: it.Source,
			DecisionID: it.DecisionID, ExperimentID: it.ExperimentID, Variant: it.Variant,
		})
	}
	return res
}
