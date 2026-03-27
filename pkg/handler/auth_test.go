package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"

	"github.com/gin-gonic/gin"
)

type authHandlerServiceSpy struct {
	registerFn    func(ctx context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error)
	registerCalls int
	gotRegister   models.RegisterRequest

	loginFn    func(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error)
	loginCalls int
	gotLogin   models.LoginRequest

	refreshFn    func(ctx context.Context, rawRefreshToken string) (*models.AuthTokens, *apperror.Error)
	refreshCalls int
	gotRefresh   string

	logoutFn    func(ctx context.Context, userID int64, rawRefreshToken, rawAccessToken string) *apperror.Error
	logoutCalls int
	gotLogout   struct {
		userID       int64
		refreshToken string
		accessToken  string
	}
}

func (s *authHandlerServiceSpy) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
	s.registerCalls++
	s.gotRegister = req
	if s.registerFn == nil {
		panic("unexpected call: Register")
	}
	return s.registerFn(ctx, req)
}

func (s *authHandlerServiceSpy) Login(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
	s.loginCalls++
	s.gotLogin = req
	if s.loginFn == nil {
		panic("unexpected call: Login")
	}
	return s.loginFn(ctx, req)
}

func (s *authHandlerServiceSpy) Refresh(ctx context.Context, rawRefreshToken string) (*models.AuthTokens, *apperror.Error) {
	s.refreshCalls++
	s.gotRefresh = rawRefreshToken
	if s.refreshFn == nil {
		panic("unexpected call: Refresh")
	}
	return s.refreshFn(ctx, rawRefreshToken)
}

func (s *authHandlerServiceSpy) Logout(ctx context.Context, userID int64, rawRefreshToken, rawAccessToken string) *apperror.Error {
	s.logoutCalls++
	s.gotLogout.userID = userID
	s.gotLogout.refreshToken = rawRefreshToken
	s.gotLogout.accessToken = rawAccessToken
	if s.logoutFn == nil {
		panic("unexpected call: Logout")
	}
	return s.logoutFn(ctx, userID, rawRefreshToken, rawAccessToken)
}

func TestNewAuthHandler(t *testing.T) {
	// Arrange
	handler := NewAuthHandler(nil)

	// Act
	got := handler

	// Assert
	if got == nil {
		t.Fatal("expected non-nil handler")
	}
	if got.authService != nil {
		t.Fatal("expected auth service to be nil")
	}
}

func TestAuthHandler_Register_Extended(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 201 and tokens", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			registerFn: func(_ context.Context, _ models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
				return &models.AuthTokens{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token-123456789012345678901234",
					ExpiresIn:    900,
				}, nil
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/register", (&AuthHandler{authService: service}).Register)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/register", `{"email":"john@example.com","password":"password123","name":"John","currency":"USD"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusCreated)
		if service.registerCalls != 1 {
			t.Fatalf("expected Register to be called once, got %d", service.registerCalls)
		}
		if service.gotRegister.Email != "john@example.com" || service.gotRegister.Password != "password123" {
			t.Fatalf("unexpected register request: %+v", service.gotRegister)
		}
		if service.gotRegister.Name != "John" || service.gotRegister.Currency != "USD" {
			t.Fatalf("unexpected register request: %+v", service.gotRegister)
		}

		var got models.AuthTokens
		decodeJSON(t, rec, &got)
		if got.AccessToken != "access-token" || got.RefreshToken != "refresh-token-123456789012345678901234" || got.ExpiresIn != 900 {
			t.Fatalf("unexpected response body: %+v", got)
		}
	})

	t.Run("invalid payload returns 400 and does not call service", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{}
		router := gin.New()
		router.POST("/api/v1/auth/register", (&AuthHandler{authService: service}).Register)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/register", `{"email":"invalid-email","password":"short","name":"","currency":"usd"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusBadRequest)
		if service.registerCalls != 0 {
			t.Fatalf("expected Register not to be called, got %d calls", service.registerCalls)
		}

		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})

	t.Run("service conflict returns 409", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			registerFn: func(_ context.Context, _ models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Conflict("email already exists")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/register", (&AuthHandler{authService: service}).Register)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/register", `{"email":"john@example.com","password":"password123","name":"John","currency":"USD"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusConflict)
		if service.registerCalls != 1 {
			t.Fatalf("expected Register to be called once, got %d", service.registerCalls)
		}
		assertErrorEnvelope(t, rec, "CONFLICT", "email already exists")
	})

	t.Run("service internal error returns 500", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			registerFn: func(_ context.Context, _ models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Internal("database unavailable")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/register", (&AuthHandler{authService: service}).Register)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/register", `{"email":"john@example.com","password":"password123","name":"John","currency":"USD"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusInternalServerError)
		if service.registerCalls != 1 {
			t.Fatalf("expected Register to be called once, got %d", service.registerCalls)
		}
		assertErrorEnvelope(t, rec, "INTERNAL_ERROR", "database unavailable")
	})
}

