package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bankease/payment-service/server/db"
)

// ═══════════════════════════════════════════════════════════
// HTTP HandleGetProviders
// ═══════════════════════════════════════════════════════════

func TestHandleGetProviders_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("1", "Biznet").
		AddRow("2", "Indihome")
	mock.ExpectQuery("SELECT id, name FROM provider").WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/pay-the-bill/providers", nil)
	w := httptest.NewRecorder()

	srv.HandleGetProviders(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result []db.ServiceProvider
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Len(t, result, 2)
	assert.Equal(t, "Biznet", result[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetProviders_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, name FROM provider").WillReturnError(errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/pay-the-bill/providers", nil)
	w := httptest.NewRecorder()

	srv.HandleGetProviders(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// HTTP HandleGetInternetBill
// ═══════════════════════════════════════════════════════════

func TestHandleGetInternetBill_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "customer_id", "name", "address", "phone_number", "code", "bill_from", "bill_to", "internet_fee", "tax", "total"}).
		AddRow("bill-1", "user-001", "CUST-001", "John", "Jl. Testing 1", "08123", "BZN-001", "2026-01-01", "2026-01-31", "250000", "25000", "275000")
	mock.ExpectQuery("SELECT id, user_id, customer_id, name, address, phone_number, code, bill_from, bill_to, internet_fee, tax, total FROM internet_bill").
		WithArgs("user-001").
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/pay-the-bill/internet-bill", nil)
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandleGetInternetBill(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result db.InternetBill
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "CUST-001", result.CustomerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetInternetBill_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/pay-the-bill/internet-bill", nil)
	w := httptest.NewRecorder()

	srv.HandleGetInternetBill(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGetInternetBill_NotFound(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, user_id, customer_id, name, address, phone_number, code, bill_from, bill_to, internet_fee, tax, total FROM internet_bill").
		WithArgs("user-001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "customer_id", "name", "address", "phone_number", "code", "bill_from", "bill_to", "internet_fee", "tax", "total"}))

	req := httptest.NewRequest(http.MethodGet, "/api/pay-the-bill/internet-bill", nil)
	req = req.WithContext(ctxWithClaims())
	w := httptest.NewRecorder()

	srv.HandleGetInternetBill(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// HTTP HandleGetCurrencyList
// ═══════════════════════════════════════════════════════════

func TestHandleGetCurrencyList_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"code", "label", "rate"}).
		AddRow("USD", "US Dollar", 15800.50).
		AddRow("EUR", "Euro", 17200.75)
	mock.ExpectQuery("SELECT code, label, rate FROM currency").WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/currency-list", nil)
	w := httptest.NewRecorder()

	srv.HandleGetCurrencyList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result []db.Currency
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Len(t, result, 2)
	assert.Equal(t, "USD", result[0].Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetCurrencyList_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT code, label, rate FROM currency").WillReturnError(errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/currency-list", nil)
	w := httptest.NewRecorder()

	srv.HandleGetCurrencyList(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// Helper writeJSON / writeError / writeAuthError
// ═══════════════════════════════════════════════════════════

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusInternalServerError, "Something failed")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var result standardResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, 500, result.Code)
	assert.Equal(t, "Something failed", result.Description)
}

func TestWriteAuthError(t *testing.T) {
	w := httptest.NewRecorder()
	writeAuthError(w, http.StatusUnauthorized, "Not authorized")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var result errorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.True(t, result.Error)
	assert.Equal(t, 401, result.Code)
	assert.Equal(t, "Not authorized", result.Message)
}

// ═══════════════════════════════════════════════════════════
// error helpers
// ═══════════════════════════════════════════════════════════

func TestServerError(t *testing.T) {
	srv, _ := newTestServer(t)
	err := srv.serverError()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Internal Error")
}

func TestUnauthorizedError(t *testing.T) {
	srv, _ := newTestServer(t)
	err := srv.unauthorizedError()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unauthorized")
}
