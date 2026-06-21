package query

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"his-system/pkg/crypto"
)

type GetMeQuery struct {
	UserID uuid.UUID
}

type GetMeResult struct {
	ID                uuid.UUID  `json:"id"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	Roles             []RoleDTO  `json:"roles"`
	MFAEnabled        bool       `json:"mfa_enabled"`
	PreferredLanguage string     `json:"preferred_language"`
	StaffProfile      *StaffProfileDTO `json:"staff_profile,omitempty"`
	Department        *DepartmentDTO   `json:"department,omitempty"`
}

type StaffProfileDTO struct {
	FullName  string `json:"full_name"`
	Title     string `json:"title"`
	Specialty string `json:"specialty"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

type DepartmentDTO struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

type GetMeHandler struct {
	db     *pgxpool.Pool
	cipher *crypto.FieldCipher
}

func NewGetMeHandler(db *pgxpool.Pool, cipher *crypto.FieldCipher) *GetMeHandler {
	return &GetMeHandler{db: db, cipher: cipher}
}

func (h *GetMeHandler) Handle(ctx context.Context, q GetMeQuery) (*GetMeResult, error) {
	// Query user info
	userQ := `SELECT id, username, COALESCE(email_encrypted, ''), COALESCE(email_hmac, ''), COALESCE(mfa_enabled, false), COALESCE(preferred_language, 'vi')
	          FROM users WHERE id = $1`
	
	var res GetMeResult
	var emailEnc, emailHmac string

	err := h.db.QueryRow(ctx, userQ, q.UserID).Scan(
		&res.ID, &res.Username, &emailEnc, &emailHmac, &res.MFAEnabled, &res.PreferredLanguage,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Fetch roles
	roleQ := `SELECT r.id, r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = $1`
	rows, err := h.db.Query(ctx, roleQ, q.UserID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var role RoleDTO
			if err := rows.Scan(&role.ID, &role.Name); err == nil {
				res.Roles = append(res.Roles, role)
			}
		}
	}

	// Fetch staff profile and department
	staffQ := `
		SELECT 
			sp.full_name, sp.title, sp.specialty, sp.avatar_url, sp.bio,
			d.id, d.code, d.name
		FROM staff_profiles sp
		LEFT JOIN departments d ON sp.department_id = d.id
		WHERE sp.user_id = $1
	`
	
	var sp StaffProfileDTO
	var dID *uuid.UUID
	var dCode, dName *string
	err = h.db.QueryRow(ctx, staffQ, q.UserID).Scan(
		&sp.FullName, &sp.Title, &sp.Specialty, &sp.AvatarURL, &sp.Bio,
		&dID, &dCode, &dName,
	)
	if err == nil {
		res.StaffProfile = &sp
		if dID != nil {
			res.Department = &DepartmentDTO{
				ID:   *dID,
				Code: *dCode,
				Name: *dName,
			}
		}
	}

	return &res, nil
}
