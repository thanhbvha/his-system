package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Department struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DepartmentRepository interface {
	Create(ctx context.Context, department *Department) error
	GetByID(ctx context.Context, id uuid.UUID) (*Department, error)
	List(ctx context.Context, includeInactive bool) ([]*Department, error)
}
