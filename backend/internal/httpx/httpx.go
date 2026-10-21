// Package httpx centralises JSON responses and error mapping.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/requestctx"
)

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      domain.ErrorCode `json:"code"`
	Message   string           `json:"message"`
	Details   map[string]any   `json:"details,omitempty"`
	RequestID string           `json:"request_id,omitempty"`
}

// StatusFor maps a stable error code to an HTTP status.
func StatusFor(code domain.ErrorCode) int {
	switch code {
	case domain.CodeValidation:
		return http.StatusBadRequest
	case domain.CodeUnauthorized:
		return http.StatusUnauthorized
	case domain.CodeForbidden:
		return http.StatusForbidden
	case domain.CodeNotFound:
		return http.StatusNotFound
	case domain.CodeConflict, domain.CodeSlotUnavailable, domain.CodeHoldExpired, domain.CodeHoldNotOwned:
		return http.StatusConflict
	case domain.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// WriteError writes the standard error envelope. Internal causes are logged,
// never returned to the client.
func WriteError(c *gin.Context, log *zap.Logger, err error) {
	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		log.Error("unhandled request error",
			zap.String("request_id", requestctx.RequestID(c)),
			zap.String("path", c.FullPath()),
			zap.Error(err),
		)
		domainErr = domain.ErrInternal
	} else if domainErr.Cause != nil {
		log.Error("request failed",
			zap.String("request_id", requestctx.RequestID(c)),
			zap.String("path", c.FullPath()),
			zap.String("code", string(domainErr.Code)),
			zap.Error(domainErr.Cause),
		)
	}

	c.AbortWithStatusJSON(StatusFor(domainErr.Code), errorEnvelope{Error: errorBody{
		Code:      domainErr.Code,
		Message:   domainErr.Message,
		Details:   domainErr.Details,
		RequestID: requestctx.RequestID(c),
	}})
}

// BadRequest writes a validation error.
func BadRequest(c *gin.Context, log *zap.Logger, message string, details map[string]any) {
	WriteError(c, log, domain.Validation(message, details))
}

// BindBody decodes and validates a JSON body. It writes a 400 response and
// returns false when binding fails.
func BindBody(c *gin.Context, log *zap.Logger, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		WriteError(c, log, bindError(err))
		return false
	}
	return true
}

func bindError(err error) error {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		details := make(map[string]any, len(validationErrs))
		for _, fieldErr := range validationErrs {
			field := toSnakeCase(fieldErr.Field())
			details[field] = validationMessage(fieldErr)
		}
		return domain.Validation("request validation failed", details)
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		field := toSnakeCase(typeErr.Field)
		return domain.Validation("request validation failed", map[string]any{field: "has the wrong type"})
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return domain.Validation("malformed JSON body", nil)
	}

	if errors.Is(err, io.EOF) {
		return domain.Validation("request body is required", nil)
	}

	return domain.Validation("invalid request body", nil)
}

func validationMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "uuid", "uuid4":
		return "must be a valid UUID"
	case "min":
		return "must be at least " + fieldErr.Param()
	case "max":
		return "must be at most " + fieldErr.Param()
	case "gte":
		return "must be greater than or equal to " + fieldErr.Param()
	case "lte":
		return "must be less than or equal to " + fieldErr.Param()
	case "oneof":
		return "must be one of: " + fieldErr.Param()
	case "len":
		return "must be exactly " + fieldErr.Param() + " characters"
	default:
		return "is invalid"
	}
}

func toSnakeCase(value string) string {
	var b strings.Builder
	for i, r := range value {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Data writes {"data": ...}.
func Data(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// OK writes 200 with a data envelope.
func OK(c *gin.Context, data any) { Data(c, http.StatusOK, data) }

// Created writes 201 with a data envelope.
func Created(c *gin.Context, data any) { Data(c, http.StatusCreated, data) }

// NoContent writes 204.
func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

// Pagination is the metadata returned alongside list responses.
type Pagination struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}

// Page writes {"data": [...], "pagination": {...}}.
func Page(c *gin.Context, items any, pagination Pagination) {
	c.JSON(http.StatusOK, gin.H{"data": items, "pagination": pagination})
}
