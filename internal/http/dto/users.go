package dto

import (
	"AB_system/internal/domain/models"
	"github.com/google/uuid"
	"time"
)

type UserRequest struct {
	Email    string    `json:"email" binding:"required,email"`
	Name     string    `json:"name" binding:"required,max=255"`
	RoleID   uuid.UUID `json:"role_id" binding:"required"`
	IsActive *bool     `json:"is_active"` // по умолчанию true
}
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	RoleID    uuid.UUID `json:"role_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
