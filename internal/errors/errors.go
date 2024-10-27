package errors

import "fmt"

// ErrorType represents the type of error
type ErrorType string

const (
	// Error types
	ErrorTypeValidation           ErrorType = "VALIDATION_ERROR"
	ErrorTypeAuthorization        ErrorType = "AUTHORIZATION_ERROR"
	ErrorTypeNotFound             ErrorType = "NOT_FOUND"
	ErrorTypeConflict             ErrorType = "CONFLICT"
	ErrorTypeInternal             ErrorType = "INTERNAL_ERROR"
	ErrorTypeExternal             ErrorType = "EXTERNAL_SERVICE_ERROR"
	ErrorTypeNotSufficientBalance ErrorType = "NOT_SUFFICIENT_BALANCE"
)

// ServiceError represents a service-level error
type ServiceError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Error constructors
func NewValidationError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

func NewAuthorizationError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeAuthorization,
		Message: message,
		Err:     err,
	}
}

func NewNotFoundError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Err:     err,
	}
}

func NewConflictError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeConflict,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

func NewExternalError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeExternal,
		Message: message,
		Err:     err,
	}
}

func NewNotSufficientBalanceError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeNotSufficientBalance,
		Message: message,
		Err:     err,
	}
}
