package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thanhbvha/go-common/logger"
	"golang.org/x/crypto/bcrypt"

	"his-system/internal/identity/domain"
	"his-system/pkg/crypto"
)

type CreateStaffCommand struct {
	Username     string
	Password     string
	Email        string
	RoleIDs      []uuid.UUID
	DepartmentID uuid.UUID
	FullName     string
}

type CreateStaffResult struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type CreateStaffHandler struct {
	userRepo domain.UserRepository
	cipher   *crypto.FieldCipher
	db       *pgxpool.Pool
}

func NewCreateStaffHandler(userRepo domain.UserRepository, cipher *crypto.FieldCipher, db *pgxpool.Pool) *CreateStaffHandler {
	return &CreateStaffHandler{userRepo: userRepo, cipher: cipher, db: db}
}

func (h *CreateStaffHandler) Handle(ctx context.Context, cmd CreateStaffCommand) (*CreateStaffResult, error) {
	emailVO, err := domain.NewEmail(cmd.Email, h.cipher)
	if err != nil {
		return nil, err
	}

	if cmd.Password == "" {
		cmd.Password = "Welcome123!" // Fallback if not provided
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:             uuid.New(),
		Username:       cmd.Username,
		EmailEncrypted: emailVO.Encrypted(),
		EmailHMAC:      emailVO.HMAC(),
		PasswordHash:   string(hash),
		RoleIDs:        cmd.RoleIDs,
		IsActive:       true,
		MFAEnabled:     false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create User
	if err := h.userRepo.Create(ctx, user); err != nil {
		logger.ErrorAsync("CreateStaffHandler.Handle: failed to create user", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	// Insert into staff_profiles
	if cmd.DepartmentID != uuid.Nil || cmd.FullName != "" {
		_, err := h.db.Exec(ctx, `
			INSERT INTO staff_profiles (id, user_id, full_name, department_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING
		`, uuid.New(), user.ID, cmd.FullName, cmd.DepartmentID)
		if err != nil {
			logger.ErrorAsync("CreateStaffHandler.Handle: failed to create staff profile", slog.String("error", err.Error()))
			// we can ignore this error or return it. Let's log it.
		}
	}

	// TODO: Enqueue Job to send email with password if needed
	fmt.Printf("TODO: Created user %s with password: %s\n", cmd.Email, cmd.Password)

	return &CreateStaffResult{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}
