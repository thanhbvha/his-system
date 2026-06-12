package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var (
	ErrEmptyKey = errors.New("crypto: empty key provided")
)

// FieldCipher là wrapper tiện lợi cho việc mã hoá field-level trong database.
type FieldCipher struct {
	key     []byte // AES-256 key
	hmacKey []byte // HMAC-SHA256 key
}

// NewFieldCipher tạo một FieldCipher mới.
func NewFieldCipher(encKey, hmacKey []byte) (*FieldCipher, error) {
	if len(encKey) != 32 {
		return nil, ErrInvalidKeySize
	}
	if len(hmacKey) == 0 {
		return nil, ErrEmptyKey
	}

	return &FieldCipher{
		key:     encKey,
		hmacKey: hmacKey,
	}, nil
}

// Encrypt mã hoá một string field, trả về chuỗi base64.
func (f *FieldCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil // often empty strings are kept empty or null in DB
	}
	enc, err := EncryptAESGCM([]byte(plaintext), f.key, nil)
	if err != nil {
		return "", err
	}
	return string(enc), nil
}

// Decrypt giải mã một string field đã mã hoá.
func (f *FieldCipher) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	dec, err := DecryptAESGCM([]byte(ciphertext), f.key, nil)
	if err != nil {
		return "", err
	}
	return string(dec), nil
}

// HMAC tạo HMAC-SHA256 deterministic của value, dùng để tìm kiếm (lookup).
// Trả về hex string.
func (f *FieldCipher) HMAC(value string) string {
	if value == "" {
		return ""
	}
	h := hmac.New(sha256.New, f.hmacKey)
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}
