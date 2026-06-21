package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"his-system/internal/identity/domain"
)

type UserRepositoryPG struct {
	db *pgxpool.Pool
}

func NewUserRepositoryPG(db *pgxpool.Pool) *UserRepositoryPG {
	return &UserRepositoryPG{db: db}
}

func (r *UserRepositoryPG) Create(ctx context.Context, user *domain.User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := `INSERT INTO users (id, username, email_encrypted, email_hmac, password_hash, is_active, mfa_enabled, preferred_language, created_at, updated_at) 
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.Exec(ctx, q,
		user.ID, user.Username, user.EmailEncrypted, user.EmailHMAC,
		user.PasswordHash, user.IsActive, user.MFAEnabled, user.PreferredLanguage,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if len(user.RoleIDs) > 0 {
		for _, roleID := range user.RoleIDs {
			_, err = tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, user.ID, roleID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *UserRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.getBy(ctx, "id", id)
}

func (r *UserRepositoryPG) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.getBy(ctx, "username", username)
}

func (r *UserRepositoryPG) GetByEmailHMAC(ctx context.Context, emailHMAC string) (*domain.User, error) {
	return r.getBy(ctx, "email_hmac", emailHMAC)
}

func (r *UserRepositoryPG) getBy(ctx context.Context, field string, value interface{}) (*domain.User, error) {
	q := `SELECT id, username, COALESCE(email_encrypted, ''), COALESCE(email_hmac, ''), password_hash, is_active, COALESCE(mfa_enabled, false), COALESCE(preferred_language, 'vi'), created_at, updated_at
	      FROM users WHERE ` + field + ` = $1`

	row := r.db.QueryRow(ctx, q, value)
	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.EmailEncrypted, &u.EmailHMAC, &u.PasswordHash,
		&u.IsActive, &u.MFAEnabled, &u.PreferredLanguage, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil when not found
		}
		return nil, err
	}

	// fetch roles
	rows, err := r.db.Query(ctx, `SELECT role_id FROM user_roles WHERE user_id = $1`, u.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rid uuid.UUID
			if err := rows.Scan(&rid); err == nil {
				u.RoleIDs = append(u.RoleIDs, rid)
			}
		}
	}

	return &u, nil
}

func (r *UserRepositoryPG) Update(ctx context.Context, user *domain.User) error {
	q := `UPDATE users SET 
			username = $1, email_encrypted = $2, email_hmac = $3, 
			password_hash = $4, is_active = $5, mfa_enabled = $6, preferred_language = $7, updated_at = $8
		  WHERE id = $9`
	_, err := r.db.Exec(ctx, q,
		user.Username, user.EmailEncrypted, user.EmailHMAC,
		user.PasswordHash, user.IsActive, user.MFAEnabled, user.PreferredLanguage, user.UpdatedAt,
		user.ID,
	)
	return err
}

func (r *UserRepositoryPG) UpdateRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Xóa role cũ
	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return err
	}

	// Thêm role mới
	for _, roleID := range roleIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, roleID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *UserRepositoryPG) List(ctx context.Context, page, limit int, search, searchHMAC string) ([]*domain.User, int64, error) {
	var total int64
	var err error

	baseCountQ := `SELECT COUNT(*) FROM users`
	baseQ := `
		SELECT id, username, COALESCE(email_encrypted, ''), COALESCE(email_hmac, ''), password_hash, is_active, COALESCE(mfa_enabled, false), COALESCE(preferred_language, 'vi'), created_at, updated_at 
		FROM users`

	if search != "" {
		searchPattern := "%" + search + "%"
		whereClause := ` WHERE username ILIKE $1 OR email_hmac = $2`
		err = r.db.QueryRow(ctx, baseCountQ+whereClause, searchPattern, searchHMAC).Scan(&total)
		if err != nil {
			return nil, 0, err
		}

		offset := (page - 1) * limit
		orderLimitOffset := ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		
		rows, err := r.db.Query(ctx, baseQ+whereClause+orderLimitOffset, searchPattern, searchHMAC, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()
		return r.scanUsers(ctx, rows, total)
	} else {
		err = r.db.QueryRow(ctx, baseCountQ).Scan(&total)
		if err != nil {
			return nil, 0, err
		}

		offset := (page - 1) * limit
		orderLimitOffset := ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		
		rows, err := r.db.Query(ctx, baseQ+orderLimitOffset, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()
		return r.scanUsers(ctx, rows, total)
	}
}

func (r *UserRepositoryPG) scanUsers(ctx context.Context, rows pgx.Rows, total int64) ([]*domain.User, int64, error) {

	var users []*domain.User
	var userIDs []uuid.UUID
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.EmailEncrypted, &u.EmailHMAC, &u.PasswordHash,
			&u.IsActive, &u.MFAEnabled, &u.PreferredLanguage, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
		userIDs = append(userIDs, u.ID)
	}

	// Lấy tất cả roles của các users vừa query
	if len(userIDs) > 0 {
		roleRows, err := r.db.Query(ctx, `SELECT user_id, role_id FROM user_roles WHERE user_id = ANY($1)`, userIDs)
		if err == nil {
			defer roleRows.Close()
			roleMap := make(map[uuid.UUID][]uuid.UUID)
			for roleRows.Next() {
				var uid, rid uuid.UUID
				if err := roleRows.Scan(&uid, &rid); err == nil {
					roleMap[uid] = append(roleMap[uid], rid)
				}
			}
			for _, u := range users {
				u.RoleIDs = roleMap[u.ID]
			}
		}
	}

	return users, total, nil
}
