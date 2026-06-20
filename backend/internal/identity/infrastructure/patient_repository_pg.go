package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"his-system/internal/identity/domain"
)

type PatientRepositoryPG struct {
	db *pgxpool.Pool
}

func NewPatientRepositoryPG(db *pgxpool.Pool) *PatientRepositoryPG {
	return &PatientRepositoryPG{db: db}
}

func (r *PatientRepositoryPG) Create(ctx context.Context, patient *domain.Patient) error {
	q := `INSERT INTO patients (id, full_name, dob, gender, phone_encrypted, phone_hmac, email_encrypted, email_hmac, created_at, updated_at) 
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING patient_code`
	err := r.db.QueryRow(ctx, q,
		patient.ID, patient.FullName, patient.DOB, patient.Gender,
		patient.PhoneEncrypted, patient.PhoneHMAC,
		patient.EmailEncrypted, patient.EmailHMAC,
		patient.CreatedAt, patient.UpdatedAt,
	).Scan(&patient.PatientCode)
	return err
}

func (r *PatientRepositoryPG) GetByPhoneHMAC(ctx context.Context, phoneHMAC string) (*domain.Patient, error) {
	q := `SELECT id, patient_code, full_name, dob, gender, phone_encrypted, phone_hmac, email_encrypted, email_hmac, created_at, updated_at
	      FROM patients WHERE phone_hmac = $1`

	row := r.db.QueryRow(ctx, q, phoneHMAC)
	var p domain.Patient
	err := row.Scan(&p.ID, &p.PatientCode, &p.FullName, &p.DOB, &p.Gender, &p.PhoneEncrypted, &p.PhoneHMAC, &p.EmailEncrypted, &p.EmailHMAC, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil when not found
		}
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Patient, error) {
	q := `SELECT id, patient_code, full_name, dob, gender, phone_encrypted, phone_hmac, email_encrypted, email_hmac, created_at, updated_at
	      FROM patients WHERE id = $1`

	row := r.db.QueryRow(ctx, q, id)
	var p domain.Patient
	err := row.Scan(&p.ID, &p.PatientCode, &p.FullName, &p.DOB, &p.Gender, &p.PhoneEncrypted, &p.PhoneHMAC, &p.EmailEncrypted, &p.EmailHMAC, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil when not found
		}
		return nil, err
	}
	return &p, nil
}
