package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"github.com/google/uuid"
)

type RoleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}
func (s *RoleService) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.GetAllRoles(ctx)
}
func (s *RoleService) GetRoleByID(ctx context.Context, id uuid.UUID) (models.Role, error) {
	if id == uuid.Nil {
		return models.Role{}, errs.ErrIdIsEmpty
	}
	existingRole, err := s.roleRepo.ExistsRole(ctx, id)
	if err != nil {
		return models.Role{}, err
	}
	if !existingRole {
		return models.Role{}, errs.ErrRoleNotFound
	}
	return s.roleRepo.GetRoleByID(ctx, id)
}
func (s *RoleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errs.ErrIdIsEmpty
	}
	existingRole, err := s.roleRepo.ExistsRole(ctx, id)
	if err != nil {
		return err
	}
	if !existingRole {
		return errs.ErrRoleNotFound
	}
	return s.roleRepo.DeleteRole(ctx, id)
}
