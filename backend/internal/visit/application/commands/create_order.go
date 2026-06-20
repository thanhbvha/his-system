package commands

import (
	"context"
	"time"

	"github.com/google/uuid"
	commonQueue "github.com/thanhbvha/go-common/queue"
	"his-system/internal/visit/domain"
)

const JobLabOrderCreated = "HIS.VISIT.LabOrderCreated"

type CreateVisitOrderCommand struct {
	VisitID   uuid.UUID
	OrderType domain.OrderType
	Details   string
}

type LabOrderCreatedPayload struct {
	OrderID   string `json:"order_id"`
	VisitID   string `json:"visit_id"`
	OrderType string `json:"order_type"`
	Details   string `json:"details"`
}

func HandleCreateVisitOrder(ctx context.Context, cmd CreateVisitOrderCommand, repo domain.VisitRepository, q *commonQueue.Queue) (*domain.VisitOrder, error) {
	// Update visit status to ORDERED if currently IN_PROGRESS
	v, err := repo.FindByID(ctx, cmd.VisitID)
	if err != nil {
		return nil, err
	}
	if v.Status == domain.VisitInProgress {
		_ = v.SetOrdered()
		_ = repo.UpdateStatus(ctx, v.ID, v.Status, v)
	}

	order := &domain.VisitOrder{
		VisitID:   cmd.VisitID,
		OrderType: cmd.OrderType,
		Details:   cmd.Details,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	if err := repo.SaveOrder(ctx, order); err != nil {
		return nil, err
	}

	// Publish event
	if q != nil {
		_ = q.Enqueue(ctx, JobLabOrderCreated, LabOrderCreatedPayload{
			OrderID:   order.ID.String(),
			VisitID:   order.VisitID.String(),
			OrderType: string(order.OrderType),
			Details:   order.Details,
		})
	}

	return order, nil
}
