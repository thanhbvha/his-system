package queries

import (
	"context"

	"his-system/internal/visit/domain"
)

type SearchICD10Query struct {
	Query string
	Limit int
}

func HandleSearchICD10(ctx context.Context, q SearchICD10Query, repo domain.VisitRepository) ([]*domain.ICD10Code, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	return repo.SearchICD10(ctx, q.Query, limit)
}
