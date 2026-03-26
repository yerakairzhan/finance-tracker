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

type accountServiceSpy struct {
	listFn     func(ctx context.Context, userID int64) ([]models.Account, *apperror.Error)
	listCalls  int
	listUserID int64

	createFn     func(ctx context.Context, userID int64, req models.CreateAccountRequest) (*models.Account, *apperror.Error)
	createCalls  int
	createUserID int64
	createReq    models.CreateAccountRequest

	getByIDFn     func(ctx context.Context, userID, accountID int64) (*models.Account, *apperror.Error)
	getByIDCalls  int
	getByIDUserID int64
	getByIDID     int64

	updateFn     func(ctx context.Context, userID, accountID int64, req models.UpdateAccountRequest) (*models.Account, *apperror.Error)
	updateCalls  int
	updateUserID int64
	updateID     int64
	updateReq    models.UpdateAccountRequest

	deleteFn     func(ctx context.Context, userID, accountID int64) *apperror.Error
	deleteCalls  int
	deleteUserID int64
	deleteID     int64
}

func (s *accountServiceSpy) List(ctx context.Context, userID int64) ([]models.Account, *apperror.Error) {
	s.listCalls++
	s.listUserID = userID
	if s.listFn == nil {
		panic("unexpected call: List")
	}
	return s.listFn(ctx, userID)
}

func (s *accountServiceSpy) Create(ctx context.Context, userID int64, req models.CreateAccountRequest) (*models.Account, *apperror.Error) {
	s.createCalls++
	s.createUserID = userID
	s.createReq = req
	if s.createFn == nil {
		panic("unexpected call: Create")
	}
	return s.createFn(ctx, userID, req)
}

func (s *accountServiceSpy) GetByID(ctx context.Context, userID, accountID int64) (*models.Account, *apperror.Error) {
	s.getByIDCalls++
	s.getByIDUserID = userID
	s.getByIDID = accountID
	if s.getByIDFn == nil {
		panic("unexpected call: GetByID")
	}
	return s.getByIDFn(ctx, userID, accountID)
}

func (s *accountServiceSpy) Update(ctx context.Context, userID, accountID int64, req models.UpdateAccountRequest) (*models.Account, *apperror.Error) {
	s.updateCalls++
	s.updateUserID = userID
	s.updateID = accountID
	s.updateReq = req
	if s.updateFn == nil {
		panic("unexpected call: Update")
	}
	return s.updateFn(ctx, userID, accountID, req)
}

func (s *accountServiceSpy) Delete(ctx context.Context, userID, accountID int64) *apperror.Error {
	s.deleteCalls++
	s.deleteUserID = userID
	s.deleteID = accountID
	if s.deleteFn == nil {
		panic("unexpected call: Delete")
	}
	return s.deleteFn(ctx, userID, accountID)
}

func TestNewAccountHandler(t *testing.T) {
	handler := NewAccountHandler(nil)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if handler.accountService != nil {
		t.Fatal("expected account service to be nil")
	}
}

func TestAccountHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200 and accounts", func(t *testing.T) {
		service := &accountServiceSpy{
			listFn: func(_ context.Context, userID int64) ([]models.Account, *apperror.Error) {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				return []models.Account{sampleAccount()}, nil
			},
		}
		router := gin.New()
		router.GET("/accounts", withUserID(42, (&AccountHandler{accountService: service}).List))
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.listCalls != 1 || service.listUserID != 42 {
			t.Fatalf("unexpected list call state: calls=%d userID=%d", service.listCalls, service.listUserID)
		}
		if !strings.Contains(rec.Body.String(), `"name":"Cash"`) {
			t.Fatalf("unexpected response body: %s", rec.Body.String())
		}
	})

	t.Run("missing user context returns 401", func(t *testing.T) {
		service := &accountServiceSpy{}
		router := gin.New()
		router.GET("/accounts", (&AccountHandler{accountService: service}).List)
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusUnauthorized)
		if service.listCalls != 0 {
			t.Fatalf("expected List not to be called, got %d calls", service.listCalls)
		}
		assertErrorEnvelope(t, rec, "UNAUTHORIZED", "invalid token context")
	})
}

func TestAccountHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 201", func(t *testing.T) {
		service := &accountServiceSpy{
			createFn: func(_ context.Context, userID int64, req models.CreateAccountRequest) (*models.Account, *apperror.Error) {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				if req.Name != "Cash" || req.AccountType != "cash" || req.Currency != "USD" || req.Balance != "100.50" {
					t.Fatalf("unexpected create request: %+v", req)
				}
				account := sampleAccount()
				return &account, nil
			},
		}
		router := gin.New()
		router.POST("/accounts", withUserID(42, (&AccountHandler{accountService: service}).Create))
		req := newJSONRequest(t, http.MethodPost, "/accounts", `{"name":"Cash","account_type":"cash","currency":"USD","balance":"100.50"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusCreated)
		if service.createCalls != 1 || service.createUserID != 42 {
			t.Fatalf("unexpected create call state: calls=%d userID=%d", service.createCalls, service.createUserID)
		}
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		service := &accountServiceSpy{}
		router := gin.New()
		router.POST("/accounts", withUserID(42, (&AccountHandler{accountService: service}).Create))
		req := newJSONRequest(t, http.MethodPost, "/accounts", `{"name":"","account_type":"card","currency":"usd","balance":""}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.createCalls != 0 {
			t.Fatalf("expected Create not to be called, got %d calls", service.createCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})

	t.Run("service returns internal error", func(t *testing.T) {
		service := &accountServiceSpy{
			createFn: func(_ context.Context, _ int64, _ models.CreateAccountRequest) (*models.Account, *apperror.Error) {
				return nil, apperror.Internal("failed to create account")
			},
		}
		router := gin.New()
		router.POST("/accounts", withUserID(42, (&AccountHandler{accountService: service}).Create))
		req := newJSONRequest(t, http.MethodPost, "/accounts", `{"name":"Cash","account_type":"cash","currency":"USD","balance":"100.50"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusInternalServerError)
		assertErrorEnvelope(t, rec, "INTERNAL_ERROR", "failed to create account")
	})
}

func TestAccountHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200", func(t *testing.T) {
		service := &accountServiceSpy{
			getByIDFn: func(_ context.Context, userID, accountID int64) (*models.Account, *apperror.Error) {
				if userID != 42 || accountID != 7 {
					t.Fatalf("unexpected ids: user=%d account=%d", userID, accountID)
				}
				account := sampleAccount()
				return &account, nil
			},
		}
		router := gin.New()
		router.GET("/accounts/:id", withUserID(42, (&AccountHandler{accountService: service}).GetByID))
		req := httptest.NewRequest(http.MethodGet, "/accounts/7", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.getByIDCalls != 1 {
			t.Fatalf("expected GetByID to be called once, got %d", service.getByIDCalls)
		}
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		service := &accountServiceSpy{}
		router := gin.New()
		router.GET("/accounts/:id", withUserID(42, (&AccountHandler{accountService: service}).GetByID))
		req := httptest.NewRequest(http.MethodGet, "/accounts/abc", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.getByIDCalls != 0 {
			t.Fatalf("expected GetByID not to be called, got %d calls", service.getByIDCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "invalid account id")
	})

	t.Run("service not found returns 404", func(t *testing.T) {
		service := &accountServiceSpy{
			getByIDFn: func(_ context.Context, _, _ int64) (*models.Account, *apperror.Error) {
				return nil, apperror.NotFound("account not found")
			},
		}
		router := gin.New()
		router.GET("/accounts/:id", withUserID(42, (&AccountHandler{accountService: service}).GetByID))
		req := httptest.NewRequest(http.MethodGet, "/accounts/7", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusNotFound)
		assertErrorEnvelope(t, rec, "NOT_FOUND", "account not found")
	})
}

func TestAccountHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200", func(t *testing.T) {
		service := &accountServiceSpy{
			updateFn: func(_ context.Context, userID, accountID int64, req models.UpdateAccountRequest) (*models.Account, *apperror.Error) {
				if userID != 42 || accountID != 7 {
					t.Fatalf("unexpected ids: user=%d account=%d", userID, accountID)
				}
				if req.Name == nil || *req.Name != "Travel" {
					t.Fatalf("unexpected update request: %+v", req)
				}
				account := sampleAccount()
				account.Name = "Travel"
				return &account, nil
			},
		}
		router := gin.New()
		router.PATCH("/accounts/:id", withUserID(42, (&AccountHandler{accountService: service}).Update))
		req := newJSONRequest(t, http.MethodPatch, "/accounts/7", `{"name":"Travel"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.updateCalls != 1 {
			t.Fatalf("expected Update to be called once, got %d", service.updateCalls)
		}
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		service := &accountServiceSpy{}
		router := gin.New()
		router.PATCH("/accounts/:id", withUserID(42, (&AccountHandler{accountService: service}).Update))
		req := newJSONRequest(t, http.MethodPatch, "/accounts/abc", `{"name":"Travel"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.updateCalls != 0 {
			t.Fatalf("expected Update not to be called, got %d calls", service.updateCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "invalid account id")
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		service := &accountServiceSpy{}
		router := gin.New()
		router.PATCH("/accounts/:id", withUserID(42, (&AccountHandler{accountService: service}).Update))
		req := newJSONRequest(t, http.MethodPatch, "/accounts/7", `{"currency":"usd"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.updateCalls != 0 {
			t.Fatalf("expected Update not to be called, got %d calls", service.updateCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})
}

func TestAccountHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 204", func(t *testing.T) {
		service := &accountServiceSpy{
			deleteFn: func(_ context.Context, userID, accountID int64) *apperror.Error {
				if userID != 42 || accountID != 7 {
					t.Fatalf("unexpected ids: user=%d account=%d", userID, accountID)
				}
				return nil
			},
		}
		router := gin.New()
		router.DELETE("/accounts/:id", withUserID(42, (&AccountHandler{accountService: service}).Delete))
		req := httptest.NewRequest(http.MethodDelete, "/accounts/7", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusNoContent)
		if service.deleteCalls != 1 {
			t.Fatalf("expected Delete to be called once, got %d", service.deleteCalls)
		}
		if strings.TrimSpace(rec.Body.String()) != "" {
			t.Fatalf("expected empty body, got %q", rec.Body.String())
		}
	})

	t.Run("service returns not found", func(t *testing.T) {
		service := &accountServiceSpy{
			deleteFn: func(_ context.Context, _, _ int64) *apperror.Error {
				return apperror.NotFound("account not found")
			},
		}
		router := gin.New()
		router.DELETE("/accounts/:id", withUserID(42, (&AccountHandler{accountService: service}).Delete))
		req := httptest.NewRequest(http.MethodDelete, "/accounts/7", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusNotFound)
		assertErrorEnvelope(t, rec, "NOT_FOUND", "account not found")
	})
}

func sampleAccount() models.Account {
	now := time.Unix(1700000000, 0).UTC()
	return models.Account{
		ID:          7,
		UserID:      42,
		Name:        "Cash",
		AccountType: "cash",
		Balance:     "100.50",
		Currency:    "USD",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func withUserID(userID int64, next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_user_id", userID)
		next(c)
	}
}
