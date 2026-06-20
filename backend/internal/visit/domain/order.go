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
	ID        uuid.UUID
	VisitID   uuid.UUID
	OrderType OrderType
	RefID     *uuid.UUID
	Details   string
	Status    string
	CreatedAt time.Time
}

type ICD10Code struct {
	Code          string `json:"code"`
	DescriptionVI string `json:"description_vi"`
	Category      string `json:"category"`
}
