package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"his-system/internal/identity/domain"
)

type RoleRepositoryPG struct {
	db *pgxpool.Pool
}

func NewRoleRepositoryPG(db *pgxpool.Pool) *RoleRepositoryPG {
	return &RoleRepositoryPG{db: db}
}

func (r *RoleRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return r.getBy(ctx, "id", id)
}

func (r *RoleRepositoryPG) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	return r.getBy(ctx, "name", name)
}

func (r *RoleRepositoryPG) getBy(ctx context.Context, field string, value interface{}) (*domain.Role, error) {
	q := `SELECT id, name FROM roles WHERE ` + field + ` = $1`
	var role domain.Role
	err := r.db.QueryRow(ctx, q, value).Scan(&role.ID, &role.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	perms, err := r.getPermissionsForRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.Permissions = perms
	return &role, nil
}

func (r *RoleRepositoryPG) getPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	q := `SELECT p.id, p.resource, p.action 
	      FROM permissions p
		  JOIN role_permissions rp ON p.id = rp.permission_id
		  WHERE rp.role_id = $1`

	rows, err := r.db.Query(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Resource, &p.Action); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *RoleRepositoryPG) List(ctx context.Context) ([]*domain.Role, error) {
	q := `SELECT id, name FROM roles`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	for _, role := range roles {
		perms, _ := r.getPermissionsForRole(ctx, role.ID)
		role.Permissions = perms
	}

	return roles, nil
}

func (r *RoleRepositoryPG) UpdatePermissions(ctx context.Context, roleID uuid.UUID, perms []domain.Permission) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	if err != nil {
		return err
	}

	for _, p := range perms {
		_, err = tx.Exec(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`, roleID, p.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
