package queries

import (
	"context"

	"his-system/internal/reception/domain"
)

type GetCurrentQueueQuery struct {
	ServiceType string
}

func HandleGetCurrentQueue(ctx context.Context, query GetCurrentQueueQuery, repo domain.QueueRepository) ([]*domain.QueueEntry, error) {
	return repo.FindTodayQueue(ctx, query.ServiceType)
}
