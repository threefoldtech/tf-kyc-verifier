/*
Package errors contains custom error types and constructors for the application.
This layer is responsible for defining the error types and constructors for the application.
*/
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
	Type ErrorType
	Msg  string
	Err  error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Msg, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Msg)
}

// Error constructors
func NewValidationError(msg string, err error) *ServiceError {
	return &ServiceError{
		Type: ErrorTypeValidation,
		Msg:  msg,
		Err:  err,
	}
}

func NewAuthorizationError(msg string, err error) *ServiceError {
	return &ServiceError{
		Type: ErrorTypeAuthorization,
		Msg:  msg,
		Err:  err,
	}
}

func NewNotFoundError(msg string, err error) *ServiceError {
	return &ServiceError{
		Type: ErrorTypeNotFound,
		Msg:  msg,
		Err:  err,
	}
}

func NewConflictError(msg string, err error) *ServiceError {
	return &ServiceError{
		Type: ErrorTypeConflict,
		Msg:  msg,
		Err:  err,
	}
}

func NewInternalError(msg string, err error) *ServiceError {
	return &ServiceError{
		Type: ErrorTypeInternal,
		Msg:  msg,
		Err:  err,
	}
}

func NewExternalError(msg string, err error) *ServiceError {
	return &ServiceError{
		Type: ErrorTypeExternal,
		Msg:  msg,
		Err:  err,
	}
}

func NewNotSufficientBalanceError(msg string, err error) *ServiceError {
	return &ServiceError{
		Type: ErrorTypeNotSufficientBalance,
		Msg:  msg,
		Err:  err,
	}
}
