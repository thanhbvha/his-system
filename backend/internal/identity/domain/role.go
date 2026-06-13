package domain

import (
	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID
	Name        string
	Permissions []Permission
}

type Permission struct {
	ID       uuid.UUID
	Resource string
	Action   string
}

func (p Permission) String() string {
	return p.Resource + ":" + p.Action
}
