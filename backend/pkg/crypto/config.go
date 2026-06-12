package crypto

import (
	"encoding/base64"
	"errors"
	"os"
)

// LoadFieldCipherFromEnv loads keys from environment and creates a FieldCipher
func LoadFieldCipherFromEnv() (*FieldCipher, error) {
	encKeyB64 := os.Getenv("FIELD_ENCRYPTION_KEY")
	hmacKeyB64 := os.Getenv("FIELD_HMAC_KEY")

	if encKeyB64 == "" || hmacKeyB64 == "" {
		return nil, errors.New("FIELD_ENCRYPTION_KEY and FIELD_HMAC_KEY must be set")
	}

	encKey, err := base64.StdEncoding.DecodeString(encKeyB64)
	if err != nil {
		return nil, errors.New("failed to decode FIELD_ENCRYPTION_KEY")
	}

	hmacKey, err := base64.StdEncoding.DecodeString(hmacKeyB64)
	if err != nil {
		return nil, errors.New("failed to decode FIELD_HMAC_KEY")
	}

	return NewFieldCipher(encKey, hmacKey)
}
