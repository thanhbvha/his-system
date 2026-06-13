package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Department struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DepartmentRepository interface {
	Create(ctx context.Context, department *Department) error
	GetByID(ctx context.Context, id uuid.UUID) (*Department, error)
	List(ctx context.Context, includeInactive bool) ([]*Department, error)
}
