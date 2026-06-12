package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrInvalidKeySize    = errors.New("crypto: invalid key size, must be 32 bytes")
	ErrInvalidCiphertext = errors.New("crypto: invalid ciphertext")
)

// EncryptAESGCM mã hoá plaintext bằng AES-256-GCM.
// Output format: nonce(12B) | ciphertext | tag(16B) → base64 URL-safe
func EncryptAESGCM(plaintext, key, additionalData []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, additionalData)

	encoded := make([]byte, base64.RawURLEncoding.EncodedLen(len(ciphertext)))
	base64.RawURLEncoding.Encode(encoded, ciphertext)

	return encoded, nil
}

// DecryptAESGCM giải mã ciphertext đã được mã hoá bởi EncryptAESGCM.
func DecryptAESGCM(encodedCiphertext, key, additionalData []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}

	decodedLen := base64.RawURLEncoding.DecodedLen(len(encodedCiphertext))
	ciphertext := make([]byte, decodedLen)
	n, err := base64.RawURLEncoding.Decode(ciphertext, encodedCiphertext)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}
	ciphertext = ciphertext[:n]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, actualCiphertext, additionalData)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	return plaintext, nil
}
