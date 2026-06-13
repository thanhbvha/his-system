package middleware

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"his-system/internal/identity/domain"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
)

// RequestSignature verifies the ECDSA-P256 signature from Desktop clients.
func RequestSignature(deviceRepo domain.DeviceRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := GetClaims(c)
		if !ok {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		if claims.PublicKeyHash == "" {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		// 1. Read Headers
		timestampStr := c.Get("X-Timestamp")
		signatureB64 := c.Get("X-Signature")

		if timestampStr == "" || signatureB64 == "" {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		// 2. Validate Timestamp (Anti-Replay)
		ts, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		reqTime := time.Unix(ts, 0)
		diff := time.Since(reqTime)
		if diff < -5*time.Minute || diff > 5*time.Minute {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		// 3. Lookup Public Key by Hash
		device, err := deviceRepo.GetByUserAndPubKeyHash(c.Context(), claims.UserID, claims.PublicKeyHash)
		if err != nil || device == nil || !device.IsActive {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		// 4. Parse Public Key
		block, _ := pem.Decode([]byte(device.PublicKeyPEM))
		if block == nil {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}
		ecdsaPub, ok := pub.(*ecdsa.PublicKey)
		if !ok {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		// 5. Verify Signature
		method := c.Method()
		url := c.OriginalURL()
		body := string(c.Body())

		message := fmt.Sprintf("%s%s%s%s", method, url, timestampStr, body)
		hash := sha256.Sum256([]byte(message))

		sigBytes, err := base64.StdEncoding.DecodeString(signatureB64)
		if err != nil {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		// Verify ASN.1 ECDSA signature
		if !ecdsa.VerifyASN1(ecdsaPub, hash[:], sigBytes) {
			return response.Fail(c, appErrors.ErrUnauthorized)
		}

		return c.Next()
	}
}
