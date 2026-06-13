package command

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"his-system/internal/identity/domain"
	"his-system/pkg/crypto"
)

type SetupMFACommand struct {
	UserID   uuid.UUID
	Username string
}

type SetupMFAResult struct {
	QRUri       string   `json:"qr_uri"`
	BackupCodes []string `json:"backup_codes"`
}

type SetupMFAHandler struct {
	mfaRepo domain.MFARepository
	encKey  []byte // AES-GCM key for encrypting secrets
}

func NewSetupMFAHandler(mfaRepo domain.MFARepository, encKey []byte) *SetupMFAHandler {
	return &SetupMFAHandler{mfaRepo: mfaRepo, encKey: encKey}
}

func (h *SetupMFAHandler) Handle(ctx context.Context, cmd SetupMFACommand) (*SetupMFAResult, error) {
	// 1. Generate TOTP Key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "HIS",
		AccountName: cmd.Username,
		SecretSize:  32,
	})
	if err != nil {
		return nil, err
	}

	secret := key.Secret()
	qrUri := key.URL()

	// 2. Encrypt Secret
	encSecret, err := crypto.EncryptAESGCM([]byte(secret), h.encKey, []byte(cmd.UserID.String()))
	if err != nil {
		return nil, err
	}

	// 3. Generate Backup Codes (8 codes, 12 hex chars each)
	plainCodes := make([]string, 8)
	hashedCodes := make([]string, 8)
	for i := 0; i < 8; i++ {
		b := make([]byte, 6)
		rand.Read(b)
		code := hex.EncodeToString(b)
		plainCodes[i] = code

		hashed, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		hashedCodes[i] = string(hashed)
	}

	// 4. Save to DB
	if err := h.mfaRepo.SaveSecret(ctx, cmd.UserID, string(encSecret), hashedCodes); err != nil {
		return nil, err
	}

	return &SetupMFAResult{
		QRUri:       qrUri,
		BackupCodes: plainCodes,
	}, nil
}
