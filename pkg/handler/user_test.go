package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"

	"github.com/gin-gonic/gin"
)

type userServiceSpy struct {
	meFn     func(ctx context.Context, userID int64) (*models.User, *apperror.Error)
	meCalls  int
	meUserID int64

	updateMeFn     func(ctx context.Context, userID int64, req models.UpdateMeRequest) (*models.User, *apperror.Error)
	updateMeCalls  int
	updateMeUserID int64
	updateMeReq    models.UpdateMeRequest

	changePasswordFn     func(ctx context.Context, userID int64, req models.ChangePasswordRequest) *apperror.Error
	changePasswordCalls  int
	changePasswordUserID int64
	changePasswordReq    models.ChangePasswordRequest
}

func (s *userServiceSpy) Me(ctx context.Context, userID int64) (*models.User, *apperror.Error) {
	s.meCalls++
	s.meUserID = userID
	if s.meFn == nil {
		panic("unexpected call: Me")
	}
	return s.meFn(ctx, userID)
}

func (s *userServiceSpy) UpdateMe(ctx context.Context, userID int64, req models.UpdateMeRequest) (*models.User, *apperror.Error) {
	s.updateMeCalls++
	s.updateMeUserID = userID
	s.updateMeReq = req
	if s.updateMeFn == nil {
		panic("unexpected call: UpdateMe")
	}
	return s.updateMeFn(ctx, userID, req)
}

func (s *userServiceSpy) ChangePassword(ctx context.Context, userID int64, req models.ChangePasswordRequest) *apperror.Error {
	s.changePasswordCalls++
	s.changePasswordUserID = userID
	s.changePasswordReq = req
	if s.changePasswordFn == nil {
		panic("unexpected call: ChangePassword")
	}
	return s.changePasswordFn(ctx, userID, req)
}

func TestNewUserHandler(t *testing.T) {
	handler := NewUserHandler(nil)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if handler.userService != nil {
		t.Fatal("expected user service to be nil")
	}
}

func TestUserHandler_Me(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200", func(t *testing.T) {
		service := &userServiceSpy{
			meFn: func(_ context.Context, userID int64) (*models.User, *apperror.Error) {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				user := sampleUser()
				return &user, nil
			},
		}
		router := gin.New()
		router.GET("/users/me", withUserID(42, (&UserHandler{userService: service}).Me))
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.meCalls != 1 || service.meUserID != 42 {
			t.Fatalf("unexpected Me call state: calls=%d userID=%d", service.meCalls, service.meUserID)
		}
		if !strings.Contains(rec.Body.String(), `"email":"john@example.com"`) {
			t.Fatalf("unexpected response body: %s", rec.Body.String())
		}
	})

	t.Run("missing user context returns 401", func(t *testing.T) {
		service := &userServiceSpy{}
		router := gin.New()
		router.GET("/users/me", (&UserHandler{userService: service}).Me)
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusUnauthorized)
		if service.meCalls != 0 {
			t.Fatalf("expected Me not to be called, got %d calls", service.meCalls)
		}
		assertErrorEnvelope(t, rec, "UNAUTHORIZED", "invalid token context")
	})
}

func TestUserHandler_UpdateMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200", func(t *testing.T) {
		service := &userServiceSpy{
			updateMeFn: func(_ context.Context, userID int64, req models.UpdateMeRequest) (*models.User, *apperror.Error) {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				if req.Name == nil || *req.Name != "Jane" {
					t.Fatalf("unexpected update request: %+v", req)
				}
				user := sampleUser()
				user.Name = "Jane"
				return &user, nil
			},
		}
		router := gin.New()
		router.PATCH("/users/me", withUserID(42, (&UserHandler{userService: service}).UpdateMe))
		req := newJSONRequest(t, http.MethodPatch, "/users/me", `{"name":"Jane"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.updateMeCalls != 1 || service.updateMeUserID != 42 {
			t.Fatalf("unexpected UpdateMe call state: calls=%d userID=%d", service.updateMeCalls, service.updateMeUserID)
		}
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		service := &userServiceSpy{}
		router := gin.New()
		router.PATCH("/users/me", withUserID(42, (&UserHandler{userService: service}).UpdateMe))
		req := newJSONRequest(t, http.MethodPatch, "/users/me", `{"currency":"usd"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.updateMeCalls != 0 {
			t.Fatalf("expected UpdateMe not to be called, got %d calls", service.updateMeCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})
}

func TestUserHandler_ChangePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 204", func(t *testing.T) {
		service := &userServiceSpy{
			changePasswordFn: func(_ context.Context, userID int64, req models.ChangePasswordRequest) *apperror.Error {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				if req.CurrentPassword != "password123" || req.NewPassword != "newpassword123" {
					t.Fatalf("unexpected request: %+v", req)
				}
				return nil
			},
		}
		router := gin.New()
		router.PATCH("/users/me/password", withUserID(42, (&UserHandler{userService: service}).ChangePassword))
		req := newJSONRequest(t, http.MethodPatch, "/users/me/password", `{"current_password":"password123","new_password":"newpassword123"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusNoContent)
		if service.changePasswordCalls != 1 || service.changePasswordUserID != 42 {
			t.Fatalf("unexpected ChangePassword call state: calls=%d userID=%d", service.changePasswordCalls, service.changePasswordUserID)
		}
		if strings.TrimSpace(rec.Body.String()) != "" {
			t.Fatalf("expected empty body, got %q", rec.Body.String())
		}
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		service := &userServiceSpy{}
		router := gin.New()
		router.PATCH("/users/me/password", withUserID(42, (&UserHandler{userService: service}).ChangePassword))
		req := newJSONRequest(t, http.MethodPatch, "/users/me/password", `{"current_password":"short","new_password":"tiny"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.changePasswordCalls != 0 {
			t.Fatalf("expected ChangePassword not to be called, got %d calls", service.changePasswordCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})

	t.Run("service unauthorized returns 401", func(t *testing.T) {
		service := &userServiceSpy{
			changePasswordFn: func(_ context.Context, _ int64, _ models.ChangePasswordRequest) *apperror.Error {
				return apperror.Unauthorized("current password is incorrect")
			},
		}
		router := gin.New()
		router.PATCH("/users/me/password", withUserID(42, (&UserHandler{userService: service}).ChangePassword))
		req := newJSONRequest(t, http.MethodPatch, "/users/me/password", `{"current_password":"password123","new_password":"newpassword123"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusUnauthorized)
		assertErrorEnvelope(t, rec, "UNAUTHORIZED", "current password is incorrect")
	})
}

func sampleUser() models.User {
	now := time.Unix(1700000000, 0).UTC()
	return models.User{
		ID:        42,
		Email:     "john@example.com",
		Name:      "John",
		Currency:  "USD",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
