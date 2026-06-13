package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"his-system/internal/patient/domain"
)

type PatientRepositoryPG struct {
	db *pgxpool.Pool
}

func NewPatientRepositoryPG(db *pgxpool.Pool) *PatientRepositoryPG {
	return &PatientRepositoryPG{db: db}
}

func (r *PatientRepositoryPG) Create(ctx context.Context, patient *domain.Patient) error {
	query := `
		INSERT INTO patients (
			id, full_name, dob, gender, blood_type, is_active,
			phone_encrypted, phone_hmac, cccd_encrypted, cccd_hmac,
			email_encrypted, email_hmac, address_detail_encrypted, avatar_url,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16
		)`
	_, err := r.db.Exec(ctx, query,
		patient.ID, patient.FullName, patient.DOB, patient.Gender, patient.BloodType, patient.IsActive,
		patient.PhoneEncrypted, patient.PhoneHMAC, patient.CCCDEncrypted, patient.CCCDHMAC,
		patient.EmailEncrypted, patient.EmailHMAC, patient.AddressDetailEncrypted, patient.AvatarURL,
		patient.CreatedAt, patient.UpdatedAt,
	)
	return err
}

func (r *PatientRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Patient, error) {
	query := `
		SELECT 
			id, full_name, dob, gender, blood_type, is_active,
			phone_encrypted, phone_hmac, cccd_encrypted, cccd_hmac,
			email_encrypted, email_hmac, address_detail_encrypted, avatar_url,
			created_at, updated_at
		FROM patients
		WHERE id = $1`

	var p domain.Patient
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.FullName, &p.DOB, &p.Gender, &p.BloodType, &p.IsActive,
		&p.PhoneEncrypted, &p.PhoneHMAC, &p.CCCDEncrypted, &p.CCCDHMAC,
		&p.EmailEncrypted, &p.EmailHMAC, &p.AddressDetailEncrypted, &p.AvatarURL,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepositoryPG) GetByPhoneHMAC(ctx context.Context, phoneHMAC string) (*domain.Patient, error) {
	query := `
		SELECT 
			id, full_name, dob, gender, blood_type, is_active,
			phone_encrypted, phone_hmac, cccd_encrypted, cccd_hmac,
			email_encrypted, email_hmac, address_detail_encrypted, avatar_url,
			created_at, updated_at
		FROM patients
		WHERE phone_hmac = $1`

	var p domain.Patient
	err := r.db.QueryRow(ctx, query, phoneHMAC).Scan(
		&p.ID, &p.FullName, &p.DOB, &p.Gender, &p.BloodType, &p.IsActive,
		&p.PhoneEncrypted, &p.PhoneHMAC, &p.CCCDEncrypted, &p.CCCDHMAC,
		&p.EmailEncrypted, &p.EmailHMAC, &p.AddressDetailEncrypted, &p.AvatarURL,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepositoryPG) GetByCCCDHMAC(ctx context.Context, cccdHMAC string) (*domain.Patient, error) {
	query := `
		SELECT 
			id, full_name, dob, gender, blood_type, is_active,
			phone_encrypted, phone_hmac, cccd_encrypted, cccd_hmac,
			email_encrypted, email_hmac, address_detail_encrypted, avatar_url,
			created_at, updated_at
		FROM patients
		WHERE cccd_hmac = $1`

	var p domain.Patient
	err := r.db.QueryRow(ctx, query, cccdHMAC).Scan(
		&p.ID, &p.FullName, &p.DOB, &p.Gender, &p.BloodType, &p.IsActive,
		&p.PhoneEncrypted, &p.PhoneHMAC, &p.CCCDEncrypted, &p.CCCDHMAC,
		&p.EmailEncrypted, &p.EmailHMAC, &p.AddressDetailEncrypted, &p.AvatarURL,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepositoryPG) SearchByName(ctx context.Context, q string, page, limit int) ([]*domain.Patient, int64, error) {
	offset := (page - 1) * limit
	countQuery := `SELECT COUNT(*) FROM patients WHERE full_name_search @@ plainto_tsquery('simple', $1)`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, q).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id, full_name, dob, gender, blood_type, is_active,
			phone_encrypted, phone_hmac, cccd_encrypted, cccd_hmac,
			email_encrypted, email_hmac, address_detail_encrypted, avatar_url,
			created_at, updated_at
		FROM patients
		WHERE full_name_search @@ plainto_tsquery('simple', $1)
		ORDER BY ts_rank(full_name_search, plainto_tsquery('simple', $1)) DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var patients []*domain.Patient
	for rows.Next() {
		var p domain.Patient
		if err := rows.Scan(
			&p.ID, &p.FullName, &p.DOB, &p.Gender, &p.BloodType, &p.IsActive,
			&p.PhoneEncrypted, &p.PhoneHMAC, &p.CCCDEncrypted, &p.CCCDHMAC,
			&p.EmailEncrypted, &p.EmailHMAC, &p.AddressDetailEncrypted, &p.AvatarURL,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		patients = append(patients, &p)
	}
	return patients, total, nil
}

func (r *PatientRepositoryPG) Update(ctx context.Context, patient *domain.Patient) error {
	query := `
		UPDATE patients SET
			full_name = $1, dob = $2, gender = $3, blood_type = $4, is_active = $5,
			phone_encrypted = $6, phone_hmac = $7, cccd_encrypted = $8, cccd_hmac = $9,
			email_encrypted = $10, email_hmac = $11, address_detail_encrypted = $12, avatar_url = $13,
			updated_at = $14
		WHERE id = $15`

	_, err := r.db.Exec(ctx, query,
		patient.FullName, patient.DOB, patient.Gender, patient.BloodType, patient.IsActive,
		patient.PhoneEncrypted, patient.PhoneHMAC, patient.CCCDEncrypted, patient.CCCDHMAC,
		patient.EmailEncrypted, patient.EmailHMAC, patient.AddressDetailEncrypted, patient.AvatarURL,
		patient.UpdatedAt, patient.ID,
	)
	return err
}

func (r *PatientRepositoryPG) List(ctx context.Context, page, limit int) ([]*domain.Patient, int64, error) {
	offset := (page - 1) * limit
	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM patients").Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id, full_name, dob, gender, blood_type, is_active,
			phone_encrypted, phone_hmac, cccd_encrypted, cccd_hmac,
			email_encrypted, email_hmac, address_detail_encrypted, avatar_url,
			created_at, updated_at
		FROM patients
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var patients []*domain.Patient
	for rows.Next() {
		var p domain.Patient
		if err := rows.Scan(
			&p.ID, &p.FullName, &p.DOB, &p.Gender, &p.BloodType, &p.IsActive,
			&p.PhoneEncrypted, &p.PhoneHMAC, &p.CCCDEncrypted, &p.CCCDHMAC,
			&p.EmailEncrypted, &p.EmailHMAC, &p.AddressDetailEncrypted, &p.AvatarURL,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		patients = append(patients, &p)
	}
	return patients, total, nil
}

func (r *PatientRepositoryPG) UpsertInsurance(ctx context.Context, ins *domain.PatientInsurance) error {
	query := `
		INSERT INTO patient_insurance (
			id, patient_id, bhyt_number_encrypted, bhyt_hmac,
			valid_from, valid_to, coverage_level, issuing_province
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			bhyt_number_encrypted = EXCLUDED.bhyt_number_encrypted,
			bhyt_hmac = EXCLUDED.bhyt_hmac,
			valid_from = EXCLUDED.valid_from,
			valid_to = EXCLUDED.valid_to,
			coverage_level = EXCLUDED.coverage_level,
			issuing_province = EXCLUDED.issuing_province,
			updated_at = NOW()`

	_, err := r.db.Exec(ctx, query,
		ins.ID, ins.PatientID, ins.BHYTNumberEncrypted, ins.BHYTNumberHMAC,
		ins.ValidFrom, ins.ValidTo, ins.CoverageLevel, ins.IssuingProvince,
	)
	return err
}

func (r *PatientRepositoryPG) GetInsurance(ctx context.Context, patientID uuid.UUID) (*domain.PatientInsurance, error) {
	query := `
		SELECT 
			id, patient_id, bhyt_number_encrypted, bhyt_hmac,
			valid_from, valid_to, coverage_level, issuing_province
		FROM patient_insurance
		WHERE patient_id = $1
		ORDER BY created_at DESC LIMIT 1`

	var ins domain.PatientInsurance
	err := r.db.QueryRow(ctx, query, patientID).Scan(
		&ins.ID, &ins.PatientID, &ins.BHYTNumberEncrypted, &ins.BHYTNumberHMAC,
		&ins.ValidFrom, &ins.ValidTo, &ins.CoverageLevel, &ins.IssuingProvince,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &ins, nil
}
