package query

import (
	"context"

	"his-system/internal/identity/domain"
)

type ListUsersQuery struct {
	Page  int
	Limit int
	// Note: Role and Search filters can be added here and supported in Repo
}

type ListUsersResult struct {
	Users []*domain.User
	Total int64
}

type ListUsersHandler struct {
	userRepo domain.UserRepository
}

func NewListUsersHandler(userRepo domain.UserRepository) *ListUsersHandler {
	return &ListUsersHandler{userRepo: userRepo}
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (*ListUsersResult, error) {
	// Defaults
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 10
	}

	users, total, err := h.userRepo.List(ctx, q.Page, q.Limit)
	if err != nil {
		return nil, err
	}

	return &ListUsersResult{
		Users: users,
		Total: total,
	}, nil
}
