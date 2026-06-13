package domain

import (
	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	ID       uuid.UUID `json:"id"`
	Resource string    `json:"resource"`
	Action   string    `json:"action"`
}

func (p Permission) String() string {
	return p.Resource + ":" + p.Action
}
