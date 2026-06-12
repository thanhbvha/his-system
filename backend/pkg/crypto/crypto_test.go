package crypto

import (
	"bytes"
	"encoding/base64"
	"os"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("hello world 123")

	ciphertext, err := EncryptAESGCM(plaintext, key, nil)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := DecryptAESGCM(ciphertext, key, nil)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Expected %s, got %s", plaintext, decrypted)
	}
}

func TestEncryptProducesUniqueCiphertexts(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("secret")

	c1, _ := EncryptAESGCM(plaintext, key, nil)
	c2, _ := EncryptAESGCM(plaintext, key, nil)

	if bytes.Equal(c1, c2) {
		t.Errorf("Ciphertexts should be unique due to random nonce")
	}
}

func TestInvalidKeySize(t *testing.T) {
	key := make([]byte, 31) // invalid
	plaintext := []byte("hello")

	_, err := EncryptAESGCM(plaintext, key, nil)
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got %v", err)
	}

	_, err = DecryptAESGCM([]byte("some_data"), key, nil)
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got %v", err)
	}
}

func TestTamperedCiphertextDetected(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("tamper me")

	ciphertext, _ := EncryptAESGCM(plaintext, key, nil)

	// Tamper: flip a bit in base64. First decode base64.
	decodedLen := base64.RawURLEncoding.DecodedLen(len(ciphertext))
	decoded := make([]byte, decodedLen)
	n, _ := base64.RawURLEncoding.Decode(decoded, ciphertext)
	decoded = decoded[:n]

	// Flip last byte (part of tag)
	decoded[len(decoded)-1] ^= 0x01

	tamperedEncoded := make([]byte, base64.RawURLEncoding.EncodedLen(len(decoded)))
	base64.RawURLEncoding.Encode(tamperedEncoded, decoded)

	_, err := DecryptAESGCM(tamperedEncoded, key, nil)
	if err != ErrInvalidCiphertext {
		t.Errorf("Expected ErrInvalidCiphertext for tampered data, got %v", err)
	}
	
	// Test invalid base64
	_, err = DecryptAESGCM([]byte("invalid^base64"), key, nil)
	if err != ErrInvalidCiphertext {
		t.Errorf("Expected ErrInvalidCiphertext for invalid base64, got %v", err)
	}
	
	// Test too short ciphertext
	short := make([]byte, 10)
	shortEncoded := make([]byte, base64.RawURLEncoding.EncodedLen(len(short)))
	base64.RawURLEncoding.Encode(shortEncoded, short)
	_, err = DecryptAESGCM(shortEncoded, key, nil)
	if err != ErrInvalidCiphertext {
		t.Errorf("Expected ErrInvalidCiphertext for too short ciphertext, got %v", err)
	}
}

func TestHMACIsDeterministic(t *testing.T) {
	encKey := make([]byte, 32)
	hmacKey := []byte("secret-hmac-key")
	fc, _ := NewFieldCipher(encKey, hmacKey)

	h1 := fc.HMAC("value1")
	h2 := fc.HMAC("value1")

	if h1 != h2 {
		t.Errorf("HMAC should be deterministic")
	}

	h3 := fc.HMAC("value2")
	if h1 == h3 {
		t.Errorf("HMAC should differ for different inputs")
	}
}

func TestFieldCipherRoundTrip(t *testing.T) {
	encKey := make([]byte, 32)
	hmacKey := []byte("secret")
	fc, _ := NewFieldCipher(encKey, hmacKey)

	plaintext := "my private data"
	encrypted, err := fc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := fc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if plaintext != decrypted {
		t.Errorf("Expected %s, got %s", plaintext, decrypted)
	}
}

func TestEncryptEmptyString(t *testing.T) {
	encKey := make([]byte, 32)
	hmacKey := []byte("secret")
	fc, _ := NewFieldCipher(encKey, hmacKey)

	enc, err := fc.Encrypt("")
	if err != nil {
		t.Errorf("Encrypt empty string failed: %v", err)
	}
	if enc != "" {
		t.Errorf("Expected empty string, got %s", enc)
	}

	dec, err := fc.Decrypt("")
	if err != nil {
		t.Errorf("Decrypt empty string failed: %v", err)
	}
	if dec != "" {
		t.Errorf("Expected empty string, got %s", dec)
	}

	h := fc.HMAC("")
	if h != "" {
		t.Errorf("Expected empty string for empty HMAC, got %s", h)
	}
}

func TestFieldCipherInvalidInit(t *testing.T) {
	_, err := NewFieldCipher(make([]byte, 10), []byte("secret"))
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize")
	}

	_, err = NewFieldCipher(make([]byte, 32), nil)
	if err != ErrEmptyKey {
		t.Errorf("Expected ErrEmptyKey")
	}
}

func TestLoadFieldCipherFromEnv(t *testing.T) {
	os.Clearenv()
	_, err := LoadFieldCipherFromEnv()
	if err == nil {
		t.Errorf("Expected error when env vars are missing")
	}

	// Valid base64 but wrong length
	os.Setenv("FIELD_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 10)))
	os.Setenv("FIELD_HMAC_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	_, err = LoadFieldCipherFromEnv()
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize")
	}

	// Valid 32-byte key
	os.Setenv("FIELD_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	fc, err := LoadFieldCipherFromEnv()
	if err != nil {
		t.Errorf("Expected success, got %v", err)
	}
	if fc == nil {
		t.Errorf("Expected non-nil FieldCipher")
	}
	
	// Invalid base64
	os.Setenv("FIELD_ENCRYPTION_KEY", "not-base64")
	_, err = LoadFieldCipherFromEnv()
	if err == nil {
		t.Errorf("Expected error on invalid base64")
	}
	
	os.Setenv("FIELD_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	os.Setenv("FIELD_HMAC_KEY", "not-base64")
	_, err = LoadFieldCipherFromEnv()
	if err == nil {
		t.Errorf("Expected error on invalid base64")
	}
}

func TestFieldCipherUnreachableErrors(t *testing.T) {
	fc := &FieldCipher{
		key: make([]byte, 10), // invalid key length
	}
	_, err := fc.Encrypt("test")
	if err == nil {
		t.Errorf("Expected error")
	}
	
	// Valid base64 but will fail decrypt due to invalid key
	validBase64 := base64.RawURLEncoding.EncodeToString([]byte("dummydata"))
	_, err = fc.Decrypt(validBase64)
	if err == nil {
		t.Errorf("Expected error")
	}
}

func TestAESGCMWithAdditionalData(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("secret with AD")
	ad := []byte("some metadata context")

	ciphertext, err := EncryptAESGCM(plaintext, key, ad)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Decrypt with correct AD
	decrypted, err := DecryptAESGCM(ciphertext, key, ad)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Expected %s, got %s", plaintext, decrypted)
	}

	// Decrypt with wrong AD should fail
	_, err = DecryptAESGCM(ciphertext, key, []byte("wrong AD"))
	if err == nil {
		t.Fatalf("Decrypt should fail with wrong AD")
	}
}
