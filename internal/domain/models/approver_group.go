package models

import (
	"time"

	"github.com/google/uuid"
)

// ApproverGroup — группа согласующих конкретного экспериментатора.
// Эксперимент считается одобренным, когда набрано MinApprovals одобрений
// от участников группы. Если группы нет, см. ApprovalService (fallback).
type ApproverGroup struct {
	ID             uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ExperimenterID uuid.UUID `json:"experimenter_id" gorm:"type:uuid;not null;uniqueIndex"`
	MinApprovals   int       `json:"min_approvals" gorm:"not null;check:chk_group_min_approvals,min_approvals >= 1"`

	Members []ApproverGroupMember `json:"members" gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ApproverGroupMember struct {
	GroupID    uuid.UUID `json:"group_id" gorm:"type:uuid;primaryKey"`
	ApproverID uuid.UUID `json:"approver_id" gorm:"type:uuid;primaryKey"`
}

func (g ApproverGroup) MemberIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(g.Members))
	for _, m := range g.Members {
		ids = append(ids, m.ApproverID)
	}
	return ids
}

func (g ApproverGroup) HasMember(id uuid.UUID) bool {
	for _, m := range g.Members {
		if m.ApproverID == id {
			return true
		}
	}
	return false
}
