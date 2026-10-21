package httpx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/httpx"
)

func init() { gin.SetMode(gin.TestMode) }

func perform(t *testing.T, body string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	engine := gin.New()
	engine.POST("/", handler)
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

type errorEnvelope struct {
	Error struct {
		Code      string         `json:"code"`
		Message   string         `json:"message"`
		Details   map[string]any `json:"details"`
		RequestID string         `json:"request_id"`
	} `json:"error"`
}

func TestStatusFor(t *testing.T) {
	require.Equal(t, http.StatusBadRequest, httpx.StatusFor(domain.CodeValidation))
	require.Equal(t, http.StatusUnauthorized, httpx.StatusFor(domain.CodeUnauthorized))
	require.Equal(t, http.StatusForbidden, httpx.StatusFor(domain.CodeForbidden))
	require.Equal(t, http.StatusNotFound, httpx.StatusFor(domain.CodeNotFound))
	require.Equal(t, http.StatusConflict, httpx.StatusFor(domain.CodeSlotUnavailable))
	require.Equal(t, http.StatusTooManyRequests, httpx.StatusFor(domain.CodeRateLimited))
	require.Equal(t, http.StatusInternalServerError, httpx.StatusFor(domain.CodeInternal))
}

func TestBindBodyReturnsFieldErrors(t *testing.T) {
	recorder := perform(t, `{"email":"not-an-email"}`, func(c *gin.Context) {
		var request struct {
			Email string `json:"email" binding:"required,email"`
		}
		if !httpx.BindBody(c, zap.NewNop(), &request) {
			return
		}
		c.Status(http.StatusOK)
	})

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var envelope errorEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, string(domain.CodeValidation), envelope.Error.Code)
	require.Contains(t, envelope.Error.Details, "email")
}

func TestWriteErrorDoesNotLeakInternalCause(t *testing.T) {
	recorder := perform(t, `{}`, func(c *gin.Context) {
		httpx.WriteError(c, zap.NewNop(), domain.ErrInternal.WithCause(errors.New("database password leaked")))
	})

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "database password leaked")
	var envelope errorEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, string(domain.CodeInternal), envelope.Error.Code)
}

func TestBindBodyReportsJSONFieldPathsAndUnits(t *testing.T) {
	type interval struct {
		StartTime string `json:"start_time" binding:"required"`
	}
	type request struct {
		Password    string     `json:"password" binding:"required,min=12"`
		DateOfBirth string     `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
		RoleIDs     []string   `json:"role_ids" binding:"max=1,dive,uuid"`
		Intervals   []interval `json:"intervals" binding:"required,dive"`
	}

	body := `{"password":"short","date_of_birth":"31/01/1990","role_ids":["x"],"intervals":[{"start_time":""}]}`
	recorder := perform(t, body, func(c *gin.Context) {
		if !httpx.BindBody(c, zap.NewNop(), &request{}) {
			return
		}
		c.Status(http.StatusOK)
	})

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var envelope errorEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, string(domain.CodeValidation), envelope.Error.Code)
	require.Equal(t, "must be at least 12 characters", envelope.Error.Details["password"])
	require.Equal(t, "must be a date in YYYY-MM-DD format", envelope.Error.Details["date_of_birth"])
	require.Equal(t, "must be a valid UUID", envelope.Error.Details["role_ids[0]"])
	require.Equal(t, "is required", envelope.Error.Details["intervals[0].start_time"])
}

func TestBindBodyRejectsBadTimestampsAndMalformedJSON(t *testing.T) {
	type request struct {
		StartTime *time.Time `json:"start_time" binding:"required"`
	}
	handler := func(c *gin.Context) {
		if !httpx.BindBody(c, zap.NewNop(), &request{}) {
			return
		}
		c.Status(http.StatusOK)
	}

	recorder := perform(t, `{"start_time":"tomorrow"}`, handler)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "RFC3339")

	recorder = perform(t, `{"start_time":`, handler)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), string(domain.CodeValidation))

	recorder = perform(t, ``, handler)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "request body is required")
}
