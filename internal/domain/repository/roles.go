package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type RoleReader interface {
	GetAllRoles(
		ctx context.Context,
	) ([]models.Role, error)
	GetRoleByID(
		ctx context.Context,
		roleID uuid.UUID,
	) (models.Role, error)
	ExistsRole(
		ctx context.Context,
		roleID uuid.UUID,
	) (bool, error)
}

type RoleDeleter interface {
	DeleteRole(
		ctx context.Context,
		roleID uuid.UUID,
	) error
}
type RoleRepository interface {
	RoleReader
	RoleDeleter
}
