package errors

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrNotFound     = &AppError{Code: "NOT_FOUND", Status: 404, Message: "Resource not found"}
	ErrValidation   = &AppError{Code: "VALIDATION_ERROR", Status: 422, Message: "Validation failed"}
	ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Status: 401, Message: "Unauthorized access"}
	ErrForbidden    = &AppError{Code: "FORBIDDEN", Status: 403, Message: "Access forbidden"}
	ErrConflict     = &AppError{Code: "CONFLICT", Status: 409, Message: "Resource conflict"}
	ErrInternal     = &AppError{Code: "INTERNAL_ERROR", Status: 500, Message: "Internal server error"}
)
