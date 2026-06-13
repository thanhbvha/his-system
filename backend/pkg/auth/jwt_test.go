package auth

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
)

func generateKey(t *testing.T) []byte {
	k := make([]byte, 32)
	_, err := rand.Read(k)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func TestIssueVerifyRoundTrip(t *testing.T) {
	signKey := generateKey(t)
	encKey := generateKey(t)

	claims := Claims{
		UserID:      uuid.New(),
		Username:    "testuser",
		Roles:       []string{"admin"},
		Permissions: []string{"read", "write"},
		IssuedAt:    time.Now().Unix(),
		ExpiresAt:   time.Now().Add(15 * time.Minute).Unix(),
	}

	token, err := IssueAccessToken(claims, signKey, encKey, "")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	verClaims, err := VerifyAccessToken(token, signKey, encKey)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if verClaims.UserID != claims.UserID || verClaims.Username != claims.Username {
		t.Errorf("Claims mismatch: got %+v, want %+v", verClaims, claims)
	}
}

func TestExpiredToken(t *testing.T) {
	signKey := generateKey(t)
	encKey := generateKey(t)

	claims := Claims{
		UserID:    uuid.New(),
		IssuedAt:  time.Now().Add(-30 * time.Minute).Unix(),
		ExpiresAt: time.Now().Add(-15 * time.Minute).Unix(),
	}

	token, _ := IssueAccessToken(claims, signKey, encKey, "")

	_, err := VerifyAccessToken(token, signKey, encKey)
	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got: %v", err)
	}
}

func TestTamperedToken(t *testing.T) {
	signKey := generateKey(t)
	encKey := generateKey(t)

	claims := Claims{
		UserID:    uuid.New(),
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}

	token, _ := IssueAccessToken(claims, signKey, encKey, "")

	// Tamper the token signature
	tampered := token[:len(token)-5] + "A" + token[len(token)-4:]
	_, err := VerifyAccessToken(tampered, signKey, encKey)
	if err == nil {
		t.Error("Expected error on tampered token")
	}

	// Tamper the enc payload by decrypting with wrong key
	wrongEncKey := generateKey(t)
	_, err = VerifyAccessToken(token, signKey, wrongEncKey)
	if err != ErrTamperedToken {
		t.Errorf("Expected ErrTamperedToken when using wrong enc key, got %v", err)
	}
}

func TestHardwareBoundToken(t *testing.T) {
	signKey := generateKey(t)
	encKey := generateKey(t)

	claims := Claims{
		UserID:    uuid.New(),
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}
	jkt := "some-sha256-hash"

	token, _ := IssueAccessToken(claims, signKey, encKey, jkt)

	verClaims, err := VerifyAccessToken(token, signKey, encKey)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if verClaims.PublicKeyHash != jkt {
		t.Errorf("Expected PublicKeyHash %s, got %s", jkt, verClaims.PublicKeyHash)
	}
}

func TestWebTokenNoCnf(t *testing.T) {
	signKey := generateKey(t)
	encKey := generateKey(t)

	claims := Claims{
		UserID:    uuid.New(),
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}

	token, _ := IssueAccessToken(claims, signKey, encKey, "")

	verClaims, err := VerifyAccessToken(token, signKey, encKey)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if verClaims.PublicKeyHash != "" {
		t.Errorf("Expected empty PublicKeyHash, got %s", verClaims.PublicKeyHash)
	}
}

func TestRefreshTokenUnique(t *testing.T) {
	t1, err := IssueRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	t2, err := IssueRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if t1 == t2 {
		t.Error("Refresh tokens are not unique")
	}
}

func TestHashToken(t *testing.T) {
	token := "some-random-token"
	hash := HashToken(token)
	if len(hash) != 64 { // SHA-256 hex string is 64 chars
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
}

func TestInvalidEncryptionKey(t *testing.T) {
	signKey := generateKey(t)
	encKey := []byte("too-short")

	claims := Claims{
		UserID: uuid.New(),
	}

	_, err := IssueAccessToken(claims, signKey, encKey, "")
	if err == nil {
		t.Error("Expected error when using invalid encryption key size")
	}
}

func TestInvalidTokenString(t *testing.T) {
	signKey := generateKey(t)
	encKey := generateKey(t)
	_, err := VerifyAccessToken("not.a.jwt", signKey, encKey)
	if err == nil {
		t.Error("Expected error for malformed token string")
	}
}
