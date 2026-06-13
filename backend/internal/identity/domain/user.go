package domain

import (
	"time"

	"github.com/google/uuid"
	"his-system/pkg/crypto"
)

type User struct {
	ID             uuid.UUID
	Username       string
	EmailEncrypted string
	EmailHMAC      string
	PasswordHash   string
	RoleIDs        []uuid.UUID
	IsActive       bool
	MFAEnabled     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u *User) SetEmail(plain string, cipher *crypto.FieldCipher) error {
	email, err := NewEmail(plain, cipher)
	if err != nil {
		return err
	}
	u.EmailEncrypted = email.Encrypted()
	u.EmailHMAC = email.HMAC()
	return nil
}

func (u *User) GetEmail(cipher *crypto.FieldCipher) (string, error) {
	email := ReconstructEmail(u.EmailEncrypted, u.EmailHMAC)
	return email.Reveal(cipher)
}
