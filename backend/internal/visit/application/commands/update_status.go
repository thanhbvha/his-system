package commands

import (
	"context"

	"github.com/google/uuid"
	"his-system/internal/visit/domain"
)

type UpdateVisitStatusCommand struct {
	VisitID   uuid.UUID
	NewStatus domain.VisitStatus
}

func HandleUpdateVisitStatus(ctx context.Context, cmd UpdateVisitStatusCommand, repo domain.VisitRepository) error {
	v, err := repo.FindByID(ctx, cmd.VisitID)
	if err != nil {
		return err
	}

	switch cmd.NewStatus {
	case domain.VisitInProgress:
		if err := v.Start(); err != nil {
			return err
		}
	case domain.VisitCompleted:
		if err := v.Complete(); err != nil {
			return err
		}
	case domain.VisitCancelled:
		if err := v.Cancel(); err != nil {
			return err
		}
	default:
		v.Status = cmd.NewStatus
	}

	return repo.UpdateStatus(ctx, v.ID, v.Status, v)
}
