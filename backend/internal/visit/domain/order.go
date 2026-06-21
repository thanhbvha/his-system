package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderType string

const (
	OrderLab       OrderType = "LAB"
	OrderRadiology OrderType = "RADIOLOGY"
	OrderProcedure OrderType = "PROCEDURE"
)

type VisitOrder struct {
	ID        uuid.UUID  `json:"id"`
	VisitID   uuid.UUID  `json:"visit_id"`
	OrderType OrderType  `json:"order_type"`
	RefID     *uuid.UUID `json:"ref_id,omitempty"`
	Details   string     `json:"details"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type ICD10Code struct {
	Code          string `json:"code"`
	DescriptionVI string `json:"description_vi"`
	Category      string `json:"category"`
}
