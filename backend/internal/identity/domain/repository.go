package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmailHMAC(ctx context.Context, emailHMAC string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	List(ctx context.Context, page, limit int, search, searchHMAC string) ([]*User, int64, error)
}

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	UpdatePermissions(ctx context.Context, roleID uuid.UUID, perms []Permission) error
	ListPermissions(ctx context.Context) ([]Permission, error)
}

type DeviceRepository interface {
	Upsert(ctx context.Context, device *Device) error
	GetByUserAndFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*Device, error)
	GetByUserAndPubKeyHash(ctx context.Context, userID uuid.UUID, pubKeyHash string) (*Device, error)
	DeactivateByUser(ctx context.Context, userID uuid.UUID) error
}

type MFARepository interface {
	SaveSecret(ctx context.Context, userID uuid.UUID, encryptedSecret string, backupCodes []string) error
	GetSecret(ctx context.Context, userID uuid.UUID) (string, []string, error)
}

type PatientRepository interface {
	Create(ctx context.Context, patient *Patient) error
	GetByPhoneHMAC(ctx context.Context, phoneHMAC string) (*Patient, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Patient, error)
}
