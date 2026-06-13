package query

import (
	"context"

	"github.com/google/uuid"

	"his-system/internal/identity/domain"
)

type GetUserByIDQuery struct {
	ID uuid.UUID
}

type GetUserByIDResult struct {
	User *domain.User
}

type GetUserByIDHandler struct {
	userRepo domain.UserRepository
}

func NewGetUserByIDHandler(userRepo domain.UserRepository) *GetUserByIDHandler {
	return &GetUserByIDHandler{userRepo: userRepo}
}

func (h *GetUserByIDHandler) Handle(ctx context.Context, q GetUserByIDQuery) (*GetUserByIDResult, error) {
	user, err := h.userRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &GetUserByIDResult{User: user}, nil
}
