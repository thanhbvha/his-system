package commands

import (
	"context"
	"time"

	"github.com/google/uuid"
	commonQueue "github.com/thanhbvha/go-common/queue"
	"his-system/internal/visit/domain"
)

const JobVisitStarted = "HIS.VISIT.VisitStarted"

type CreateVisitCommand struct {
	PatientID      uuid.UUID
	DoctorID       uuid.UUID
	QueueEntryID   *uuid.UUID
	ChiefComplaint *string
}

type VisitStartedPayload struct {
	VisitID   string    `json:"visit_id"`
	PatientID string    `json:"patient_id"`
	DoctorID  string    `json:"doctor_id"`
	StartedAt time.Time `json:"started_at"`
}

func HandleCreateVisit(ctx context.Context, cmd CreateVisitCommand, repo domain.VisitRepository, q *commonQueue.Queue) (*domain.Visit, error) {
	v := &domain.Visit{
		PatientID:      cmd.PatientID,
		DoctorID:       cmd.DoctorID,
		QueueEntryID:   cmd.QueueEntryID,
		ChiefComplaint: cmd.ChiefComplaint,
		Status:         domain.VisitRegistered,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := repo.Save(ctx, v); err != nil {
		return nil, err
	}

	// Publish event to queue
	if q != nil {
		_ = q.Enqueue(ctx, JobVisitStarted, VisitStartedPayload{
			VisitID:   v.ID.String(),
			PatientID: v.PatientID.String(),
			DoctorID:  v.DoctorID.String(),
			StartedAt: v.CreatedAt,
		})
	}

	return v, nil
}
