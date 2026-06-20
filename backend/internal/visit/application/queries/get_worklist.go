package queries

import (
	"context"
	"time"

	"github.com/google/uuid"
	"his-system/internal/visit/domain"
)

type GetDoctorWorklistQuery struct {
	DoctorID uuid.UUID
	Date     time.Time
	Status   domain.VisitStatus
}

func HandleGetDoctorWorklist(ctx context.Context, q GetDoctorWorklistQuery, repo domain.VisitRepository) ([]*domain.VisitWithPatient, error) {
	return repo.FindWorklist(ctx, q.DoctorID, q.Date, q.Status)
}
