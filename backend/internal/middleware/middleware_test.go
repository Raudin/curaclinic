package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/middleware"
	"curaclinic/backend/internal/platform/token"
	"curaclinic/backend/internal/requestctx"
)

func init() { gin.SetMode(gin.TestMode) }

func TestRequirePermission(t *testing.T) {
	actor := &domain.Actor{
		UserID:      uuid.New(),
		Permissions: map[string]struct{}{domain.PermAppointmentsCreateSelf: {}},
	}

	withActor := func(c *gin.Context) {
		requestctx.SetActor(c, actor)
		c.Next()
	}

	engine := gin.New()
	engine.Use(withActor, middleware.RequirePermission(zap.NewNop(), domain.PermAppointmentsCreateSelf))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, recorder.Code)

	denied := gin.New()
	denied.Use(withActor, middleware.RequirePermission(zap.NewNop(), domain.PermUsersManage))
	denied.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder = httptest.NewRecorder()
	denied.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

type stubUserLoader struct{ revoked bool }

func (stubUserLoader) GetActiveUserByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	return &domain.User{ID: id, Email: "patient@example.com", Status: domain.UserStatusActive}, nil
}

func (s stubUserLoader) SessionActive(_ context.Context, _ uuid.UUID) (bool, error) {
	return !s.revoked, nil
}

type stubAccessReader struct{}

func (stubAccessReader) Access(_ context.Context, _ uuid.UUID) (domain.UserAccess, error) {
	return domain.UserAccess{Roles: []string{"patient"}, Permissions: []string{domain.PermAppointmentsCreateSelf}}, nil
}

func TestAuthMiddleware(t *testing.T) {
	tokens := token.New("01234567890123456789012345678901", "curaclinic-api", "curaclinic-web", 15*time.Minute)
	userID := uuid.New()
	sessionID := uuid.New()
	accessToken, _, err := tokens.IssueAccess(userID, sessionID, time.Now().UTC())
	require.NoError(t, err)

	engine := gin.New()
	engine.Use(middleware.Auth(tokens, stubUserLoader{}, stubAccessReader{}, zap.NewNop()))
	engine.GET("/", func(c *gin.Context) {
		actor := requestctx.Actor(c)
		if actor == nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": actor.UserID.String(), "can_create": actor.Can(domain.PermAppointmentsCreateSelf)})
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), userID.String())
	require.Contains(t, recorder.Body.String(), "true")

	recorder = httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Basic abc")
	recorder = httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)

	revoked := gin.New()
	revoked.Use(middleware.Auth(tokens, stubUserLoader{revoked: true}, stubAccessReader{}, zap.NewNop()))
	revoked.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	request = httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	recorder = httptest.NewRecorder()
	revoked.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusUnauthorized, recorder.Code, "a valid JWT for a revoked session must be rejected")
}

func TestRateLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	engine := gin.New()
	engine.Use(middleware.RateLimit(client, zap.NewNop(), middleware.RateLimitConfig{
		Scope:  "test",
		Limit:  1,
		Window: time.Minute,
		Identity: func(*gin.Context) string {
			return "fixed-identity"
		},
	}))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "1", recorder.Header().Get("X-RateLimit-Limit"))

	recorder = httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	require.NotEmpty(t, recorder.Header().Get("Retry-After"))
}
