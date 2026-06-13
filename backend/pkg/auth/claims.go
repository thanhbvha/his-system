package auth

import "github.com/google/uuid"

type Claims struct {
	UserID        uuid.UUID `json:"sub"`
	Username      string    `json:"usr"`
	Roles         []string  `json:"roles"`
	Permissions   []string  `json:"perms"`
	IssuedAt      int64     `json:"iat"`
	ExpiresAt     int64     `json:"exp"`
	PublicKeyHash string    `json:"-"` // Extracted from outer JWT cnf.jkt claim
}
