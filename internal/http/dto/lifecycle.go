package dto

import (
	"AB_system/internal/domain/models"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TransitionRequest struct {
	Status string `json:"status" binding:"required"`
}

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
		ID: a.ID, ApproverID: a.ApproverID, Version: a.Version,
		Status: string(a.Status), Comment: a.Comment, DecidedAt: a.DecidedAt,
	}
}

type VersionResponse struct {
	Version   int             `json:"version"`
	Snapshot  json.RawMessage `json:"snapshot"`
	CreatedAt time.Time       `json:"created_at"`
}

func NewVersionResponse(v models.ExperimentVersion) VersionResponse {
	return VersionResponse{Version: v.Version, Snapshot: json.RawMessage(v.Snapshot), CreatedAt: v.CreatedAt}
}

type SetApproverGroupRequest struct {
	ApproverIDs  []uuid.UUID `json:"approver_ids" binding:"required,min=1"`
	MinApprovals int         `json:"min_approvals" binding:"required,gte=1"`
}

type ApproverGroupResponse struct {
	ExperimenterID uuid.UUID   `json:"experimenter_id"`
	ApproverIDs    []uuid.UUID `json:"approver_ids"`
	MinApprovals   int         `json:"min_approvals"`
}

func NewApproverGroupResponse(g models.ApproverGroup) ApproverGroupResponse {
	return ApproverGroupResponse{
		ExperimenterID: g.ExperimenterID, ApproverIDs: g.MemberIDs(), MinApprovals: g.MinApprovals,
	}
}
