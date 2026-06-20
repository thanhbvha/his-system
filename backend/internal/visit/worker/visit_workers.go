package worker

import (
	"encoding/json"
	"fmt"

	commonLogger "github.com/thanhbvha/go-common/logger"
	"github.com/thanhbvha/go-common/queue"
	"his-system/internal/visit/application/commands"
)

// RegisterVisitWorkers registers all Visit domain job handlers into the queue.
func RegisterVisitWorkers(q *queue.Queue) {
	q.RegisterJobType(commands.JobVisitStarted, queue.JobTypeOptions{Concurrency: 2})
	q.RegisterJobType(commands.JobLabOrderCreated, queue.JobTypeOptions{Concurrency: 2})
	q.RegisterJobType(commands.JobVisitClosed, queue.JobTypeOptions{Concurrency: 2})

	q.RegisterHandler(commands.JobVisitStarted, handleVisitStarted)
	q.RegisterHandler(commands.JobLabOrderCreated, handleLabOrderCreated)
	q.RegisterHandler(commands.JobVisitClosed, handleVisitClosed)
}

func handleVisitStarted(job queue.Job) error {
	var payload commands.VisitStartedPayload
	if err := unmarshalPayload(job.Data, &payload); err != nil {
		return err
	}
	commonLogger.InfoAsync("Worker: VisitStarted",
		"visit_id", payload.VisitID,
		"patient_id", payload.PatientID,
		"doctor_id", payload.DoctorID,
	)
	// TODO: Sprint 7 — trigger billing pre-authorization, notify scheduler, etc.
	return nil
}

func handleLabOrderCreated(job queue.Job) error {
	var payload commands.LabOrderCreatedPayload
	if err := unmarshalPayload(job.Data, &payload); err != nil {
		return err
	}
	commonLogger.InfoAsync("Worker: LabOrderCreated",
		"order_id", payload.OrderID,
		"visit_id", payload.VisitID,
		"order_type", payload.OrderType,
	)
	// TODO: Sprint 5 — route to Lab module, create lab_order record, etc.
	return nil
}

func handleVisitClosed(job queue.Job) error {
	var payload commands.VisitClosedPayload
	if err := unmarshalPayload(job.Data, &payload); err != nil {
		return err
	}
	commonLogger.InfoAsync("Worker: VisitClosed",
		"visit_id", payload.VisitID,
		"patient_id", payload.PatientID,
	)
	// TODO: Sprint 7 — trigger billing finalization, EMR archival, etc.
	return nil
}

func unmarshalPayload(data interface{}, dst interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}
	return json.Unmarshal(b, dst)
}
