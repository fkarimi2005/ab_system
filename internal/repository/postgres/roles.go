package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}
func (r *RoleRepository) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	const op = "GetAllRoles"
	var roles []models.Role
	result := r.db.WithContext(ctx).Find(&roles)
	if err := repository.CheckError(ctx, op, result.Error); err != nil {
		return nil, err
	}
	return roles, nil
}
func (r *RoleRepository) GetRoleByID(ctx context.Context, roleID uuid.UUID) (models.Role, error) {
	const op = "GetRoleByID"
	var role models.Role
	result := r.db.WithContext(ctx).First(&role, "id = ?", roleID)
	if err := repository.CheckError(ctx, op, result.Error); err != nil {
		return models.Role{}, err
	}
	return role, nil
}
func (r *RoleRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	const op = "DeleteRole"
	var role models.Role
	result := r.db.WithContext(ctx).Delete(&role, "id = ?", roleID)
	if err := repository.CheckError(ctx, op, result.Error); err != nil {
		return err
	}
	return nil
}
func (r *RoleRepository) ExistsRole(ctx context.Context, roleID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.Role{}).Where("id = ?", roleID).Count(&count)
	if result.Error != nil {
		return false, repository.CheckError(ctx, `role`, result.Error)
	}
	return count > 0, nil
}
