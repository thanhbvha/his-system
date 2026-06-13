package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MFARepositoryPG struct {
	db *pgxpool.Pool
}

func NewMFARepositoryPG(db *pgxpool.Pool) *MFARepositoryPG {
	return &MFARepositoryPG{db: db}
}

func (r *MFARepositoryPG) SaveSecret(ctx context.Context, userID uuid.UUID, encryptedSecret string, backupCodes []string) error {
	q := `INSERT INTO mfa_secrets (user_id, encrypted_secret, backup_codes) 
	      VALUES ($1, $2, $3)
		  ON CONFLICT (user_id) 
		  DO UPDATE SET encrypted_secret = EXCLUDED.encrypted_secret, backup_codes = EXCLUDED.backup_codes`

	_, err := r.db.Exec(ctx, q, userID, encryptedSecret, backupCodes)
	return err
}

func (r *MFARepositoryPG) GetSecret(ctx context.Context, userID uuid.UUID) (string, []string, error) {
	q := `SELECT encrypted_secret, backup_codes FROM mfa_secrets WHERE user_id = $1`
	var encryptedSecret string
	var backupCodes []string

	err := r.db.QueryRow(ctx, q, userID).Scan(&encryptedSecret, &backupCodes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, nil
		}
		return "", nil, err
	}
	return encryptedSecret, backupCodes, nil
}
