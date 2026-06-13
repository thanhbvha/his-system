package query

import (
	"context"
	"time"

	"github.com/google/uuid"

	"his-system/internal/identity/domain"
	"his-system/pkg/crypto"
)

type ListUsersQuery struct {
	Page   int
	Limit  int
	Search string
}

type UserDTO struct {
	ID         uuid.UUID   `json:"id"`
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	Roles      []RoleDTO   `json:"roles"`
	IsActive   bool        `json:"is_active"`
	MFAEnabled bool        `json:"mfa_enabled"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type RoleDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ListUsersResult struct {
	Users []*UserDTO
	Total int64
}

type ListUsersHandler struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
	cipher   *crypto.FieldCipher
}

func NewListUsersHandler(userRepo domain.UserRepository, roleRepo domain.RoleRepository, cipher *crypto.FieldCipher) *ListUsersHandler {
	return &ListUsersHandler{userRepo: userRepo, roleRepo: roleRepo, cipher: cipher}
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

	// In a real app, optimize this with IN query or caching
	roleCache := make(map[uuid.UUID]RoleDTO)

	var dtos []*UserDTO
	for _, u := range users {
		email, _ := u.GetEmail(h.cipher)
		
		var roles []RoleDTO
		for _, rid := range u.RoleIDs {
			if rDTO, ok := roleCache[rid]; ok {
				roles = append(roles, rDTO)
			} else {
				role, err := h.roleRepo.GetByID(ctx, rid)
				if err == nil && role != nil {
					dto := RoleDTO{ID: role.ID, Name: role.Name}
					roleCache[rid] = dto
					roles = append(roles, dto)
				}
			}
		}

		dtos = append(dtos, &UserDTO{
			ID:         u.ID,
			Username:   u.Username,
			Email:      email,
			Roles:      roles,
			IsActive:   u.IsActive,
			MFAEnabled: u.MFAEnabled,
			CreatedAt:  u.CreatedAt,
			UpdatedAt:  u.UpdatedAt,
		})
	}

	return &ListUsersResult{
		Users: dtos,
		Total: total,
	}, nil
}
