package command

import (
	"context"

	"github.com/google/uuid"

	"his-system/internal/identity/domain"
	"his-system/pkg/crypto"
)

type UpdateUserEmailCommand struct {
	UserID uuid.UUID
	Email  string
}

type UpdateUserEmailHandler struct {
	userRepo domain.UserRepository
	cipher   *crypto.FieldCipher
}

func NewUpdateUserEmailHandler(userRepo domain.UserRepository, cipher *crypto.FieldCipher) *UpdateUserEmailHandler {
	return &UpdateUserEmailHandler{userRepo: userRepo, cipher: cipher}
}

func (h *UpdateUserEmailHandler) Handle(ctx context.Context, cmd UpdateUserEmailCommand) error {
	user, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	email, _ := user.GetEmail(h.cipher)
	if email != "" {
		// already has email, do nothing
		return nil
	}

	emailVO, err := domain.NewEmail(cmd.Email, h.cipher)
	if err != nil {
		return err
	}

	user.EmailEncrypted = emailVO.Encrypted()
	user.EmailHMAC = emailVO.HMAC()

	return h.userRepo.Update(ctx, user)
}
