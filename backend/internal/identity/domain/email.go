package domain

import (
	"errors"
	"net/mail"

	"his-system/pkg/crypto"
)

var ErrInvalidEmail = errors.New("domain: invalid email format")

type Email struct {
	plain     string
	encrypted string
	hmac      string
}

// NewEmail validates and encrypts an email address
func NewEmail(plain string, cipher *crypto.FieldCipher) (Email, error) {
	_, err := mail.ParseAddress(plain)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}

	encrypted, err := cipher.Encrypt(plain)
	if err != nil {
		return Email{}, err
	}

	hmacStr := cipher.HMAC(plain)

	return Email{
		plain:     plain,
		encrypted: encrypted,
		hmac:      hmacStr,
	}, nil
}

// ReconstructEmail rebuilds an Email object from DB fields
func ReconstructEmail(encrypted, hmacStr string) Email {
	return Email{
		encrypted: encrypted,
		hmac:      hmacStr,
	}
}

func (e Email) Encrypted() string {
	return e.encrypted
}

func (e Email) HMAC() string {
	return e.hmac
}

func (e *Email) Reveal(cipher *crypto.FieldCipher) (string, error) {
	if e.plain != "" {
		return e.plain, nil
	}
	dec, err := cipher.Decrypt(e.encrypted)
	if err != nil {
		return "", err
	}
	e.plain = dec
	return dec, nil
}
