package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"his-system/internal/identity/domain"
)

type DeactivateUserCommand struct {
	UserID uuid.UUID
}

type DeactivateUserResult struct {
	Success bool
}

type DeactivateUserHandler struct {
	userRepo   domain.UserRepository
	deviceRepo domain.DeviceRepository
}

func NewDeactivateUserHandler(userRepo domain.UserRepository, deviceRepo domain.DeviceRepository) *DeactivateUserHandler {
	return &DeactivateUserHandler{userRepo: userRepo, deviceRepo: deviceRepo}
}

func (h *DeactivateUserHandler) Handle(ctx context.Context, cmd DeactivateUserCommand) (*DeactivateUserResult, error) {
	user, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // user not found
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	err = h.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	// Deactivate devices
	err = h.deviceRepo.DeactivateByUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Note: Refresh tokens in Redis should also be cleared.
	// We could publish an event here or directly access Redis if injected.

	return &DeactivateUserResult{Success: true}, nil
}
