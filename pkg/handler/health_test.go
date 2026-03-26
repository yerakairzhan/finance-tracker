package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-tracker/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type healthServiceSpy struct {
	readyFn    func(ctx context.Context) *apperror.Error
	readyCalls int
}

func (s *healthServiceSpy) Ready(ctx context.Context) *apperror.Error {
	s.readyCalls++
	if s.readyFn == nil {
		panic("unexpected call: Ready")
	}
	return s.readyFn(ctx)
}

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler(nil)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if handler.healthService != nil {
		t.Fatal("expected health service to be nil")
	}
}

func TestHealthHandler_Live(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/health", (&HealthHandler{}).Live)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertStatus(t, rec, http.StatusOK)
	if rec.Body.String() != "{\"status\":\"ok\"}" {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestHealthHandler_Ready(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200", func(t *testing.T) {
		service := &healthServiceSpy{
			readyFn: func(_ context.Context) *apperror.Error {
				return nil
			},
		}
		router := gin.New()
		router.GET("/health/ready", (&HealthHandler{healthService: service}).Ready)
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.readyCalls != 1 {
			t.Fatalf("expected Ready to be called once, got %d", service.readyCalls)
		}
		if rec.Body.String() != "{\"status\":\"ready\"}" {
			t.Fatalf("unexpected response body: %s", rec.Body.String())
		}
	})

	t.Run("service internal error returns 500", func(t *testing.T) {
		service := &healthServiceSpy{
			readyFn: func(_ context.Context) *apperror.Error {
				return apperror.Internal("database is not ready")
			},
		}
		router := gin.New()
		router.GET("/health/ready", (&HealthHandler{healthService: service}).Ready)
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusInternalServerError)
		assertErrorEnvelope(t, rec, "INTERNAL_ERROR", "database is not ready")
	})
}
