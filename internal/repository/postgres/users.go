package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/http/observability"
	"AB_system/internal/repository"
	"AB_system/pkg/errs"
	"context"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	const op = "CreateUser"
	err := r.db.WithContext(ctx).Create(user).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "user created",
		"op", op,
		"trace_id", observability.GetTraceID(ctx),
		"user_id", user.ID,
	)
	return user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, u *models.User) error {
	const op = "UpdateUser"
	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", u.ID).
		Select("Name", "RoleID", "IsActive").
		Updates(u)
	if err := repository.CheckError(ctx, op, res.Error); err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errs.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) GetAllUser(ctx context.Context) ([]models.User, error) {
	const op = "GetAllUser"
	var users []models.User
	err := r.db.WithContext(ctx).Preload("Role").Find(&users).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	const op = "GetUserByID"
	var user models.User
	err := r.db.WithContext(ctx).Preload("Role").First(&user, "id = ?", id).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) ExistsUser(ctx context.Context, id uuid.UUID) (bool, error) {
	const op = "ExistsUser"
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Count(&count).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	const op = "ExistsUserByEmail"
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return false, err
	}
	return count > 0, nil
}
