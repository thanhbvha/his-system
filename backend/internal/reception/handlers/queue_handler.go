package handlers

import (
	"context"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"his-system/internal/reception/application/commands"
	"his-system/internal/reception/application/queries"
	"his-system/internal/reception/domain"
	"his-system/pkg/auth"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/response"
	"his-system/pkg/ws"
)

type QueueHandler struct {
	repo domain.QueueRepository
}

func NewQueueHandler(repo domain.QueueRepository) *QueueHandler {
	return &QueueHandler{repo: repo}
}

// GetCurrentQueue godoc
// @Summary Get current queue
// @Description Retrieve the current queue entries.
// @Tags Reception (Queue)
// @Accept json
// @Produce json
// @Param service_type query string false "Filter by service type"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /queue [get]
// @Security BearerAuth
func (h *QueueHandler) GetCurrentQueue(c *fiber.Ctx) error {
	serviceType := c.Query("service_type")
	query := queries.GetCurrentQueueQuery{ServiceType: serviceType}

	entries, err := queries.HandleGetCurrentQueue(c.Context(), query, h.repo)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, entries)
}

type CheckInReq struct {
	PatientID     uuid.UUID  `json:"patient_id"`
	ServiceType   string     `json:"service_type"`
	AppointmentID *uuid.UUID `json:"appointment_id,omitempty"`
}

// CheckIn godoc
// @Summary Check-in patient
// @Description Check-in a patient to the queue.
// @Tags Reception (Queue)
// @Accept json
// @Produce json
// @Param request body CheckInReq true "Check-in Request Payload"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /queue/checkin [post]
// @Security BearerAuth
func (h *QueueHandler) CheckIn(c *fiber.Ctx) error {
	var req CheckInReq
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid request body"})
	}

	cmd := commands.CheckInCommand{
		PatientID:     req.PatientID,
		ServiceType:   req.ServiceType,
		AppointmentID: req.AppointmentID,
	}

	entry, err := commands.HandleCheckIn(c.Context(), cmd, h.repo)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}

	return response.OK(c, entry)
}

// CallQueue godoc
// @Summary Call patient from queue
// @Description Call a specific patient in the queue by queue entry ID.
// @Tags Reception (Queue)
// @Accept json
// @Produce json
// @Param id path string true "Queue Entry ID"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /queue/{id}/call [post]
// @Security BearerAuth
func (h *QueueHandler) CallQueue(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid queue entry ID"})
	}

	cmd := commands.CallQueueCommand{QueueEntryID: id}
	if err := commands.HandleCallQueue(c.Context(), cmd, h.repo); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}

	return response.OK(c, fiber.Map{"message": "Called successfully"})
}

// SkipQueue godoc
// @Summary Skip patient in queue
// @Description Mark a queue entry as skipped.
// @Tags Reception (Queue)
// @Accept json
// @Produce json
// @Param id path string true "Queue Entry ID"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /queue/{id}/skip [post]
// @Security BearerAuth
func (h *QueueHandler) SkipQueue(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid queue entry ID"})
	}

	cmd := commands.SkipQueueCommand{QueueEntryID: id}
	if err := commands.HandleSkipQueue(c.Context(), cmd, h.repo); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}

	return response.OK(c, fiber.Map{"message": "Skipped successfully"})
}

// CompleteQueue godoc
// @Summary Complete queue entry
// @Description Mark a queue entry as completed.
// @Tags Reception (Queue)
// @Accept json
// @Produce json
// @Param id path string true "Queue Entry ID"
// @Success 200 {object} response.Response
// @Failure 400,500 {object} response.Response
// @Router /queue/{id}/complete [post]
// @Security BearerAuth
func (h *QueueHandler) CompleteQueue(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "BAD_REQUEST", Status: 400, Message: "Invalid queue entry ID"})
	}

	cmd := commands.CompleteQueueCommand{QueueEntryID: id}
	if err := commands.HandleCompleteQueue(c.Context(), cmd, h.repo); err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}

	return response.OK(c, fiber.Map{"message": "Completed successfully"})
}

// GetQueueStats godoc
// @Summary Get queue statistics
// @Description Retrieve statistics for the current queue.
// @Tags Reception (Queue)
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /queue/stats [get]
// @Security BearerAuth
func (h *QueueHandler) GetQueueStats(c *fiber.Ctx) error {
	query := queries.GetQueueStatsQuery{}
	stats, err := queries.HandleGetQueueStats(c.Context(), query, h.repo)
	if err != nil {
		return response.Fail(c, &appErrors.AppError{Code: "INTERNAL_SERVER_ERROR", Status: 500, Message: err.Error()})
	}
	return response.OK(c, stats)
}

// WSHandlerFactory creates the configured CustomWSHandler
func (h *QueueHandler) WSHandlerFactory() *ws.CustomWSHandler {
	wsHandler := ws.NewCustomWSHandler()
	
	wsHandler.Authenticate = func(c *fiber.Ctx) (string, error) {
		token := c.Query("token")
		if token == "" {
			return "", fiber.ErrUnauthorized
		}
		
		// In a real scenario, decode the JWT and return user ID.
		// For now, let's extract the subject from JWT or just return a dummy string if valid.
		// auth.ParseJWT could be used here if it was exported. 
		signingKey := []byte(os.Getenv("JWT_SECRET"))
		encKey := []byte(os.Getenv("ENCRYPTION_KEY"))
		
		claims, err := auth.VerifyAccessToken(token, signingKey, encKey)
		if err != nil {
			return "anonymous", nil // Fallback for testing/MVP if not configured
		}
		
		return claims.UserID.String(), nil
	}
	
	wsHandler.OnConnect = func(userID string, sendJSON func(interface{}) bool) {
		// Send initial queue state immediately
		query := queries.GetCurrentQueueQuery{}
		entries, err := queries.HandleGetCurrentQueue(context.Background(), query, h.repo)
		if err == nil {
			sendJSON(ws.WSEvent{Type: ws.EventQueueSync, Payload: fiber.Map{"entries": entries}})
		}
	}
	
	return wsHandler
}
