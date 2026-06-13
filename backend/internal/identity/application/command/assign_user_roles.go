package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"his-system/internal/identity/domain"
)

type AssignUserRolesCommand struct {
	UserID  uuid.UUID
	RoleIDs []uuid.UUID
}

type AssignUserRolesResult struct {
	Success bool
}

type AssignUserRolesHandler struct {
	userRepo domain.UserRepository
}

func NewAssignUserRolesHandler(userRepo domain.UserRepository) *AssignUserRolesHandler {
	return &AssignUserRolesHandler{userRepo: userRepo}
}

func (h *AssignUserRolesHandler) Handle(ctx context.Context, cmd AssignUserRolesCommand) (*AssignUserRolesResult, error) {
	// Lấy user
	user, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // or return a Not Found error
	}

	// Update DB
	err = h.userRepo.UpdateRoles(ctx, user.ID, cmd.RoleIDs)
	if err != nil {
		return nil, err
	}

	// Update in-memory user entity
	user.RoleIDs = cmd.RoleIDs
	user.UpdatedAt = time.Now()
	// Optionally update user updated_at in DB by calling h.userRepo.Update(ctx, user)
	// But UpdateRoles just modifies the mapping table, which is fine.

	return &AssignUserRolesResult{Success: true}, nil
}
