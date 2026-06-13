package command

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/thanhbvha/go-common/logger"

	"his-system/internal/identity/domain"
)

type CreateDepartmentCommand struct {
	Code        string
	Name        string
	Description string
}

type CreateDepartmentResult struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CreateDepartmentHandler struct {
	deptRepo domain.DepartmentRepository
}

func NewCreateDepartmentHandler(deptRepo domain.DepartmentRepository) *CreateDepartmentHandler {
	return &CreateDepartmentHandler{deptRepo: deptRepo}
}

func (h *CreateDepartmentHandler) Handle(ctx context.Context, cmd CreateDepartmentCommand) (*CreateDepartmentResult, error) {
	dept := &domain.Department{
		ID:          uuid.New(),
		Code:        cmd.Code,
		Name:        cmd.Name,
		Description: cmd.Description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := h.deptRepo.Create(ctx, dept)
	if err != nil {
		logger.ErrorAsync("CreateDepartmentHandler.Handle: failed to create department", slog.String("error", err.Error()), slog.String("dispatch_time", time.Now().Format(time.RFC3339Nano)))
		return nil, err
	}

	return &CreateDepartmentResult{ID: dept.ID, Name: dept.Name}, nil
}
