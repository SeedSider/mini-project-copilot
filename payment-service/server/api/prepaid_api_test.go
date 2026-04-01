package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	manager "github.com/bankease/payment-service/server/jwt"
	"github.com/bankease/payment-service/server/db"
	"github.com/bankease/payment-service/server/lib/database"
	databasemock "github.com/bankease/payment-service/server/lib/database/mock"
	"github.com/bankease/payment-service/server/lib/logger"
)

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dbSql := &database.DbSql{
		SqlDb: sqlDB,
		Dbw:   &databasemock.DatabaseMock{DbPq: sqlDB},
	}
	dbSql.Conn = sqlDB

	testLogger := logger.New(&logger.LoggerConfig{
		Env:         "DEV",
		ServiceName: "test",
		ProductName: "test",
		LogLevel:    "error",
		LogOutput:   "stdout",
	})

	prov := db.New(dbSql, testLogger)
	srv := New("test-secret", prov, testLogger)
	return srv, mock
}

func ctxWithClaims() context.Context {
	return context.WithValue(context.Background(), "user_claims", &manager.UserClaims{
		UserID:   "user-001",
		Username: "testuser",
	})
}

// ═══════════════════════════════════════════════════════════
// HTTP HandleGetBeneficiaries
// ═══════════════════════════════════════════════════════════

