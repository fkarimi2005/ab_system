package models

import (
	"time"

	"github.com/google/uuid"
)

type ExperimentApproval struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ExperimentID uuid.UUID `json:"experiment_id" gorm:"type:uuid;not null;index;uniqueIndex:idx_approval_once"`
	ApproverID   uuid.UUID `json:"approver_id" gorm:"type:uuid;not null;index;uniqueIndex:idx_approval_once"`
	Version      int       `json:"version" gorm:"not null;default:1;uniqueIndex:idx_approval_once"`

	Status  ApprovalStatus `json:"status" gorm:"type:varchar(30);not null"`
	Comment string         `json:"comment" gorm:"type:text"`

	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DecidedAt *time.Time `json:"decided_at,omitempty"`

	Experiment *Experiment `json:"-" gorm:"foreignKey:ExperimentID"`
	Approver   *User       `json:"-" gorm:"foreignKey:ApproverID"`
}
type ApprovalStatus string

const (
	ApprovalPending       ApprovalStatus = "pending"
	ApprovalApproved      ApprovalStatus = "approved"
	ApprovalChangesNeeded ApprovalStatus = "changes_requested"
	ApprovalRejected      ApprovalStatus = "rejected"
)
