package commands

import (
	"context"

	"github.com/google/uuid"
	"his-system/internal/reception/domain"
	"his-system/pkg/ws"
)

type SkipQueueCommand struct {
	QueueEntryID uuid.UUID
}

func HandleSkipQueue(ctx context.Context, cmd SkipQueueCommand, repo domain.QueueRepository) error {
	entry, err := repo.FindByID(ctx, cmd.QueueEntryID)
	if err != nil {
		return err
	}

	if err := entry.Skip(); err != nil {
		return err
	}

	if err := repo.UpdateStatus(ctx, entry.ID, entry.Status); err != nil {
		return err
	}

	ws.BroadcastToAll(ws.EventQueueUpdated, entry)
	return nil
}
