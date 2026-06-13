package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"his-system/internal/appointment/domain"
)

type AppointmentRepositoryPG struct {
	db *pgxpool.Pool
}

func NewAppointmentRepositoryPG(db *pgxpool.Pool) *AppointmentRepositoryPG {
	return &AppointmentRepositoryPG{db: db}
}

func (r *AppointmentRepositoryPG) Create(ctx context.Context, appt *domain.Appointment) error {
	query := `
		INSERT INTO appointments (
			id, patient_id, doctor_id, service_id, slot_id,
			scheduled_at, status, note, cancel_reason,
			booked_by, booked_at, confirmed_at, checked_in_at, completed_at, cancelled_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16, $17
		)`
	_, err := r.db.Exec(ctx, query,
		appt.ID, appt.PatientID, appt.DoctorID, appt.ServiceID, appt.SlotID,
		appt.ScheduledAt, appt.Status, appt.Note, appt.CancelReason,
		appt.BookedBy, appt.BookedAt, appt.ConfirmedAt, appt.CheckedInAt, appt.CompletedAt, appt.CancelledAt,
		appt.CreatedAt, appt.UpdatedAt,
	)
	return err
}

func (r *AppointmentRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Appointment, error) {
	query := `
		SELECT 
			id, patient_id, doctor_id, service_id, slot_id,
			scheduled_at, status, note, cancel_reason,
			booked_by, booked_at, confirmed_at, checked_in_at, completed_at, cancelled_at,
			created_at, updated_at
		FROM appointments
		WHERE id = $1`

	var a domain.Appointment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.PatientID, &a.DoctorID, &a.ServiceID, &a.SlotID,
		&a.ScheduledAt, &a.Status, &a.Note, &a.CancelReason,
		&a.BookedBy, &a.BookedAt, &a.ConfirmedAt, &a.CheckedInAt, &a.CompletedAt, &a.CancelledAt,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AppointmentRepositoryPG) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Appointment, int64, error) {
	queryBase := `
		SELECT 
			id, patient_id, doctor_id, service_id, slot_id,
			scheduled_at, status, note, cancel_reason,
			booked_by, booked_at, confirmed_at, checked_in_at, completed_at, cancelled_at,
			created_at, updated_at
		FROM appointments
		WHERE 1=1`
	countBase := `SELECT COUNT(*) FROM appointments WHERE 1=1`

	var args []interface{}
	var conditions []string
	argID := 1

	if filter.Date != nil {
		conditions = append(conditions, fmt.Sprintf("DATE(scheduled_at) = $%d", argID))
		args = append(args, filter.Date.Format("2006-01-02"))
		argID++
	}
	if filter.DoctorID != nil {
		conditions = append(conditions, fmt.Sprintf("doctor_id = $%d", argID))
		args = append(args, *filter.DoctorID)
		argID++
	}
	if filter.PatientID != nil {
		conditions = append(conditions, fmt.Sprintf("patient_id = $%d", argID))
		args = append(args, *filter.PatientID)
		argID++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argID))
		args = append(args, *filter.Status)
		argID++
	}

	if len(conditions) > 0 {
		condStr := " AND " + strings.Join(conditions, " AND ")
		queryBase += condStr
		countBase += condStr
	}

	var total int64
	if err := r.db.QueryRow(ctx, countBase, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	offset := (filter.Page - 1) * filter.Limit

	queryBase += fmt.Sprintf(" ORDER BY scheduled_at ASC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(ctx, queryBase, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var appts []*domain.Appointment
	for rows.Next() {
		var a domain.Appointment
		if err := rows.Scan(
			&a.ID, &a.PatientID, &a.DoctorID, &a.ServiceID, &a.SlotID,
			&a.ScheduledAt, &a.Status, &a.Note, &a.CancelReason,
			&a.BookedBy, &a.BookedAt, &a.ConfirmedAt, &a.CheckedInAt, &a.CompletedAt, &a.CancelledAt,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		appts = append(appts, &a)
	}

	return appts, total, nil
}

func (r *AppointmentRepositoryPG) Update(ctx context.Context, appt *domain.Appointment) error {
	query := `
		UPDATE appointments SET
			status = $1, note = $2, cancel_reason = $3,
			confirmed_at = $4, checked_in_at = $5, completed_at = $6, cancelled_at = $7,
			updated_at = $8
		WHERE id = $9`

	_, err := r.db.Exec(ctx, query,
		appt.Status, appt.Note, appt.CancelReason,
		appt.ConfirmedAt, appt.CheckedInAt, appt.CompletedAt, appt.CancelledAt,
		appt.UpdatedAt, appt.ID,
	)
	return err
}
