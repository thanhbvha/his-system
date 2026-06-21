package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"his-system/internal/reception/domain"
	"his-system/pkg/ws"
)

type CheckInCommand struct {
	PatientID     uuid.UUID
	ServiceType   string
	AppointmentID *uuid.UUID
}

func HandleCheckIn(ctx context.Context, cmd CheckInCommand, repo domain.QueueRepository) (*domain.QueueEntry, error) {
	// Prefix generation (e.g. GENERAL -> KB, LAB -> XN, RADIOLOGY -> CD)
	basePrefix := "KB"
	switch cmd.ServiceType {
	case "LAB":
		basePrefix = "XN"
	case "RADIOLOGY":
		basePrefix = "CD"
	case "PHARMACY":
		basePrefix = "DT"
	}

	dateStr := time.Now().Format("060102") // YYMMDD format
	prefix := basePrefix + dateStr

	seq, err := repo.GetNextSequence(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to generate queue number: %w", err)
	}

	queueNumber := fmt.Sprintf("%s%05d", prefix, seq)

	entry := &domain.QueueEntry{
		PatientID:     cmd.PatientID,
		AppointmentID: cmd.AppointmentID,
		ServiceType:   cmd.ServiceType,
		QueueNumber:   queueNumber,
		Status:        domain.StatusWaiting,
		CreatedAt:     time.Now(),
	}

	if err := repo.Save(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to save queue entry: %w", err)
	}

	fullEntry, err := repo.FindByID(ctx, entry.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve full entry: %w", err)
	}

	// Broadcast WS event
	ws.BroadcastToRoom(fullEntry.ServiceType, ws.EventQueueCheckedIn, fullEntry)

	return fullEntry, nil
}
