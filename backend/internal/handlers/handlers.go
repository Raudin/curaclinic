package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"curaclinic/backend/internal/config"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/requestctx"
	"curaclinic/backend/internal/service"
)

// Handlers bundles the HTTP adapters for every service.
type Handlers struct {
	Config       *config.Config
	Log          *zap.Logger
	Auth         *service.AuthService
	RBAC         *service.RBACService
	Users        *service.UserService
	Doctors      *service.DoctorService
	Availability *service.AvailabilityService
	Holds        *service.HoldService
	Appointments *service.AppointmentService
	Audit        *service.AuditService
}

// New creates the handler bundle.
func New(cfg *config.Config, log *zap.Logger, auth *service.AuthService, rbac *service.RBACService, users *service.UserService, doctors *service.DoctorService, availability *service.AvailabilityService, holds *service.HoldService, appointments *service.AppointmentService, audit *service.AuditService) *Handlers {
	return &Handlers{
		Config:       cfg,
		Log:          log,
		Auth:         auth,
		RBAC:         rbac,
		Users:        users,
		Doctors:      doctors,
		Availability: availability,
		Holds:        holds,
		Appointments: appointments,
		Audit:        audit,
	}
}

func actorOf(c *gin.Context) *domain.Actor {
	return requestctx.Actor(c)
}

func parseUUIDParam(c *gin.Context, log *zap.Logger, name string) (uuid.UUID, bool) {
	value, err := uuid.Parse(c.Param(name))
	if err != nil {
		httpx.BadRequest(c, log, "invalid path parameter", map[string]any{name: "must be a valid UUID"})
		return uuid.Nil, false
	}
	return value, true
}

func parseUUIDValue(c *gin.Context, log *zap.Logger, field, value string) (uuid.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		httpx.BadRequest(c, log, "request validation failed", map[string]any{field: "must be a valid UUID"})
		return uuid.Nil, false
	}
	return id, true
}

func parseUUIDList(c *gin.Context, log *zap.Logger, field string, values []string) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := uuid.Parse(strings.TrimSpace(value))
		if err != nil {
			httpx.BadRequest(c, log, "request validation failed", map[string]any{field: "must contain valid UUIDs"})
			return nil, false
		}
		ids = append(ids, id)
	}
	return ids, true
}

func parsePagination(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func parseOptionalTimeQuery(c *gin.Context, name string) (*time.Time, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, domain.ValidationField(name, "must be an RFC3339 timestamp")
	}
	value = value.UTC()
	return &value, nil
}

func parseClock(value string) (int, error) {
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, domain.ValidationField("interval", "times must use the HH:MM format")
	}
	return parsed.Hour()*60 + parsed.Minute(), nil
}

func parseOptionalUUIDQuery(c *gin.Context, log *zap.Logger, name string) (*uuid.UUID, bool) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, true
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		httpx.BadRequest(c, log, "request validation failed", map[string]any{name: "must be a valid UUID"})
		return nil, false
	}
	return &id, true
}

func noContent(c *gin.Context) { c.Status(http.StatusNoContent) }
