package repository

import (
	"AB_system/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type UserReader interface {
	GetAllUser(
		ctx context.Context,
	) ([]models.User, error)
	GetUserByID(
		ctx context.Context,
		ID uuid.UUID,
	) (models.User, error)
	ExistsUser(
		ctx context.Context,
		ID uuid.UUID,
	) (bool, error)
	ExistsUserByEmail(
		ctx context.Context,
		email string,
	) (bool, error)
}
type UserWriter interface {
	CreateUser(
		ctx context.Context,
		user *models.User,
	) (*models.User, error)
	UpdateUser(
		ctx context.Context,
		user *models.User,
	) error
}

type UserRepository interface {
	UserReader
	UserWriter
}
