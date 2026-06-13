package domain

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	DeviceFingerprint string
	PublicKeyPEM      string
	PublicKeyHash     string
	RegisteredAt      time.Time
	IsActive          bool
}
