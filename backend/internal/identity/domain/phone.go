package domain

import (
	"errors"
	"regexp"

	"his-system/pkg/crypto"
)

var ErrInvalidPhone = errors.New("domain: invalid phone format")

// Validate 10 digits starting with 0
var phoneRegex = regexp.MustCompile(`^0\d{9}$`)

type Phone struct {
	plain     string
	encrypted string
	hmac      string
}

// NewPhone validates and encrypts a phone number
func NewPhone(plain string, cipher *crypto.FieldCipher) (Phone, error) {
	if !phoneRegex.MatchString(plain) {
		return Phone{}, ErrInvalidPhone
	}

	encrypted, err := cipher.Encrypt(plain)
	if err != nil {
		return Phone{}, err
	}

	hmacStr := cipher.HMAC(plain)

	return Phone{
		plain:     plain,
		encrypted: encrypted,
		hmac:      hmacStr,
	}, nil
}

// ReconstructPhone rebuilds a Phone object from DB fields
func ReconstructPhone(encrypted, hmacStr string) Phone {
	return Phone{
		encrypted: encrypted,
		hmac:      hmacStr,
	}
}

func (p Phone) Encrypted() string {
	return p.encrypted
}

func (p Phone) HMAC() string {
	return p.hmac
}

func (p *Phone) Reveal(cipher *crypto.FieldCipher) (string, error) {
	if p.plain != "" {
		return p.plain, nil
	}
	dec, err := cipher.Decrypt(p.encrypted)
	if err != nil {
		return "", err
	}
	p.plain = dec
	return dec, nil
}
