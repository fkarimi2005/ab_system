package dto

import (
	"AB_system/internal/domain/models"

	"github.com/google/uuid"
)

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
		ExperimenterID: g.ExperimenterID,
		ApproverIDs:    g.MemberIDs(),
		MinApprovals:   g.MinApprovals,
	}
}
