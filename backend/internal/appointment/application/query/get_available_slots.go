package query

import (
	"context"
	"time"

	"github.com/google/uuid"
	"his-system/internal/appointment/domain"
)

type GetAvailableSlotsQuery struct {
	DoctorID uuid.UUID
	Date     time.Time
}

type SlotResult struct {
	ID        string `json:"id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type GetAvailableSlotsHandler struct {
	slotRepo domain.SlotRepository
}

func NewGetAvailableSlotsHandler(slotRepo domain.SlotRepository) *GetAvailableSlotsHandler {
	return &GetAvailableSlotsHandler{slotRepo: slotRepo}
}

func (h *GetAvailableSlotsHandler) Handle(ctx context.Context, q GetAvailableSlotsQuery) ([]*SlotResult, error) {
	slots, err := h.slotRepo.GetAvailable(ctx, q.DoctorID, q.Date)
	if err != nil {
		return nil, err
	}

	var results []*SlotResult
	for _, s := range slots {
		results = append(results, &SlotResult{
			ID:        s.ID.String(),
			StartTime: s.StartTime.Format("15:04"),
			EndTime:   s.EndTime.Format("15:04"),
		})
	}
	if results == nil {
		results = make([]*SlotResult, 0)
	}
	return results, nil
}
