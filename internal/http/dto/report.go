package dto

import "github.com/google/uuid"

type ReportResponse struct {
	ExperimentID uuid.UUID `json:"experiment_id"`

}
