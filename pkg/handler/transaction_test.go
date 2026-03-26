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

type transactionServiceSpy struct {
	listFn     func(ctx context.Context, userID int64, query models.ListTransactionsQuery) ([]models.Transaction, *apperror.Error)
	listCalls  int
	listUserID int64
	listQuery  models.ListTransactionsQuery

	createFn     func(ctx context.Context, userID int64, req models.CreateTransactionRequest) (*models.Transaction, *apperror.Error)
	createCalls  int
	createUserID int64
	createReq    models.CreateTransactionRequest

	getByIDFn     func(ctx context.Context, userID, txID int64) (*models.Transaction, *apperror.Error)
	getByIDCalls  int
	getByIDUserID int64
	getByIDID     int64

	updateFn     func(ctx context.Context, userID, txID int64, req models.UpdateTransactionRequest) (*models.Transaction, *apperror.Error)
	updateCalls  int
	updateUserID int64
	updateID     int64
	updateReq    models.UpdateTransactionRequest

	deleteFn     func(ctx context.Context, userID, txID int64) *apperror.Error
	deleteCalls  int
	deleteUserID int64
	deleteID     int64
}

func (s *transactionServiceSpy) List(ctx context.Context, userID int64, query models.ListTransactionsQuery) ([]models.Transaction, *apperror.Error) {
	s.listCalls++
	s.listUserID = userID
	s.listQuery = query
	if s.listFn == nil {
		panic("unexpected call: List")
	}
	return s.listFn(ctx, userID, query)
}

func (s *transactionServiceSpy) Create(ctx context.Context, userID int64, req models.CreateTransactionRequest) (*models.Transaction, *apperror.Error) {
	s.createCalls++
	s.createUserID = userID
	s.createReq = req
	if s.createFn == nil {
		panic("unexpected call: Create")
	}
	return s.createFn(ctx, userID, req)
}

func (s *transactionServiceSpy) GetByID(ctx context.Context, userID, txID int64) (*models.Transaction, *apperror.Error) {
	s.getByIDCalls++
	s.getByIDUserID = userID
	s.getByIDID = txID
	if s.getByIDFn == nil {
		panic("unexpected call: GetByID")
	}
	return s.getByIDFn(ctx, userID, txID)
}

func (s *transactionServiceSpy) Update(ctx context.Context, userID, txID int64, req models.UpdateTransactionRequest) (*models.Transaction, *apperror.Error) {
	s.updateCalls++
	s.updateUserID = userID
	s.updateID = txID
	s.updateReq = req
	if s.updateFn == nil {
		panic("unexpected call: Update")
	}
	return s.updateFn(ctx, userID, txID, req)
}

func (s *transactionServiceSpy) Delete(ctx context.Context, userID, txID int64) *apperror.Error {
	s.deleteCalls++
	s.deleteUserID = userID
	s.deleteID = txID
	if s.deleteFn == nil {
		panic("unexpected call: Delete")
	}
	return s.deleteFn(ctx, userID, txID)
}

func TestNewTransactionHandler(t *testing.T) {
	handler := NewTransactionHandler(nil)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if handler.txService != nil {
		t.Fatal("expected transaction service to be nil")
	}
}

func TestTransactionHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200 with bound query", func(t *testing.T) {
		service := &transactionServiceSpy{
			listFn: func(_ context.Context, userID int64, query models.ListTransactionsQuery) ([]models.Transaction, *apperror.Error) {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				if query.Page != 2 || query.Limit != 10 {
					t.Fatalf("unexpected query paging: %+v", query)
				}
				if query.Type == nil || *query.Type != "expense" {
					t.Fatalf("unexpected query type: %+v", query)
				}
				return []models.Transaction{sampleTransaction()}, nil
			},
		}
		router := gin.New()
		router.GET("/transactions", withUserID(42, (&TransactionHandler{txService: service}).List))
		req := httptest.NewRequest(http.MethodGet, "/transactions?type=expense&page=2&limit=10", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.listCalls != 1 || service.listUserID != 42 {
			t.Fatalf("unexpected List call state: calls=%d userID=%d", service.listCalls, service.listUserID)
		}
		if !strings.Contains(rec.Body.String(), `"description":"Groceries"`) {
			t.Fatalf("unexpected response body: %s", rec.Body.String())
		}
	})

	t.Run("invalid query returns 400", func(t *testing.T) {
		service := &transactionServiceSpy{}
		router := gin.New()
		router.GET("/transactions", withUserID(42, (&TransactionHandler{txService: service}).List))
		req := httptest.NewRequest(http.MethodGet, "/transactions?page=0", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.listCalls != 0 {
			t.Fatalf("expected List not to be called, got %d calls", service.listCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})
}

func TestTransactionHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 201", func(t *testing.T) {
		service := &transactionServiceSpy{
			createFn: func(_ context.Context, userID int64, req models.CreateTransactionRequest) (*models.Transaction, *apperror.Error) {
				if userID != 42 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				if req.AccountID != 7 || req.Amount != "18.90" || req.Currency != "USD" || req.Type != "expense" {
					t.Fatalf("unexpected create request: %+v", req)
				}
				tx := sampleTransaction()
				return &tx, nil
			},
		}
		router := gin.New()
		router.POST("/transactions", withUserID(42, (&TransactionHandler{txService: service}).Create))
		req := newJSONRequest(t, http.MethodPost, "/transactions", `{"account_id":7,"amount":"18.90","currency":"USD","type":"expense","description":"Groceries","transacted_at":"2024-01-02"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusCreated)
		if service.createCalls != 1 {
			t.Fatalf("expected Create to be called once, got %d", service.createCalls)
		}
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		service := &transactionServiceSpy{}
		router := gin.New()
		router.POST("/transactions", withUserID(42, (&TransactionHandler{txService: service}).Create))
		req := newJSONRequest(t, http.MethodPost, "/transactions", `{"account_id":0,"amount":"","currency":"usd","type":"bad","description":"","transacted_at":"02-01-2024"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.createCalls != 0 {
			t.Fatalf("expected Create not to be called, got %d calls", service.createCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})
}

func TestTransactionHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200", func(t *testing.T) {
		service := &transactionServiceSpy{
			getByIDFn: func(_ context.Context, userID, txID int64) (*models.Transaction, *apperror.Error) {
				if userID != 42 || txID != 9 {
					t.Fatalf("unexpected ids: user=%d tx=%d", userID, txID)
				}
				tx := sampleTransaction()
				return &tx, nil
			},
		}
		router := gin.New()
		router.GET("/transactions/:id", withUserID(42, (&TransactionHandler{txService: service}).GetByID))
		req := httptest.NewRequest(http.MethodGet, "/transactions/9", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.getByIDCalls != 1 {
			t.Fatalf("expected GetByID to be called once, got %d", service.getByIDCalls)
		}
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		service := &transactionServiceSpy{}
		router := gin.New()
		router.GET("/transactions/:id", withUserID(42, (&TransactionHandler{txService: service}).GetByID))
		req := httptest.NewRequest(http.MethodGet, "/transactions/abc", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.getByIDCalls != 0 {
			t.Fatalf("expected GetByID not to be called, got %d calls", service.getByIDCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "invalid transaction id")
	})
}

func TestTransactionHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200", func(t *testing.T) {
		service := &transactionServiceSpy{
			updateFn: func(_ context.Context, userID, txID int64, req models.UpdateTransactionRequest) (*models.Transaction, *apperror.Error) {
				if userID != 42 || txID != 9 {
					t.Fatalf("unexpected ids: user=%d tx=%d", userID, txID)
				}
				if req.Amount == nil || *req.Amount != "20.00" {
					t.Fatalf("unexpected update request: %+v", req)
				}
				tx := sampleTransaction()
				tx.Amount = "20.00"
				return &tx, nil
			},
		}
		router := gin.New()
		router.PATCH("/transactions/:id", withUserID(42, (&TransactionHandler{txService: service}).Update))
		req := newJSONRequest(t, http.MethodPatch, "/transactions/9", `{"amount":"20.00"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		if service.updateCalls != 1 {
			t.Fatalf("expected Update to be called once, got %d", service.updateCalls)
		}
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		service := &transactionServiceSpy{}
		router := gin.New()
		router.PATCH("/transactions/:id", withUserID(42, (&TransactionHandler{txService: service}).Update))
		req := newJSONRequest(t, http.MethodPatch, "/transactions/9", `{"notes":"`+strings.Repeat("a", 1001)+`"}`)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)
		if service.updateCalls != 0 {
			t.Fatalf("expected Update not to be called, got %d calls", service.updateCalls)
		}
		assertErrorEnvelope(t, rec, "VALIDATION_ERROR", "")
	})
}

func TestTransactionHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 204", func(t *testing.T) {
		service := &transactionServiceSpy{
			deleteFn: func(_ context.Context, userID, txID int64) *apperror.Error {
				if userID != 42 || txID != 9 {
					t.Fatalf("unexpected ids: user=%d tx=%d", userID, txID)
				}
				return nil
			},
		}
		router := gin.New()
		router.DELETE("/transactions/:id", withUserID(42, (&TransactionHandler{txService: service}).Delete))
		req := httptest.NewRequest(http.MethodDelete, "/transactions/9", nil)
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

	t.Run("service not found returns 404", func(t *testing.T) {
		service := &transactionServiceSpy{
			deleteFn: func(_ context.Context, _, _ int64) *apperror.Error {
				return apperror.NotFound("transaction not found")
			},
		}
		router := gin.New()
		router.DELETE("/transactions/:id", withUserID(42, (&TransactionHandler{txService: service}).Delete))
		req := httptest.NewRequest(http.MethodDelete, "/transactions/9", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusNotFound)
		assertErrorEnvelope(t, rec, "NOT_FOUND", "transaction not found")
	})
}

func sampleTransaction() models.Transaction {
	now := time.Unix(1700000000, 0).UTC()
	return models.Transaction{
		ID:           9,
		AccountID:    7,
		Amount:       "18.90",
		Currency:     "USD",
		Type:         "expense",
		Description:  "Groceries",
		TransactedAt: "2024-01-02",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
