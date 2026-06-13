package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"his-system/pkg/crypto"
)

var (
	ErrInvalidToken  = errors.New("auth: invalid token")
	ErrExpiredToken  = errors.New("auth: token has expired")
	ErrTamperedToken = errors.New("auth: tampered token or invalid AES decryption")
)

type Cnf struct {
	Jkt string `json:"jkt"`
}

type OuterClaims struct {
	Enc string `json:"enc"`
	Cnf *Cnf   `json:"cnf,omitempty"`
	jwt.RegisteredClaims
}

// IssueAccessToken creates an AES-GCM encrypted JWT payload securely bound to a unique jti.
func IssueAccessToken(claims Claims, signingKey, encKey []byte, publicKeyHash string) (string, error) {
	jti := uuid.New().String()

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encBytes, err := crypto.EncryptAESGCM(claimsJSON, encKey, []byte(jti))
	if err != nil {
		return "", err
	}

	outer := OuterClaims{
		Enc: string(encBytes),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Unix(claims.ExpiresAt, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Unix(claims.IssuedAt, 0)),
		},
	}

	if publicKeyHash != "" {
		outer.Cnf = &Cnf{Jkt: publicKeyHash}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, outer)
	return token.SignedString(signingKey)
}

// VerifyAccessToken validates the HS256 signature, expiration, and decrypts the inner payload.
func VerifyAccessToken(tokenStr string, signingKey, encKey []byte) (Claims, error) {
	var outer OuterClaims

	token, err := jwt.ParseWithClaims(tokenStr, &outer, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return signingKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, ErrExpiredToken
		}
		return Claims{}, ErrInvalidToken
	}

	if !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	decBytes, err := crypto.DecryptAESGCM([]byte(outer.Enc), encKey, []byte(outer.ID))
	if err != nil {
		return Claims{}, ErrTamperedToken
	}

	var claims Claims
	if err := json.Unmarshal(decBytes, &claims); err != nil {
		return Claims{}, ErrTamperedToken
	}

	if outer.Cnf != nil {
		claims.PublicKeyHash = outer.Cnf.Jkt
	}

	return claims, nil
}

// IssueRefreshToken generates a 256-bit random opaque string.
func IssueRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken creates a SHA-256 hash of a token for secure database storage.
func HashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
