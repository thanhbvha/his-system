package query

import (
	"context"

	"github.com/google/uuid"

	"his-system/internal/identity/domain"
)

type GetRolePermissionsQuery struct {
	RoleID uuid.UUID
}

type GetRolePermissionsResult struct {
	Role *domain.Role
}

type GetRolePermissionsHandler struct {
	roleRepo domain.RoleRepository
}

func NewGetRolePermissionsHandler(roleRepo domain.RoleRepository) *GetRolePermissionsHandler {
	return &GetRolePermissionsHandler{roleRepo: roleRepo}
}

func (h *GetRolePermissionsHandler) Handle(ctx context.Context, q GetRolePermissionsQuery) (*GetRolePermissionsResult, error) {
	role, err := h.roleRepo.GetByID(ctx, q.RoleID)
	if err != nil {
		return nil, err
	}

	return &GetRolePermissionsResult{Role: role}, nil
}