func TestAuthHandler_Login_Extended(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200 and tokens", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			loginFn: func(_ context.Context, _ models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
				return &models.AuthTokens{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token-123456789012345678901234",
					ExpiresIn:    900,
				}, nil
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/login", (&AuthHandler{authService: service}).Login)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{"email":"john@example.com","password":"password123"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusOK)
		if service.loginCalls != 1 {
			t.Fatalf("expected Login to be called once, got %d", service.loginCalls)
		}
		if service.gotLogin.Email != "john@example.com" || service.gotLogin.Password != "password123" {
			t.Fatalf("unexpected login request: %+v", service.gotLogin)
		}

		var got models.AuthTokens
		decodeJSON(t, rec, &got)
		if got.AccessToken != "access-token" || got.RefreshToken != "refresh-token-123456789012345678901234" || got.ExpiresIn != 900 {
			t.Fatalf("unexpected response body: %+v", got)
		}
	})

	t.Run("invalid payload returns 400 and does not call service", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{}
		router := gin.New()
		router.POST("/api/v1/auth/login", (&AuthHandler{authService: service}).Login)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{"email":"invalid-email","password":"short"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusBadRequest)
		if service.loginCalls != 0 {
			t.Fatalf("expected Login not to be called, got %d calls", service.loginCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})

	t.Run("service unauthorized returns 401", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			loginFn: func(_ context.Context, _ models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Unauthorized("invalid credentials")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/login", (&AuthHandler{authService: service}).Login)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{"email":"john@example.com","password":"password123"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusUnauthorized)
		if service.loginCalls != 1 {
			t.Fatalf("expected Login to be called once, got %d", service.loginCalls)
		}
		assertErrorEnvelope(t, rec, "UNAUTHORIZED", "invalid credentials")
	})

	t.Run("service internal error returns 500", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			loginFn: func(_ context.Context, _ models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Internal("database unavailable")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/login", (&AuthHandler{authService: service}).Login)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{"email":"john@example.com","password":"password123"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusInternalServerError)
		if service.loginCalls != 1 {
			t.Fatalf("expected Login to be called once, got %d", service.loginCalls)
		}
		assertErrorEnvelope(t, rec, "INTERNAL_ERROR", "database unavailable")
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200 and rotated tokens", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			refreshFn: func(_ context.Context, rawRefreshToken string) (*models.AuthTokens, *apperror.Error) {
				if rawRefreshToken != "12345678901234567890123456789012" {
					t.Fatalf("unexpected refresh token: %q", rawRefreshToken)
				}
				return &models.AuthTokens{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token-1234567890123456",
					ExpiresIn:    900,
				}, nil
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/refresh", (&AuthHandler{authService: service}).Refresh)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"12345678901234567890123456789012"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusOK)
		if service.refreshCalls != 1 {
			t.Fatalf("expected Refresh to be called once, got %d", service.refreshCalls)
		}
		if service.gotRefresh != "12345678901234567890123456789012" {
			t.Fatalf("unexpected refresh token: %q", service.gotRefresh)
		}

		var got models.AuthTokens
		decodeJSON(t, rec, &got)
		if got.AccessToken != "new-access-token" || got.RefreshToken != "new-refresh-token-1234567890123456" || got.ExpiresIn != 900 {
			t.Fatalf("unexpected response body: %+v", got)
		}
	})

	t.Run("invalid payload returns 400 and does not call service", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{}
		router := gin.New()
		router.POST("/api/v1/auth/refresh", (&AuthHandler{authService: service}).Refresh)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"too-short"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusBadRequest)
		if service.refreshCalls != 0 {
			t.Fatalf("expected Refresh not to be called, got %d calls", service.refreshCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})

	t.Run("service unauthorized returns 401", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			refreshFn: func(_ context.Context, _ string) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Unauthorized("invalid refresh token")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/refresh", (&AuthHandler{authService: service}).Refresh)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"12345678901234567890123456789012"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusUnauthorized)
		if service.refreshCalls != 1 {
			t.Fatalf("expected Refresh to be called once, got %d", service.refreshCalls)
		}
		assertErrorEnvelope(t, rec, "UNAUTHORIZED", "invalid refresh token")
	})

	t.Run("service internal error returns 500", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			refreshFn: func(_ context.Context, _ string) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Internal("failed to rotate refresh token")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/refresh", (&AuthHandler{authService: service}).Refresh)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"12345678901234567890123456789012"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusInternalServerError)
		if service.refreshCalls != 1 {
			t.Fatalf("expected Refresh to be called once, got %d", service.refreshCalls)
		}
		assertErrorEnvelope(t, rec, "INTERNAL_ERROR", "failed to rotate refresh token")
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 204", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			logoutFn: func(_ context.Context, userID int64, rawRefreshToken, rawAccessToken string) *apperror.Error {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				if rawRefreshToken != "12345678901234567890123456789012" {
					t.Fatalf("unexpected refresh token: %q", rawRefreshToken)
				}
				if rawAccessToken != "access-token-123" {
					t.Fatalf("unexpected access token: %q", rawAccessToken)
				}
				return nil
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/logout", func(c *gin.Context) {
			c.Set("auth_user_id", int64(42))
			(&AuthHandler{authService: service}).Logout(c)
		})
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":"12345678901234567890123456789012"}`)
		req.Header.Set("Authorization", "Bearer access-token-123")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusNoContent)
		if service.logoutCalls != 1 {
			t.Fatalf("expected Logout to be called once, got %d", service.logoutCalls)
		}
		if service.gotLogout.userID != 42 || service.gotLogout.refreshToken != "12345678901234567890123456789012" || service.gotLogout.accessToken != "access-token-123" {
			t.Fatalf("unexpected logout call: %+v", service.gotLogout)
		}
		if strings.TrimSpace(rec.Body.String()) != "" {
			t.Fatalf("expected empty response body, got %q", rec.Body.String())
		}
	})

	t.Run("invalid payload returns 400 and does not call service", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{}
		router := gin.New()
		router.POST("/api/v1/auth/logout", func(c *gin.Context) {
			c.Set("auth_user_id", int64(42))
			(&AuthHandler{authService: service}).Logout(c)
		})
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":"short"}`)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusBadRequest)
		if service.logoutCalls != 0 {
			t.Fatalf("expected Logout not to be called, got %d calls", service.logoutCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})

	t.Run("missing user context returns 401 and does not call service", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{}
		router := gin.New()
		router.POST("/api/v1/auth/logout", (&AuthHandler{authService: service}).Logout)
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":"12345678901234567890123456789012"}`)
		req.Header.Set("Authorization", "Bearer access-token-123")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusUnauthorized)
		if service.logoutCalls != 0 {
			t.Fatalf("expected Logout not to be called, got %d calls", service.logoutCalls)
		}
		assertErrorEnvelope(t, rec, "UNAUTHORIZED", "invalid token context")
	})

	t.Run("service not found returns 404", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			logoutFn: func(_ context.Context, _ int64, _ string, _ string) *apperror.Error {
				return apperror.NotFound("refresh token not found")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/logout", func(c *gin.Context) {
			c.Set("auth_user_id", int64(42))
			(&AuthHandler{authService: service}).Logout(c)
		})
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":"12345678901234567890123456789012"}`)
		req.Header.Set("Authorization", "Bearer access-token-123")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusNotFound)
		if service.logoutCalls != 1 {
			t.Fatalf("expected Logout to be called once, got %d", service.logoutCalls)
		}
		assertErrorEnvelope(t, rec, "NOT_FOUND", "refresh token not found")
	})

	t.Run("service internal error returns 500", func(t *testing.T) {
		// Arrange
		service := &authHandlerServiceSpy{
			logoutFn: func(_ context.Context, _ int64, _ string, _ string) *apperror.Error {
				return apperror.Internal("failed to revoke refresh token")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/logout", func(c *gin.Context) {
			c.Set("auth_user_id", int64(42))
			(&AuthHandler{authService: service}).Logout(c)
		})
		req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":"12345678901234567890123456789012"}`)
		req.Header.Set("Authorization", "Bearer access-token-123")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assertStatus(t, rec, http.StatusInternalServerError)
		if service.logoutCalls != 1 {
			t.Fatalf("expected Logout to be called once, got %d", service.logoutCalls)
		}
		assertErrorEnvelope(t, rec, "INTERNAL_ERROR", "failed to revoke refresh token")
	})
}

func newJSONRequest(t *testing.T, method, target, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("expected status %d, got %d; body=%s", want, rec.Code, rec.Body.String())
	}
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()

	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
}

func assertErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, wantCode, wantMessage string) {
	t.Helper()

	var got struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	decodeJSON(t, rec, &got)

	if got.Error.Code != wantCode {
		t.Fatalf("expected error code %q, got %q", wantCode, got.Error.Code)
	}
	if wantMessage != "" && got.Error.Message != wantMessage {
		t.Fatalf("expected error message %q, got %q", wantMessage, got.Error.Message)
	}
	if wantMessage == "" && got.Error.Message == "" {
		t.Fatal("expected non-empty error message")
	}
}
