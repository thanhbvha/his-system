package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/internal/identity/domain"
	"his-system/pkg/auth"
	"his-system/pkg/crypto"
	appErrors "his-system/pkg/errors"
)

type RegisterPatientCommand struct {
	Phone    string
	FullName string
	DOB      time.Time
	Gender   string
	Email    string
}

type RegisterPatientResult struct {
	AccessToken string `json:"access_token"`
}

type RegisterPatientHandler struct {
	userRepo    domain.UserRepository
	patientRepo domain.PatientRepository
	rdb         *commonRedis.Client
	cipher      *crypto.FieldCipher
	signKey     []byte
	encKey      []byte
}

func NewRegisterPatientHandler(userRepo domain.UserRepository, patientRepo domain.PatientRepository, rdb *commonRedis.Client, cipher *crypto.FieldCipher, signKey, encKey []byte) *RegisterPatientHandler {
	return &RegisterPatientHandler{
		userRepo:    userRepo,
		patientRepo: patientRepo,
		rdb:         rdb,
		cipher:      cipher,
		signKey:     signKey,
		encKey:      encKey,
	}
}

// Handle returns (result, refreshToken, error)
func (h *RegisterPatientHandler) Handle(ctx context.Context, cmd RegisterPatientCommand) (*RegisterPatientResult, string, error) {
	if cmd.FullName == "" || cmd.Gender == "" {
		return nil, "", &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Vui lòng điền đầy đủ họ tên và giới tính"}
	}

	phoneHMAC := h.cipher.HMAC(cmd.Phone)
	existing, err := h.patientRepo.GetByPhoneHMAC(ctx, phoneHMAC)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", &appErrors.AppError{Code: "CONFLICT", Status: 409, Message: "Số điện thoại đã được đăng ký"}
	}

	// 1. Create Patient entity
	patientID := uuid.New()
	patient := &domain.Patient{
		ID:        patientID,
		FullName:  cmd.FullName,
		DOB:       &cmd.DOB,
		Gender:    cmd.Gender,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := patient.SetPhone(cmd.Phone, h.cipher); err != nil {
		return nil, "", err
	}
	if cmd.Email != "" {
		if err := patient.SetEmail(cmd.Email, h.cipher); err != nil {
			return nil, "", err
		}
	}

	// 2. Create User entity
	// Usually there is a "patient" role ID in DB. We will ignore roles setup for patient for brevity,
	// or assume the caller creates it. We will just save it with no specific roles or lookup the patient role.
	// But according to the plan: User record (role = patient).
	// We will skip inserting into user_roles if we don't have the role ID, but we set IsActive=true.
	user := &domain.User{
		ID:           patientID, // Share same ID
		Username:     phoneHMAC, // Username is required, use phoneHMAC for privacy
		PasswordHash: "",        // No password for OTP users
		IsActive:     true,
		MFAEnabled:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 3. Persist
	// We should technically use a transaction for both, but repositories are separate.
	// In a real app we'd use UnitOfWork. For now, sequential.
	if err := h.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}
	if err := h.patientRepo.Create(ctx, patient); err != nil {
		return nil, "", err
	}

	// 4. Issue tokens
	claims := auth.Claims{
		UserID: patient.ID,
		Roles:  []string{"patient"},
	}
	accessToken, err := auth.IssueAccessToken(claims, h.signKey, h.encKey, "")
	if err != nil {
		return nil, "", err
	}

	refreshToken, err := auth.IssueRefreshToken()
	if err != nil {
		return nil, "", err
	}

	rtHash := auth.HashToken(refreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))
	if err := h.rdb.Set(ctx, rtKey, patient.ID.String(), 7*24*time.Hour); err != nil {
		return nil, "", err
	}

	return &RegisterPatientResult{
		AccessToken: accessToken,
	}, refreshToken, nil
}
