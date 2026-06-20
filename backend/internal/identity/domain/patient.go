package domain

import (
	"time"

	"github.com/google/uuid"
	"his-system/pkg/crypto"
)

type Patient struct {
	ID             uuid.UUID
	PatientCode    string
	FullName       string
	DOB            *time.Time
	Gender         string
	PhoneEncrypted string
	PhoneHMAC      string
	EmailEncrypted string
	EmailHMAC      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (p *Patient) SetPhone(plain string, cipher *crypto.FieldCipher) error {
	phone, err := NewPhone(plain, cipher)
	if err != nil {
		return err
	}
	p.PhoneEncrypted = phone.Encrypted()
	p.PhoneHMAC = phone.HMAC()
	return nil
}

func (p *Patient) GetPhone(cipher *crypto.FieldCipher) (string, error) {
	phone := ReconstructPhone(p.PhoneEncrypted, p.PhoneHMAC)
	return phone.Reveal(cipher)
}

func (p *Patient) SetEmail(plain string, cipher *crypto.FieldCipher) error {
	if plain == "" {
		return nil
	}
	email, err := NewEmail(plain, cipher)
	if err != nil {
		return err
	}
	p.EmailEncrypted = email.Encrypted()
	p.EmailHMAC = email.HMAC()
	return nil
}

func (p *Patient) GetEmail(cipher *crypto.FieldCipher) (string, error) {
	if p.EmailEncrypted == "" {
		return "", nil
	}
	email := ReconstructEmail(p.EmailEncrypted, p.EmailHMAC)
	return email.Reveal(cipher)
}
