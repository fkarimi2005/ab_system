package models

import (
	"time"

	"github.com/google/uuid"
)

type ApproverGroup struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ExperimenterID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"` // одна группа на экспериментатора
	MinApprovals   int       `gorm:"not null;check:chk_group_min_approvals,min_approvals >= 1"`

	Members []ApproverGroupMember `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type ApproverGroupMember struct {
	GroupID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	ApproverID uuid.UUID `gorm:"type:uuid;primaryKey"`
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
