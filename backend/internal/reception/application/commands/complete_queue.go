package commands

import (
	"context"

	"github.com/google/uuid"
	"his-system/internal/reception/domain"
	"his-system/pkg/ws"
)

type CompleteQueueCommand struct {
	QueueEntryID uuid.UUID
}

func HandleCompleteQueue(ctx context.Context, cmd CompleteQueueCommand, repo domain.QueueRepository) error {
	entry, err := repo.FindByID(ctx, cmd.QueueEntryID)
	if err != nil {
		return err
	}

	if err := entry.Complete(); err != nil {
		return err
	}

	if err := repo.UpdateStatus(ctx, entry.ID, entry.Status); err != nil {
		return err
	}

	ws.BroadcastToAll(ws.EventQueueCompleted, entry)
	return nil
}
