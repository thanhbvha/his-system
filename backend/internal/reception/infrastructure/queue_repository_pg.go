package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"his-system/internal/reception/domain"
)

type QueueRepositoryPG struct {
	db *pgxpool.Pool
}

func NewQueueRepositoryPG(db *pgxpool.Pool) *QueueRepositoryPG {
	return &QueueRepositoryPG{db: db}
}

func (r *QueueRepositoryPG) Save(ctx context.Context, entry *domain.QueueEntry) error {
	query := `
		INSERT INTO queue_entries (id, patient_id, visit_id, appointment_id, service_type, queue_number, status, called_at, completed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			called_at = EXCLUDED.called_at,
			completed_at = EXCLUDED.completed_at
	`
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(ctx, query,
		entry.ID, entry.PatientID, entry.VisitID, entry.AppointmentID,
		entry.ServiceType, entry.QueueNumber, entry.Status,
		entry.CalledAt, entry.CompletedAt, entry.CreatedAt,
	)
	return err
}

func (r *QueueRepositoryPG) FindByID(ctx context.Context, id uuid.UUID) (*domain.QueueEntry, error) {
	query := `
		SELECT q.id, q.patient_id, q.visit_id, q.appointment_id, q.service_type, q.queue_number, q.status, q.called_at, q.completed_at, q.created_at,
		       p.full_name, p.patient_code
		FROM queue_entries q
		LEFT JOIN patients p ON p.id = q.patient_id
		WHERE q.id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	return r.scanEntry(row)
}

func (r *QueueRepositoryPG) FindTodayQueue(ctx context.Context, serviceType string) ([]*domain.QueueEntry, error) {
	query := `
		SELECT q.id, q.patient_id, q.visit_id, q.appointment_id, q.service_type, q.queue_number, q.status, q.called_at, q.completed_at, q.created_at,
		       p.full_name, p.patient_code
		FROM queue_entries q
		LEFT JOIN patients p ON p.id = q.patient_id
		WHERE (q.created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date = (now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date
	`
	var args []interface{}
	if serviceType != "" {
		query += " AND q.service_type = $1"
		args = append(args, serviceType)
	}
	query += " ORDER BY q.created_at ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.QueueEntry
	for rows.Next() {
		entry, err := r.scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *QueueRepositoryPG) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.QueueStatus) error {
	query := `UPDATE queue_entries SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *QueueRepositoryPG) GetNextSequence(ctx context.Context, prefix string) (int, error) {
	// A simple sequence generation based on count for today.
	// In production, we might want to use a dedicated sequence table or Redis INCR to ensure atomicity.
	// Using count is fine for small scale, but let's use a robust advisory lock or sequence table approach if needed.
	// For MVP: Count + 1. Note: This can cause duplicates in highly concurrent environments without locking.
	// Let's use a basic COUNT(*) approach for now.
	query := `
		SELECT COUNT(*) FROM queue_entries
		WHERE (created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date = (now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date
		AND queue_number LIKE $1
	`
	var count int
	err := r.db.QueryRow(ctx, query, prefix+"%").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

func (r *QueueRepositoryPG) GetStats(ctx context.Context) (*domain.QueueStats, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'WAITING') as waiting_count,
			COUNT(*) FILTER (WHERE status = 'CALLED') as called_count,
			COALESCE(AVG(EXTRACT(EPOCH FROM (called_at - created_at))/60) FILTER (WHERE called_at IS NOT NULL), 0) as avg_wait_minutes
		FROM queue_entries
		WHERE (created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date = (now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date
	`
	var stats domain.QueueStats
	err := r.db.QueryRow(ctx, query).Scan(&stats.WaitingCount, &stats.CalledCount, &stats.AvgWaitMinutes)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *QueueRepositoryPG) scanEntry(row pgx.Row) (*domain.QueueEntry, error) {
	var e domain.QueueEntry
	var pFullName, pCode *string
	err := row.Scan(
		&e.ID, &e.PatientID, &e.VisitID, &e.AppointmentID,
		&e.ServiceType, &e.QueueNumber, &e.Status,
		&e.CalledAt, &e.CompletedAt, &e.CreatedAt,
		&pFullName, &pCode,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("queue entry not found")
		}
		return nil, err
	}
	
	if pFullName != nil && pCode != nil {
		e.Patient = &domain.QueuePatient{
			ID:          e.PatientID,
			FullName:    *pFullName,
			PatientCode: *pCode,
		}
	}
	
	return &e, nil
}
