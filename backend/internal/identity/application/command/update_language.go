package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"his-system/internal/identity/domain"
	appErrors "his-system/pkg/errors"
)

type UpdateLanguageCommand struct {
	UserID   uuid.UUID
	Language string
}

type UpdateLanguageHandler struct {
	userRepo domain.UserRepository
}

func NewUpdateLanguageHandler(userRepo domain.UserRepository) *UpdateLanguageHandler {
	return &UpdateLanguageHandler{userRepo: userRepo}
}

func (h *UpdateLanguageHandler) Handle(ctx context.Context, cmd UpdateLanguageCommand) error {
	user, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return appErrors.ErrNotFound
	}

	user.PreferredLanguage = cmd.Language
	user.UpdatedAt = time.Now()

	return h.userRepo.Update(ctx, user)
}
