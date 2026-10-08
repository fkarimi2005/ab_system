package dto

import (
	"AB_system/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type DecisionRequest struct {
	Comment string `json:"comment" binding:"max=2000"`
}

type ApprovalResponse struct {
	ID         uuid.UUID  `json:"id"`
	ApproverID uuid.UUID  `json:"approver_id"`
	Version    int        `json:"version"`
	Status     string     `json:"status"`
	Comment    string     `json:"comment"`
	DecidedAt  *time.Time `json:"decided_at,omitempty"`
}

func NewApprovalResponse(a models.ExperimentApproval) ApprovalResponse {
	return ApprovalResponse{
		ID:         a.ID,
		ApproverID: a.ApproverID,
		Version:    a.Version,
		Status:     string(a.Status),
		Comment:    a.Comment,
		DecidedAt:  a.DecidedAt,
	}
}
