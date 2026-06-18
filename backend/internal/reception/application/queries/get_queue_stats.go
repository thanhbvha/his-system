package queries

import (
	"context"

	"his-system/internal/reception/domain"
)

type GetQueueStatsQuery struct{}

func HandleGetQueueStats(ctx context.Context, query GetQueueStatsQuery, repo domain.QueueRepository) (*domain.QueueStats, error) {
	return repo.GetStats(ctx)
}
