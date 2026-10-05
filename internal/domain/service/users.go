package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	service "AB_system/internal/domain/service/input"
	"AB_system/pkg/errs"
	"context"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) *UserService {
	return &UserService{userRepo: userRepo,
		roleRepo: roleRepo,
	}
}
func (s *UserService) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	if user.Email == "" {
		return nil, errs.ErrEmailIsEmpty
	}
	if user.Name == "" {
		return nil, errs.ErrNameIsEmpty
	}
	if user.RoleID == uuid.Nil {
		return nil, errs.ErrIdIsEmpty
	}
	exists, err := s.userRepo.ExistsUserByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errs.ErrEmailUniquenessFailed
	}
	return s.userRepo.CreateUser(ctx, &user)
}
func (s *UserService) UpdateUser(ctx context.Context, ID uuid.UUID, in service.UpdateUserInput) (*models.User, error) {
	exists, err := s.userRepo.ExistsUser(ctx, ID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}
	u, err := s.userRepo.GetUserByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		if *in.Name == "" {
			return nil, errs.ErrNameIsEmpty
		}
		u.Name = *in.Name
	}
	if in.RoleID != nil {
		if *in.RoleID == uuid.Nil {
			return nil, errs.ErrIdIsEmpty
		}
		exists, err := s.roleRepo.ExistsRole(ctx, *in.RoleID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errs.ErrRoleNotFound
		}
		u.RoleID = *in.RoleID

	}
	if in.IsActive != nil {
		u.IsActive = *in.IsActive
	}
	if err := s.userRepo.UpdateUser(ctx, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errs.ErrIdIsEmpty
	}
	exists, err := s.userRepo.ExistsUser(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrUserNotFound
	}
	return s.userRepo.DeleteUser(ctx, userID)
}
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (models.User, error) {
	if userID == uuid.Nil {
		return models.User{}, errs.ErrIdIsEmpty
	}
	return s.userRepo.GetUserByID(ctx, userID)
}
func (s *UserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return s.userRepo.GetAllUser(ctx)
}
