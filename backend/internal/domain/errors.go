package domain

import "errors"

// ErrorCode is a stable, machine-readable error identifier returned to clients.
type ErrorCode string

// Error codes, mapped to HTTP statuses by httpx.StatusFor.
const (
	CodeValidation      ErrorCode = "validation_failed"
	CodeUnauthorized    ErrorCode = "unauthorized"
	CodeForbidden       ErrorCode = "forbidden"
	CodeNotFound        ErrorCode = "not_found"
	CodeConflict        ErrorCode = "conflict"
	CodeSlotUnavailable ErrorCode = "slot_unavailable"
	CodeHoldExpired     ErrorCode = "hold_expired"
	CodeHoldNotOwned    ErrorCode = "hold_not_owned"
	CodeRateLimited     ErrorCode = "rate_limited"
	CodeInternal        ErrorCode = "internal_error"
)

// Error is a domain error carrying a code, a client-safe message and an
// optional internal cause.
type Error struct {
	Code    ErrorCode
	Message string
	Details map[string]any
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

// Is matches any domain error with the same code, so errors.Is works with the
// sentinel errors below.
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == other.Code
}

// NewError creates a domain error.
func NewError(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Validation creates a validation error with per-field details.
func Validation(message string, details map[string]any) *Error {
	return &Error{Code: CodeValidation, Message: message, Details: details}
}

// ValidationField creates a validation error for a single field.
func ValidationField(field, message string) *Error {
	return Validation("invalid request", map[string]any{field: message})
}

// WithCause returns a copy of the error that wraps an internal cause.
func (e *Error) WithCause(err error) *Error {
	clone := *e
	clone.Cause = err
	return &clone
}

// IsCode reports whether err is a domain error with the given code.
func IsCode(err error, code ErrorCode) bool {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr.Code == code
	}
	return false
}

// ErrNotFound is returned when a resource does not exist or is soft-deleted.
var ErrNotFound = &Error{Code: CodeNotFound, Message: "resource not found"}

// ErrUnauthorized is returned for missing or invalid credentials.
var ErrUnauthorized = &Error{Code: CodeUnauthorized, Message: "authentication required"}

// ErrForbidden is returned when the actor lacks permission or ownership.
var ErrForbidden = &Error{Code: CodeForbidden, Message: "you do not have permission to perform this action"}

// ErrConflict is returned for state conflicts such as double cancellation.
var ErrConflict = &Error{Code: CodeConflict, Message: "the request conflicts with the current state of the resource"}

// ErrSlotUnavailable is returned when a slot is booked, blocked or held.
var ErrSlotUnavailable = &Error{Code: CodeSlotUnavailable, Message: "the selected slot is no longer available"}

// ErrHoldExpired is returned when a hold is missing or has expired.
var ErrHoldExpired = &Error{Code: CodeHoldExpired, Message: "your slot hold has expired; please choose a time again"}

// ErrHoldNotOwned is returned when a hold belongs to another user.
var ErrHoldNotOwned = &Error{Code: CodeHoldNotOwned, Message: "the slot hold does not belong to you"}

// ErrRateLimited is returned when a caller exceeds a rate limit.
var ErrRateLimited = &Error{Code: CodeRateLimited, Message: "too many requests; please slow down"}

// ErrInternal is returned for unexpected failures.
var ErrInternal = &Error{Code: CodeInternal, Message: "an unexpected error occurred"}
