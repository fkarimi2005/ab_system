package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Role struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	RoleName  string         `json:"role_name" gorm:"unique;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	RoleAdmin        = "admin"
	RoleExperimenter = "experimenter"
	RoleApprover     = "approver"
	RoleViewer       = "viewer"
)

type User struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Name      string         `json:"name" gorm:"not null"`
	RoleID    uuid.UUID      `json:"role_id" gorm:"type:uuid;index"`
	Role      Role           `json:"role" gorm:"foreignKey:RoleID;references:ID"`
	IsActive  bool           `json:"is_active" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
