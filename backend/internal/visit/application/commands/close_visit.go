package commands

import (
	"context"
	"time"

	"github.com/google/uuid"
	commonQueue "github.com/thanhbvha/go-common/queue"
	"his-system/internal/visit/domain"
)

const JobVisitClosed = "HIS.VISIT.VisitClosed"

type CloseVisitCommand struct {
	VisitID uuid.UUID
}

type VisitClosedPayload struct {
	VisitID     string    `json:"visit_id"`
	PatientID   string    `json:"patient_id"`
	CompletedAt time.Time `json:"completed_at"`
}

func HandleCloseVisit(ctx context.Context, cmd CloseVisitCommand, repo domain.VisitRepository, q *commonQueue.Queue) error {
	v, err := repo.FindByID(ctx, cmd.VisitID)
	if err != nil {
		return err
	}

	if err := v.Complete(); err != nil {
		return err
	}

	if err := repo.UpdateStatus(ctx, v.ID, v.Status, v); err != nil {
		return err
	}

	// Publish event
	if q != nil && v.CompletedAt != nil {
		_ = q.Enqueue(ctx, JobVisitClosed, VisitClosedPayload{
			VisitID:     v.ID.String(),
			PatientID:   v.PatientID.String(),
			CompletedAt: *v.CompletedAt,
		})
	}

	return nil
}
