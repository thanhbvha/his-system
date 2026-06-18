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
	prefix := "KB"
	switch cmd.ServiceType {
	case "LAB":
		prefix = "XN"
	case "RADIOLOGY":
		prefix = "CD"
	case "PHARMACY":
		prefix = "DT"
	}

	seq, err := repo.GetNextSequence(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to generate queue number: %w", err)
	}

	queueNumber := fmt.Sprintf("%s%03d", prefix, seq)

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

	// Broadcast WS event
	// We can fetch the updated list or just broadcast the new entry
	// It's usually better to just broadcast "something changed" or the new state
	ws.BroadcastToAll(ws.EventQueueUpdated, entry)

	return entry, nil
}
