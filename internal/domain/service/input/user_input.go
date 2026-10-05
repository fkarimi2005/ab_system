package input

import "github.com/google/uuid"

type UpdateUserInput struct {
	Name     *string
	RoleID   *uuid.UUID
	IsActive *bool
}
