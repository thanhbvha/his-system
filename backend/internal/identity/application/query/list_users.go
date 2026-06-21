package query

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"his-system/internal/identity/domain"
	"his-system/pkg/crypto"
)

type ListUsersQuery struct {
	Page   int
	Limit  int
	Search string
}

type UserDTO struct {
	ID             uuid.UUID   `json:"id"`
	Username       string      `json:"username"`
	Email          string      `json:"email"`
	FullName       string      `json:"full_name"`
	DepartmentID   *uuid.UUID  `json:"department_id,omitempty"`
	DepartmentName string      `json:"department_name"`
	Roles          []RoleDTO   `json:"roles"`
	IsActive       bool        `json:"is_active"`
	MFAEnabled     bool        `json:"mfa_enabled"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
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
	db       *pgxpool.Pool
}

func NewListUsersHandler(userRepo domain.UserRepository, roleRepo domain.RoleRepository, cipher *crypto.FieldCipher, db *pgxpool.Pool) *ListUsersHandler {
	return &ListUsersHandler{userRepo: userRepo, roleRepo: roleRepo, cipher: cipher, db: db}
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (*ListUsersResult, error) {
	// Defaults
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 10
	}

	var searchHMAC string
	if q.Search != "" {
		searchHMAC = h.cipher.HMAC(q.Search)
	}

	users, total, err := h.userRepo.List(ctx, q.Page, q.Limit, q.Search, searchHMAC)
	if err != nil {
		return nil, err
	}

	// Fetch roles
	roleCache := make(map[uuid.UUID]RoleDTO)

	// Fetch staff profiles for these users
	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	type profileData struct {
		FullName string
		DeptID   *uuid.UUID
		DeptName string
	}
	profileMap := make(map[uuid.UUID]profileData)

	if len(userIDs) > 0 {
		qProfiles := `
			SELECT sp.user_id, sp.full_name, d.id, d.name 
			FROM staff_profiles sp
			LEFT JOIN departments d ON sp.department_id = d.id
			WHERE sp.user_id = ANY($1)
		`
		rows, err := h.db.Query(ctx, qProfiles, userIDs)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var uid uuid.UUID
				var fn string
				var did *uuid.UUID
				var dname *string
				if err := rows.Scan(&uid, &fn, &did, &dname); err == nil {
					pd := profileData{FullName: fn}
					if did != nil {
						pd.DeptID = did
						pd.DeptName = *dname
					}
					profileMap[uid] = pd
				}
			}
		}
	}

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

		pd := profileMap[u.ID]

		dtos = append(dtos, &UserDTO{
			ID:             u.ID,
			Username:       u.Username,
			Email:          email,
			FullName:       pd.FullName,
			DepartmentID:   pd.DeptID,
			DepartmentName: pd.DeptName,
			Roles:          roles,
			IsActive:       u.IsActive,
			MFAEnabled:     u.MFAEnabled,
			CreatedAt:      u.CreatedAt,
			UpdatedAt:      u.UpdatedAt,
		})
	}

	return &ListUsersResult{
		Users: dtos,
		Total: total,
	}, nil
}

