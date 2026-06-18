package commands

import (
	"context"

	"github.com/google/uuid"
	"his-system/internal/reception/domain"
	"his-system/pkg/ws"
)

type CallQueueCommand struct {
	QueueEntryID uuid.UUID
}

func HandleCallQueue(ctx context.Context, cmd CallQueueCommand, repo domain.QueueRepository) error {
	entry, err := repo.FindByID(ctx, cmd.QueueEntryID)
	if err != nil {
		return err
	}

	if err := entry.Call(); err != nil {
		return err
	}

	if err := repo.UpdateStatus(ctx, entry.ID, entry.Status); err != nil {
		return err
	}

	// Broadcast WS event
	ws.BroadcastToAll(ws.EventQueueCalled, entry)
	return nil
}
