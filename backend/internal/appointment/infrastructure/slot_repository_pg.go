package infrastructure

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"his-system/internal/appointment/domain"
)

type SlotRepositoryPG struct {
	db *pgxpool.Pool
}

func NewSlotRepositoryPG(db *pgxpool.Pool) *SlotRepositoryPG {
	return &SlotRepositoryPG{db: db}
}

func (r *SlotRepositoryPG) GetAvailable(ctx context.Context, doctorID uuid.UUID, date time.Time) ([]*domain.Slot, error) {
	query := `
		SELECT id, schedule_id, doctor_id, slot_date, start_time, end_time, is_booked, created_at
		FROM appointment_slots
		WHERE doctor_id = $1 AND slot_date = $2 AND is_booked = false
		ORDER BY start_time ASC`

	rows, err := r.db.Query(ctx, query, doctorID, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []*domain.Slot
	for rows.Next() {
		var s domain.Slot
		if err := rows.Scan(
			&s.ID, &s.ScheduleID, &s.DoctorID, &s.Date, &s.StartTime, &s.EndTime,
			&s.IsBooked, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		slots = append(slots, &s)
	}
	return slots, nil
}

func (r *SlotRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Slot, error) {
	query := `
		SELECT id, schedule_id, doctor_id, slot_date, start_time, end_time, is_booked, created_at
		FROM appointment_slots
		WHERE id = $1`

	var s domain.Slot
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.ScheduleID, &s.DoctorID, &s.Date, &s.StartTime, &s.EndTime,
		&s.IsBooked, &s.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SlotRepositoryPG) BulkCreate(ctx context.Context, slots []*domain.Slot) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO appointment_slots (
			id, schedule_id, doctor_id, slot_date, start_time, end_time, is_booked, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING`

	for _, s := range slots {
		_, err := tx.Exec(ctx, query,
			s.ID, s.ScheduleID, s.DoctorID, s.Date, s.StartTime, s.EndTime, s.IsBooked, s.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *SlotRepositoryPG) BookSlot(ctx context.Context, slotID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// SELECT FOR UPDATE - Anti-Double-Booking
	querySelect := `SELECT id FROM appointment_slots WHERE id = $1 AND is_booked = false FOR UPDATE`
	var id uuid.UUID
	err = tx.QueryRow(ctx, querySelect, slotID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrSlotAlreadyBooked
		}
		return err
	}

	queryUpdate := `UPDATE appointment_slots SET is_booked = true WHERE id = $1`
	tag, err := tx.Exec(ctx, queryUpdate, slotID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSlotAlreadyBooked
	}

	return tx.Commit(ctx)
}

func (r *SlotRepositoryPG) ReleaseSlot(ctx context.Context, slotID uuid.UUID) error {
	query := `UPDATE appointment_slots SET is_booked = false WHERE id = $1`
	_, err := r.db.Exec(ctx, query, slotID)
	return err
}
