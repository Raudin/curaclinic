package httpx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
