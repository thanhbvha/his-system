package queries

import (
	"context"

	"github.com/google/uuid"
	"his-system/internal/visit/domain"
)

type GetVisitDetailQuery struct {
	VisitID uuid.UUID
}

type VisitDetail struct {
	*domain.VisitDetailProjection
	Vitals []*domain.VisitVital `json:"vitals"`
	Orders []*domain.VisitOrder `json:"orders"`
}

func HandleGetVisitDetail(ctx context.Context, q GetVisitDetailQuery, repo domain.VisitRepository) (*VisitDetail, error) {
	v, err := repo.FindDetailByID(ctx, q.VisitID)
	if err != nil {
		return nil, err
	}

	vitals, err := repo.FindVitalsByVisitID(ctx, q.VisitID)
	if err != nil {
		return nil, err
	}

	orders, err := repo.FindOrdersByVisitID(ctx, q.VisitID)
	if err != nil {
		return nil, err
	}

	return &VisitDetail{VisitDetailProjection: v, Vitals: vitals, Orders: orders}, nil
}
