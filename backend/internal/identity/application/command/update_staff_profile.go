package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UpdateStaffProfileCommand struct {
	UserID       uuid.UUID
	FullName     string
	DepartmentID uuid.UUID
}

type UpdateStaffProfileHandler struct {
	db *pgxpool.Pool
}

func NewUpdateStaffProfileHandler(db *pgxpool.Pool) *UpdateStaffProfileHandler {
	return &UpdateStaffProfileHandler{db: db}
}

func (h *UpdateStaffProfileHandler) Handle(ctx context.Context, cmd UpdateStaffProfileCommand) error {
	qUpdate := `
		UPDATE staff_profiles 
		SET full_name = $1, department_id = $2 
		WHERE user_id = $3
	`
	res, err := h.db.Exec(ctx, qUpdate, cmd.FullName, cmd.DepartmentID, cmd.UserID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		qInsert := `
			INSERT INTO staff_profiles (id, user_id, full_name, department_id)
			VALUES ($1, $2, $3, $4)
		`
		_, err = h.db.Exec(ctx, qInsert, uuid.New(), cmd.UserID, cmd.FullName, cmd.DepartmentID)
		return err
	}
	return nil
}
