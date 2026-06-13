package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"his-system/internal/identity/domain"
)

type DeviceRepositoryPG struct {
	db *pgxpool.Pool
}

func NewDeviceRepositoryPG(db *pgxpool.Pool) *DeviceRepositoryPG {
	return &DeviceRepositoryPG{db: db}
}

func (r *DeviceRepositoryPG) Upsert(ctx context.Context, device *domain.Device) error {
	q := `INSERT INTO device_registry (id, user_id, device_fingerprint, public_key_pem, public_key_hash, registered_at, is_active)
	      VALUES ($1, $2, $3, $4, $5, $6, $7)
		  ON CONFLICT (user_id, device_fingerprint) 
		  DO UPDATE SET public_key_pem = EXCLUDED.public_key_pem, 
		                public_key_hash = EXCLUDED.public_key_hash, 
		                registered_at = EXCLUDED.registered_at, 
		                is_active = EXCLUDED.is_active`

	_, err := r.db.Exec(ctx, q,
		device.ID, device.UserID, device.DeviceFingerprint,
		device.PublicKeyPEM, device.PublicKeyHash, device.RegisteredAt, device.IsActive,
	)
	return err
}

func (r *DeviceRepositoryPG) GetByUserAndFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error) {
	q := `SELECT id, user_id, device_fingerprint, public_key_pem, public_key_hash, registered_at, is_active 
	      FROM device_registry WHERE user_id = $1 AND device_fingerprint = $2`

	row := r.db.QueryRow(ctx, q, userID, fingerprint)
	var d domain.Device
	err := row.Scan(&d.ID, &d.UserID, &d.DeviceFingerprint, &d.PublicKeyPEM, &d.PublicKeyHash, &d.RegisteredAt, &d.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DeviceRepositoryPG) GetByUserAndPubKeyHash(ctx context.Context, userID uuid.UUID, pubKeyHash string) (*domain.Device, error) {
	q := `SELECT id, user_id, device_fingerprint, public_key_pem, public_key_hash, registered_at, is_active 
	      FROM device_registry WHERE user_id = $1 AND public_key_hash = $2`

	row := r.db.QueryRow(ctx, q, userID, pubKeyHash)
	var d domain.Device
	err := row.Scan(&d.ID, &d.UserID, &d.DeviceFingerprint, &d.PublicKeyPEM, &d.PublicKeyHash, &d.RegisteredAt, &d.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DeviceRepositoryPG) DeactivateByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE device_registry SET is_active = false WHERE user_id = $1`, userID)
	return err
}
