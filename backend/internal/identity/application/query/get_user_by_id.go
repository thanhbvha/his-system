package query

import (
	"context"

	"github.com/google/uuid"

	"his-system/internal/identity/domain"
	"his-system/pkg/crypto"
)

type GetUserByIDQuery struct {
	ID uuid.UUID
}

type GetUserByIDResult struct {
	User *UserDTO
}

type GetUserByIDHandler struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
	cipher   *crypto.FieldCipher
}

func NewGetUserByIDHandler(userRepo domain.UserRepository, roleRepo domain.RoleRepository, cipher *crypto.FieldCipher) *GetUserByIDHandler {
	return &GetUserByIDHandler{userRepo: userRepo, roleRepo: roleRepo, cipher: cipher}
}

func (h *GetUserByIDHandler) Handle(ctx context.Context, q GetUserByIDQuery) (*GetUserByIDResult, error) {
	user, err := h.userRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	email, _ := user.GetEmail(h.cipher)

	var roles []RoleDTO
	for _, rid := range user.RoleIDs {
		role, err := h.roleRepo.GetByID(ctx, rid)
		if err == nil && role != nil {
			roles = append(roles, RoleDTO{ID: role.ID, Name: role.Name})
		}
	}

	dto := &UserDTO{
		ID:         user.ID,
		Username:   user.Username,
		Email:      email,
		Roles:      roles,
		IsActive:   user.IsActive,
		MFAEnabled: user.MFAEnabled,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	return &GetUserByIDResult{User: dto}, nil
}
