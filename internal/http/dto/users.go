package dto

import (
	"AB_system/internal/domain/models"
	"github.com/google/uuid"
	"time"
)

type UserRequest struct {
	Email    string    `json:"email" binding:"required"`
	Name     string    `json:"name" binding:"required"`
	RoleID   uuid.UUID `json:"role_id" binding:"required"`
	IsActive bool      `json:"is_active" default:"true"`
}
type UserResponse struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;index"`
	Email     string    `json:"email" gorm:"unique;index"`
	Name      string    `json:"name" gorm:"unique;index"`
	RoleID    uuid.UUID `json:"role_id" gorm:"index"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func NewUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		RoleID:    u.RoleID,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type UpdateUserRequest struct {
	Name     *string    `json:"name" binding:"omitempty,min=1,max=255"`
	RoleID   *uuid.UUID `json:"role_id"`
	IsActive *bool      `json:"is_active"`
}
