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

	// Auto-skip any currently CALLED patients in the same service type
	todayQueue, err := repo.FindTodayQueue(ctx, entry.ServiceType)
	if err == nil {
		for _, qe := range todayQueue {
			if qe.Status == domain.StatusCalled && qe.ID != entry.ID {
				_ = qe.Skip()
				_ = repo.UpdateStatus(ctx, qe.ID, qe.Status)
				ws.BroadcastToAll(ws.EventQueueSkipped, qe)
			}
		}
	}

	if err := repo.UpdateStatus(ctx, entry.ID, entry.Status); err != nil {
		return err
	}

	// Broadcast WS event
	ws.BroadcastToAll(ws.EventQueueCalled, entry)
	return nil
}
