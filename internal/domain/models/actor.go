package models

import "github.com/google/uuid"

type Actor struct {
	ID   uuid.UUID
	Role string
}

func (a Actor) IsAdmin() bool { return a.Role == RoleAdmin }
