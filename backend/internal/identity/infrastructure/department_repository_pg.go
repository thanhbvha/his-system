package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"his-system/internal/identity/domain"
)

type DepartmentRepositoryPG struct {
	db *pgxpool.Pool
}

func NewDepartmentRepositoryPG(db *pgxpool.Pool) *DepartmentRepositoryPG {
	return &DepartmentRepositoryPG{db: db}
}

func (r *DepartmentRepositoryPG) Create(ctx context.Context, department *domain.Department) error {
	query := `
		INSERT INTO departments (id, code, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		department.ID, department.Code, department.Name, department.Description, department.IsActive,
		department.CreatedAt, department.UpdatedAt,
	)
	return err
}

func (r *DepartmentRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Department, error) {
	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM departments WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)

	var d domain.Department
	err := row.Scan(&d.ID, &d.Code, &d.Name, &d.Description, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DepartmentRepositoryPG) List(ctx context.Context, includeInactive bool) ([]*domain.Department, error) {
	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM departments
	`
	if !includeInactive {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var depts []*domain.Department
	for rows.Next() {
		var d domain.Department
		if err := rows.Scan(&d.ID, &d.Code, &d.Name, &d.Description, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		depts = append(depts, &d)
	}

	return depts, rows.Err()
}
