package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/domain"
)

type UpdateRolePermissionsCommand struct {
	RoleID      uuid.UUID
	Permissions []domain.Permission
}

type UpdateRolePermissionsResult struct {
	Success bool `json:"success"`
}

type UpdateRolePermissionsHandler struct {
	roleRepo domain.RoleRepository
	rdb      *commonRedis.Client
}

func NewUpdateRolePermissionsHandler(roleRepo domain.RoleRepository, rdb *commonRedis.Client) *UpdateRolePermissionsHandler {
	return &UpdateRolePermissionsHandler{roleRepo: roleRepo, rdb: rdb}
}

func (h *UpdateRolePermissionsHandler) Handle(ctx context.Context, cmd UpdateRolePermissionsCommand) (*UpdateRolePermissionsResult, error) {
	role, err := h.roleRepo.GetByID(ctx, cmd.RoleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, nil
	}

	err = h.roleRepo.UpdatePermissions(ctx, cmd.RoleID, cmd.Permissions)
	if err != nil {
		return nil, err
	}

	// Xóa cache Redis để apply quyền mới ngay lập tức
	if h.rdb != nil {
		cacheKey := h.rdb.BuildKey(fmt.Sprintf("perm:%s", role.Name))
		h.rdb.Delete(ctx, cacheKey)
	}

	return &UpdateRolePermissionsResult{Success: true}, nil
}
