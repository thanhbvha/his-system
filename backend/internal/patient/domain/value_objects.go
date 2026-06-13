package domain

import (
	"errors"
	"regexp"

	"his-system/pkg/crypto"
)

var (
	ErrInvalidPhone = errors.New("số điện thoại không hợp lệ")
	ErrInvalidCCCD  = errors.New("cccd không hợp lệ")
	ErrInvalidBHYT  = errors.New("số bhyt không hợp lệ")
	ErrPhoneExists  = errors.New("số điện thoại đã tồn tại")
	ErrCCCDExists   = errors.New("cccd đã tồn tại")
)

type PhoneNumber struct {
	encrypted string
	hmac      string
}

func NewPhoneNumber(plaintext string, cipher *crypto.FieldCipher) (PhoneNumber, error) {
	if plaintext == "" {
		return PhoneNumber{}, nil
	}

	match, _ := regexp.MatchString(`^(0[3|5|7|8|9])[0-9]{8}$`, plaintext)
	if !match {
		return PhoneNumber{}, ErrInvalidPhone
	}

	enc, err := cipher.Encrypt(plaintext)
	if err != nil {
		return PhoneNumber{}, err
	}

	return PhoneNumber{
		encrypted: enc,
		hmac:      cipher.HMAC(plaintext),
	}, nil
}

func (p PhoneNumber) Encrypted() string { return p.encrypted }
func (p PhoneNumber) HMAC() string { return p.hmac }

type CCCD struct {
	encrypted string
	hmac      string
}

func NewCCCD(plaintext string, cipher *crypto.FieldCipher) (CCCD, error) {
	if plaintext == "" {
		return CCCD{}, nil
	}

	match, _ := regexp.MatchString(`^[0-9]{12}$`, plaintext)
	if !match {
		return CCCD{}, ErrInvalidCCCD
	}

	enc, err := cipher.Encrypt(plaintext)
	if err != nil {
		return CCCD{}, err
	}

	return CCCD{
		encrypted: enc,
		hmac:      cipher.HMAC(plaintext),
	}, nil
}

func (c CCCD) Encrypted() string { return c.encrypted }
func (c CCCD) HMAC() string { return c.hmac }

type BHYTNumber struct {
	encrypted string
	hmac      string
}

func NewBHYTNumber(plaintext string, cipher *crypto.FieldCipher) (BHYTNumber, error) {
	if plaintext == "" {
		return BHYTNumber{}, nil
	}

	// 2 uppercase letters + 13 digits
	match, _ := regexp.MatchString(`^[A-Z]{2}[0-9]{13}$`, plaintext)
	if !match {
		return BHYTNumber{}, ErrInvalidBHYT
	}

	enc, err := cipher.Encrypt(plaintext)
	if err != nil {
		return BHYTNumber{}, err
	}

	return BHYTNumber{
		encrypted: enc,
		hmac:      cipher.HMAC(plaintext),
	}, nil
}

func (b BHYTNumber) Encrypted() string { return b.encrypted }
func (b BHYTNumber) HMAC() string { return b.hmac }