func TestHandleGetBeneficiaries_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"id", "account_id", "name", "phone", "avatar"}).
		AddRow("b-1", "acc-1", "Alice", "081234567890", "").
		AddRow("b-2", "acc-1", "Bob", "081234567891", "https://avatar.url")
	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WithArgs("acc-1").
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/mobile-prepaid/beneficiaries?accountId=acc-1", nil)
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandleGetBeneficiaries(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result []db.Beneficiary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Len(t, result, 2)
	assert.Equal(t, "Alice", result[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetBeneficiaries_EmptyList(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"id", "account_id", "name", "phone", "avatar"})
	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WithArgs("acc-1").
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/mobile-prepaid/beneficiaries?accountId=acc-1", nil)
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandleGetBeneficiaries(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result []db.Beneficiary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetBeneficiaries_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/mobile-prepaid/beneficiaries?accountId=acc-1", nil)
	w := httptest.NewRecorder()

	srv.HandleGetBeneficiaries(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var result prepaidErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "UNAUTHORIZED", result.Error)
}

func TestHandleGetBeneficiaries_MissingAccountId(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/mobile-prepaid/beneficiaries", nil)
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandleGetBeneficiaries(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var result prepaidErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "VALIDATION_ERROR", result.Error)
}

func TestHandleGetBeneficiaries_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WithArgs("acc-1").
		WillReturnError(errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/mobile-prepaid/beneficiaries?accountId=acc-1", nil)
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandleGetBeneficiaries(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// HTTP HandlePrepaidPay
// ═══════════════════════════════════════════════════════════

func TestHandlePrepaidPay_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	// No existing txn
	mock.ExpectQuery("SELECT id, status, message, created_at FROM prepaid_transaction").
		WithArgs("key-001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at"}))

	// Insert
	mock.ExpectExec("INSERT INTO prepaid_transaction").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"cardId":"card-001","phone":"081234567890","amount":1000}`
	req := httptest.NewRequest(http.MethodPost, "/api/mobile-prepaid/pay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "key-001")
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandlePrepaidPay(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result db.PrepaidTransaction
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "SUCCESS", result.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandlePrepaidPay_IdempotencyHit(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, status, message, created_at FROM prepaid_transaction").
		WithArgs("key-dup").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at"}).
			AddRow("txn-existing", "SUCCESS", "Top-up successful", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)))

	body := `{"cardId":"card-001","phone":"081234567890","amount":1000}`
	req := httptest.NewRequest(http.MethodPost, "/api/mobile-prepaid/pay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "key-dup")
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandlePrepaidPay(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result db.PrepaidTransaction
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "txn-existing", result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandlePrepaidPay_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)

	body := `{"cardId":"card-001","phone":"081234567890","amount":1000}`
	req := httptest.NewRequest(http.MethodPost, "/api/mobile-prepaid/pay", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "key-001")
	w := httptest.NewRecorder()

	srv.HandlePrepaidPay(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandlePrepaidPay_MissingIdempotencyKey(t *testing.T) {
	srv, _ := newTestServer(t)

	body := `{"cardId":"card-001","phone":"081234567890","amount":1000}`
	req := httptest.NewRequest(http.MethodPost, "/api/mobile-prepaid/pay", strings.NewReader(body))
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandlePrepaidPay(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlePrepaidPay_InvalidBody(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/mobile-prepaid/pay", strings.NewReader("{invalid"))
	req.Header.Set("Idempotency-Key", "key-001")
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandlePrepaidPay(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlePrepaidPay_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		msg  string
	}{
		{"missing cardId", `{"cardId":"","phone":"081234567890","amount":1000}`, "cardId is required"},
		{"missing phone", `{"cardId":"card-1","phone":"","amount":1000}`, "phone is required"},
		{"short phone", `{"cardId":"card-1","phone":"0812345","amount":1000}`, "phone must be 10-13 digits"},
		{"alpha phone", `{"cardId":"card-1","phone":"08abc123456","amount":1000}`, "phone must be 10-13 digits"},
		{"zero amount", `{"cardId":"card-1","phone":"081234567890","amount":0}`, "amount must be greater than 0"},
		{"negative amount", `{"cardId":"card-1","phone":"081234567890","amount":-5}`, "amount must be greater than 0"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			req := httptest.NewRequest(http.MethodPost, "/api/mobile-prepaid/pay", strings.NewReader(tc.body))
			req.Header.Set("Idempotency-Key", "key-001")
			req = req.WithContext(ctxWithClaims())
			w := httptest.NewRecorder()

			srv.HandlePrepaidPay(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			var result prepaidErrorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
			assert.Equal(t, tc.msg, result.Message)
		})
	}
}

func TestHandlePrepaidPay_IdempotencyCheckDBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, status, message, created_at FROM prepaid_transaction").
		WithArgs("key-err").
		WillReturnError(errors.New("db error"))

	body := `{"cardId":"card-001","phone":"081234567890","amount":1000}`
	req := httptest.NewRequest(http.MethodPost, "/api/mobile-prepaid/pay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "key-err")
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandlePrepaidPay(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandlePrepaidPay_InsertDBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, status, message, created_at FROM prepaid_transaction").
		WithArgs("key-ins-err").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at"}))

	mock.ExpectExec("INSERT INTO prepaid_transaction").
		WillReturnError(errors.New("insert error"))

	body := `{"cardId":"card-001","phone":"081234567890","amount":1000}`
	req := httptest.NewRequest(http.MethodPost, "/api/mobile-prepaid/pay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "key-ins-err")
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandlePrepaidPay(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// validatePrepaidPay
// ═══════════════════════════════════════════════════════════

func TestValidatePrepaidPay(t *testing.T) {
	tests := []struct {
		name string
		req  prepaidPayHTTPRequest
		want string
	}{
		{"valid", prepaidPayHTTPRequest{CardID: "c1", Phone: "081234567890", Amount: 100}, ""},
		{"empty cardId", prepaidPayHTTPRequest{Phone: "081234567890", Amount: 100}, "cardId is required"},
		{"empty phone", prepaidPayHTTPRequest{CardID: "c1", Amount: 100}, "phone is required"},
		{"phone too short", prepaidPayHTTPRequest{CardID: "c1", Phone: "08123", Amount: 100}, "phone must be 10-13 digits"},
		{"phone too long", prepaidPayHTTPRequest{CardID: "c1", Phone: "08123456789012345", Amount: 100}, "phone must be 10-13 digits"},
		{"phone with letters", prepaidPayHTTPRequest{CardID: "c1", Phone: "08abc12345", Amount: 100}, "phone must be 10-13 digits"},
		{"zero amount", prepaidPayHTTPRequest{CardID: "c1", Phone: "081234567890", Amount: 0}, "amount must be greater than 0"},
		{"negative amount", prepaidPayHTTPRequest{CardID: "c1", Phone: "081234567890", Amount: -1}, "amount must be greater than 0"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := validatePrepaidPay(tc.req)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ═══════════════════════════════════════════════════════════
// writePrepaidError
// ═══════════════════════════════════════════════════════════

func TestWritePrepaidError(t *testing.T) {
	w := httptest.NewRecorder()
	writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", "test message")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var result prepaidErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "VALIDATION_ERROR", result.Error)
	assert.Equal(t, "test message", result.Message)
}

// ═══════════════════════════════════════════════════════════
// phoneRegex
// ═══════════════════════════════════════════════════════════

func TestPhoneRegex(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{"0812345678", true},    // 10 digits
		{"08123456789", true},   // 11 digits
		{"081234567890", true},  // 12 digits
		{"0812345678901", true}, // 13 digits
		{"081234567", false},    // 9 digits
		{"08123456789012", false}, // 14 digits
		{"08abc123456", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("phone_%s", tc.phone), func(t *testing.T) {
			assert.Equal(t, tc.valid, phoneRegex.MatchString(tc.phone))
		})
	}
}
