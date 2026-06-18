package domain

import (
	"context"

	"github.com/google/uuid"
)

type QueueStats struct {
	WaitingCount   int     `json:"waiting_count"`
	CalledCount    int     `json:"called_count"`
	AvgWaitMinutes float64 `json:"avg_wait_minutes"`
}

type QueueRepository interface {
	Save(ctx context.Context, entry *QueueEntry) error
	FindByID(ctx context.Context, id uuid.UUID) (*QueueEntry, error)
	FindTodayQueue(ctx context.Context, serviceType string) ([]*QueueEntry, error)
	GetNextSequence(ctx context.Context, prefix string) (int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status QueueStatus) error
	GetStats(ctx context.Context) (*QueueStats, error)
}
