package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"his-system/internal/visit/domain"
)

type VisitRepositoryPG struct {
	db *pgxpool.Pool
}

func NewVisitRepositoryPG(db *pgxpool.Pool) *VisitRepositoryPG {
	return &VisitRepositoryPG{db: db}
}

func (r *VisitRepositoryPG) Save(ctx context.Context, v *domain.Visit) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	q := `
		INSERT INTO visits (id, patient_id, doctor_id, queue_entry_id, status, chief_complaint, started_at, completed_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`
	_, err := r.db.Exec(ctx, q,
		v.ID, v.PatientID, v.DoctorID, v.QueueEntryID, v.Status,
		v.ChiefComplaint, v.StartedAt, v.CompletedAt, v.CreatedAt, v.UpdatedAt,
	)
	return err
}

func (r *VisitRepositoryPG) FindByID(ctx context.Context, id uuid.UUID) (*domain.Visit, error) {
	q := `SELECT id, patient_id, doctor_id, queue_entry_id, status, chief_complaint, started_at, completed_at, created_at, updated_at FROM visits WHERE id=$1`
	row := r.db.QueryRow(ctx, q, id)
	return scanVisit(row)
}

func (r *VisitRepositoryPG) FindWorklist(ctx context.Context, doctorID uuid.UUID, date time.Time, status domain.VisitStatus) ([]*domain.VisitWithPatient, error) {
	q := `
		SELECT v.id, v.patient_id, v.doctor_id, v.queue_entry_id, v.status, v.chief_complaint,
		       v.started_at, v.completed_at, v.created_at, v.updated_at,
		       COALESCE(p.full_name, '') as patient_full_name,
		       COALESCE(p.phone, '') as patient_phone
		FROM visits v
		LEFT JOIN patients p ON p.id = v.patient_id
		WHERE v.doctor_id = $1
		  AND v.created_at::date = $2::date
	`
	args := []interface{}{doctorID, date}
	if status != "" {
		q += " AND v.status = $3"
		args = append(args, status)
	}
	q += " ORDER BY v.created_at ASC"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.VisitWithPatient
	for rows.Next() {
		var vp domain.VisitWithPatient
		err := rows.Scan(
			&vp.ID, &vp.PatientID, &vp.DoctorID, &vp.QueueEntryID, &vp.Status,
			&vp.ChiefComplaint, &vp.StartedAt, &vp.CompletedAt, &vp.CreatedAt, &vp.UpdatedAt,
			&vp.PatientFullName, &vp.PatientPhone,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &vp)
	}
	return result, nil
}

func (r *VisitRepositoryPG) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.VisitStatus, v *domain.Visit) error {
	q := `UPDATE visits SET status=$1, started_at=$2, completed_at=$3, updated_at=$4 WHERE id=$5`
	_, err := r.db.Exec(ctx, q, status, v.StartedAt, v.CompletedAt, time.Now(), id)
	return err
}

func (r *VisitRepositoryPG) SaveVital(ctx context.Context, vital *domain.VisitVital) error {
	if vital.ID == uuid.Nil {
		vital.ID = uuid.New()
	}
	q := `
		INSERT INTO visit_vitals (id, visit_id, bp_systolic, bp_diastolic, heart_rate, temperature, spo2, weight_kg, height_cm, recorded_at, recorded_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`
	_, err := r.db.Exec(ctx, q,
		vital.ID, vital.VisitID, vital.BpSystolic, vital.BpDiastolic,
		vital.HeartRate, vital.Temperature, vital.SpO2,
		vital.WeightKg, vital.HeightCm, vital.RecordedAt, vital.RecordedBy,
	)
	return err
}

func (r *VisitRepositoryPG) FindVitalsByVisitID(ctx context.Context, visitID uuid.UUID) ([]*domain.VisitVital, error) {
	q := `
		SELECT id, visit_id, bp_systolic, bp_diastolic, heart_rate, temperature, spo2, weight_kg, height_cm, recorded_at, recorded_by
		FROM visit_vitals WHERE visit_id=$1 ORDER BY recorded_at ASC
	`
	rows, err := r.db.Query(ctx, q, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.VisitVital
	for rows.Next() {
		var vt domain.VisitVital
		if err := rows.Scan(
			&vt.ID, &vt.VisitID, &vt.BpSystolic, &vt.BpDiastolic,
			&vt.HeartRate, &vt.Temperature, &vt.SpO2,
			&vt.WeightKg, &vt.HeightCm, &vt.RecordedAt, &vt.RecordedBy,
		); err != nil {
			return nil, err
		}
		result = append(result, &vt)
	}
	return result, nil
}

func (r *VisitRepositoryPG) SaveOrder(ctx context.Context, order *domain.VisitOrder) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	q := `INSERT INTO visit_orders (id, visit_id, order_type, ref_id, details, status, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.Exec(ctx, q, order.ID, order.VisitID, order.OrderType, order.RefID, order.Details, order.Status, order.CreatedAt)
	return err
}

func (r *VisitRepositoryPG) FindOrdersByVisitID(ctx context.Context, visitID uuid.UUID) ([]*domain.VisitOrder, error) {
	q := `SELECT id, visit_id, order_type, ref_id, details, status, created_at FROM visit_orders WHERE visit_id=$1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, q, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.VisitOrder
	for rows.Next() {
		var o domain.VisitOrder
		if err := rows.Scan(&o.ID, &o.VisitID, &o.OrderType, &o.RefID, &o.Details, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, &o)
	}
	return result, nil
}

func (r *VisitRepositoryPG) SearchICD10(ctx context.Context, query string, limit int) ([]*domain.ICD10Code, error) {
	q := `
		SELECT code, description_vi, COALESCE(category, '') FROM icd10_codes
		WHERE description_tsv @@ plainto_tsquery('simple', $1)
		ORDER BY code
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, q, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.ICD10Code
	for rows.Next() {
		var c domain.ICD10Code
		if err := rows.Scan(&c.Code, &c.DescriptionVI, &c.Category); err != nil {
			return nil, err
		}
		result = append(result, &c)
	}
	return result, nil
}

func scanVisit(row pgx.Row) (*domain.Visit, error) {
	var v domain.Visit
	err := row.Scan(
		&v.ID, &v.PatientID, &v.DoctorID, &v.QueueEntryID, &v.Status,
		&v.ChiefComplaint, &v.StartedAt, &v.CompletedAt, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("visit not found")
		}
		return nil, err
	}
	return &v, nil
}
