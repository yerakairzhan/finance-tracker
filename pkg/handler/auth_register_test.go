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

type authServiceSpy struct {
	registerFn    func(ctx context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error)
	registerCalls int
	gotRegister   models.RegisterRequest
	loginFn       func(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error)
	loginCalls    int
	gotLogin      models.LoginRequest
}

func (s *authServiceSpy) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
	s.registerCalls++
	s.gotRegister = req
	return s.registerFn(ctx, req)
}

func (s *authServiceSpy) Login(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
	s.loginCalls++
	s.gotLogin = req
	return s.loginFn(ctx, req)
}

func (s *authServiceSpy) Refresh(ctx context.Context, rawRefreshToken string) (*models.AuthTokens, *apperror.Error) {
	panic("unexpected call: Refresh")
}

func (s *authServiceSpy) Logout(ctx context.Context, userID int64, rawRefreshToken string) *apperror.Error {
	panic("unexpected call: Logout")
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200 and tokens", func(t *testing.T) {
		// Arrange
		service := &authServiceSpy{
			loginFn: func(_ context.Context, _ models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
				return &models.AuthTokens{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresIn:    900,
				}, nil
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/login", (&AuthHandler{authService: service}).Login)
		reqBody := `{"email":"john@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, rec.Code, rec.Body.String())
		}
		if service.loginCalls != 1 {
			t.Fatalf("expected Login to be called once, got %d", service.loginCalls)
		}
		if service.gotLogin.Email != "john@example.com" || service.gotLogin.Password != "password123" {
			t.Fatalf("unexpected Login request: %+v", service.gotLogin)
		}

		var got models.AuthTokens
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response JSON: %v", err)
		}
		if got.AccessToken != "access-token" || got.RefreshToken != "refresh-token" || got.ExpiresIn != 900 {
			t.Fatalf("unexpected response body: %+v", got)
		}
	})

	t.Run("invalid payload returns 400 and does not call service", func(t *testing.T) {
		// Arrange
		service := &authServiceSpy{
			loginFn: func(_ context.Context, _ models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
				t.Fatal("Login must not be called for invalid payload")
				return nil, nil
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/login", (&AuthHandler{authService: service}).Login)
		reqBody := `{"email":"invalid-email","password":"short"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d; body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
		if service.loginCalls != 0 {
			t.Fatalf("expected Login not to be called, got %d calls", service.loginCalls)
		}
	})

	t.Run("service unauthorized returns 401", func(t *testing.T) {
		// Arrange
		service := &authServiceSpy{
			loginFn: func(_ context.Context, _ models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Unauthorized("invalid credentials")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/login", (&AuthHandler{authService: service}).Login)
		reqBody := `{"email":"john@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d; body=%s", http.StatusUnauthorized, rec.Code, rec.Body.String())
		}
		if service.loginCalls != 1 {
			t.Fatalf("expected Login to be called once, got %d", service.loginCalls)
		}
	})

	t.Run("database is broken and service returns 500", func(t *testing.T) {
		// Arrange
		service := &authServiceSpy{
			loginFn: func(_ context.Context, _ models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Internal("database unavailable")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/login", (&AuthHandler{authService: service}).Login)
		reqBody := `{"email":"john@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d; body=%s", http.StatusInternalServerError, rec.Code, rec.Body.String())
		}
		if service.loginCalls != 1 {
			t.Fatalf("expected Login to be called once, got %d", service.loginCalls)
		}

		var got struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if got.Error.Code != "INTERNAL_ERROR" {
			t.Fatalf("expected error code INTERNAL_ERROR for login, got %q", got.Error.Code)
		}
		if got.Error.Message != "database unavailable" {
			t.Fatalf("unexpected error message: got %q", got.Error.Message)
		}
	})
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 201 and tokens", func(t *testing.T) {
		// Arrange
		service := &authServiceSpy{
			registerFn: func(_ context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
				return &models.AuthTokens{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresIn:    900,
				}, nil
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/register", (&AuthHandler{authService: service}).Register)
		reqBody := `{"email":"john@example.com","password":"password123","name":"John","currency":"USD"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d; body=%s", http.StatusCreated, rec.Code, rec.Body.String())
		}
		if service.registerCalls != 1 {
			t.Fatalf("expected Register to be called once, got %d", service.registerCalls)
		}
		if service.gotRegister.Email != "john@example.com" || service.gotRegister.Name != "John" || service.gotRegister.Currency != "USD" {
			t.Fatalf("unexpected Register request: %+v", service.gotRegister)
		}
		if service.gotRegister.Password != "password123" {
			t.Fatalf("unexpected password in Register request")
		}

		var got models.AuthTokens
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response JSON: %v", err)
		}
		if got.AccessToken != "access-token" || got.RefreshToken != "refresh-token" || got.ExpiresIn != 900 {
			t.Fatalf("unexpected response body: %+v", got)
		}
	})

	t.Run("invalid payload returns 400 and does not call service", func(t *testing.T) {
		// Arrange
		service := &authServiceSpy{
			registerFn: func(_ context.Context, _ models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
				t.Fatal("Register must not be called for invalid payload")
				return nil, nil
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/register", (&AuthHandler{authService: service}).Register)
		reqBody := `{"email":"invalid-email","password":"short","name":"","currency":"usd"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d; body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
		if service.registerCalls != 0 {
			t.Fatalf("expected Register not to be called, got %d calls", service.registerCalls)
		}

		var got struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if got.Error.Code != "VALIDATION_ERROR" {
			t.Fatalf("expected error code VALIDATION_ERROR, got %q", got.Error.Code)
		}
		if got.Error.Message == "" {
			t.Fatalf("expected non-empty validation message")
		}
	})

	t.Run("service conflict returns 409", func(t *testing.T) {
		// Arrange
		service := &authServiceSpy{
			registerFn: func(_ context.Context, _ models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Conflict("email already exists")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/register", (&AuthHandler{authService: service}).Register)
		reqBody := `{"email":"john@example.com","password":"password123","name":"John","currency":"USD"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status %d, got %d; body=%s", http.StatusConflict, rec.Code, rec.Body.String())
		}
		if service.registerCalls != 1 {
			t.Fatalf("expected Register to be called once, got %d", service.registerCalls)
		}

		var got struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if got.Error.Code != "CONFLICT" {
			t.Fatalf("expected error code CONFLICT, got %q", got.Error.Code)
		}
		if got.Error.Message != "email already exists" {
			t.Fatalf("unexpected error message: got %q", got.Error.Message)
		}
	})

	t.Run("database is broken and service returns 500", func(t *testing.T) {
		// Arrange
		service := &authServiceSpy{
			registerFn: func(_ context.Context, _ models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
				return nil, apperror.Internal("database unavailable")
			},
		}
		router := gin.New()
		router.POST("/api/v1/auth/register", (&AuthHandler{authService: service}).Register)
		reqBody := `{"email":"john@example.com","password":"password123","name":"John","currency":"USD"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d; body=%s", http.StatusInternalServerError, rec.Code, rec.Body.String())
		}
		if service.registerCalls != 1 {
			t.Fatalf("expected Register to be called once, got %d", service.registerCalls)
		}

		var got struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if got.Error.Code != "INTERNAL_ERROR" {
			t.Fatalf("expected error code INTERNAL_ERROR, got %q", got.Error.Code)
		}
		if got.Error.Message != "database unavailable" {
			t.Fatalf("unexpected error message: got %q", got.Error.Message)
		}
	})
}
