package query

import (
	"context"

	"his-system/internal/identity/domain"
)

type ListPermissionsQuery struct {
}

type ListPermissionsResult struct {
	Permissions []domain.Permission `json:"permissions"`
}

type ListPermissionsHandler struct {
	roleRepo domain.RoleRepository
}

func NewListPermissionsHandler(roleRepo domain.RoleRepository) *ListPermissionsHandler {
	return &ListPermissionsHandler{roleRepo: roleRepo}
}

func (h *ListPermissionsHandler) Handle(ctx context.Context, q ListPermissionsQuery) (*ListPermissionsResult, error) {
	perms, err := h.roleRepo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	return &ListPermissionsResult{Permissions: perms}, nil
}
