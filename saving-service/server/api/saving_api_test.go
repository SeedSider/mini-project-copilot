package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/bankease/saving-service/protogen/saving-service"
	"github.com/bankease/saving-service/server/db"
	"github.com/bankease/saving-service/server/lib/logger"
)

func newTestLogger() *logger.Logger {
	return logger.New(&logger.LoggerConfig{
		Env:         "DEV",
		ServiceName: "test",
		ProductName: "test",
		LogLevel:    "error",
		LogOutput:   "stdout",
	})
}

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	testLogger := newTestLogger()
	provider := db.New(sqlDB, testLogger)
	return New(provider, testLogger), mock
}

var testExchangeRateCols = []string{"id", "country", "currency", "country_code", "buy", "sell"}
var testInterestRateCols = []string{"id", "kind", "deposit", "rate"}
var testBranchCols = []string{"id", "name", "distance", "latitude", "longitude"}

// ═══════════════════════════════════════════
// HTTP HandleGetExchangeRates
// ═══════════════════════════════════════════

func TestHandleGetExchangeRates_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnRows(sqlmock.NewRows(testExchangeRateCols).
			AddRow("er-1", "United States", "USD", "US", 16000.0, 16200.0).
			AddRow("er-2", "Japan", "JPY", "JP", 107.5, 108.0))

	r := httptest.NewRequest(http.MethodGet, "/api/exchange-rates", nil)
	w := httptest.NewRecorder()

	srv.HandleGetExchangeRates(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetExchangeRates_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnError(fmt.Errorf("db error"))

	r := httptest.NewRequest(http.MethodGet, "/api/exchange-rates", nil)
	w := httptest.NewRecorder()

	srv.HandleGetExchangeRates(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetInterestRates
// ═══════════════════════════════════════════

func TestHandleGetInterestRates_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnRows(sqlmock.NewRows(testInterestRateCols).
			AddRow("ir-1", "Tabungan", "1 Bulan", 3.5).
			AddRow("ir-2", "Deposito", "3 Bulan", 4.25))

	r := httptest.NewRequest(http.MethodGet, "/api/interest-rates", nil)
	w := httptest.NewRecorder()

	srv.HandleGetInterestRates(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetInterestRates_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnError(fmt.Errorf("db error"))

	r := httptest.NewRequest(http.MethodGet, "/api/interest-rates", nil)
	w := httptest.NewRecorder()

	srv.HandleGetInterestRates(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetBranches
// ═══════════════════════════════════════════

func TestHandleGetBranches_NoQuery_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch`).
		WillReturnRows(sqlmock.NewRows(testBranchCols).
			AddRow("br-1", "BRI Jakarta Pusat", "1.2 km", -6.2088, 106.8456).
			AddRow("br-2", "BRI Bandung", "5.0 km", -6.9175, 107.6191))

	r := httptest.NewRequest(http.MethodGet, "/api/branches", nil)
	w := httptest.NewRecorder()

	srv.HandleGetBranches(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetBranches_WithQuery_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WithArgs("jakarta").
		WillReturnRows(sqlmock.NewRows(testBranchCols).
			AddRow("br-1", "BRI Jakarta Pusat", "1.2 km", -6.2088, 106.8456))

	r := httptest.NewRequest(http.MethodGet, "/api/branches?q=jakarta", nil)
	w := httptest.NewRecorder()

	srv.HandleGetBranches(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetBranches_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch`).
		WillReturnError(fmt.Errorf("db error"))

	r := httptest.NewRequest(http.MethodGet, "/api/branches", nil)
	w := httptest.NewRecorder()

	srv.HandleGetBranches(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// gRPC: GetExchangeRates
// ═══════════════════════════════════════════

func TestGetExchangeRatesGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnRows(sqlmock.NewRows(testExchangeRateCols).
			AddRow("er-1", "United States", "USD", "US", 16000.0, 16200.0))

	resp, err := srv.GetExchangeRates(context.Background(), &pb.GetExchangeRatesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.ExchangeRates, 1)
	assert.Equal(t, "USD", resp.ExchangeRates[0].Currency)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExchangeRatesGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetExchangeRates(context.Background(), &pb.GetExchangeRatesRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetInterestRates
// ═══════════════════════════════════════════

func TestGetInterestRatesGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnRows(sqlmock.NewRows(testInterestRateCols).
			AddRow("ir-1", "Tabungan", "1 Bulan", 3.5))

	resp, err := srv.GetInterestRates(context.Background(), &pb.GetInterestRatesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.InterestRates, 1)
	assert.Equal(t, "Tabungan", resp.InterestRates[0].Kind)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetInterestRatesGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetInterestRates(context.Background(), &pb.GetInterestRatesRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetBranches
// ═══════════════════════════════════════════

func TestGetBranchesGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch`).
		WillReturnRows(sqlmock.NewRows(testBranchCols).
			AddRow("br-1", "BRI Jakarta Pusat", "1.2 km", -6.2088, 106.8456).
			AddRow("br-2", "BRI Bandung", "5.0 km", -6.9175, 107.6191))

	resp, err := srv.GetBranches(context.Background(), &pb.GetBranchesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Branches, 2)
	assert.Equal(t, "BRI Jakarta Pusat", resp.Branches[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBranchesGRPC_WithQuery(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WithArgs("bandung").
		WillReturnRows(sqlmock.NewRows(testBranchCols).
			AddRow("br-2", "BRI Bandung", "5.0 km", -6.9175, 107.6191))

	resp, err := srv.GetBranches(context.Background(), &pb.GetBranchesRequest{Query: "bandung"})
	require.NoError(t, err)
	assert.Len(t, resp.Branches, 1)
	assert.Equal(t, "BRI Bandung", resp.Branches[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBranchesGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetBranches(context.Background(), &pb.GetBranchesRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGetBranchesGRPC_WithQuery_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WithArgs("jakarta").
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetBranches(context.Background(), &pb.GetBranchesRequest{Query: "jakarta"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// Server helper methods
// ═══════════════════════════════════════════

func TestServerError(t *testing.T) {
	srv, _ := newTestServer(t)
	err := srv.serverError()
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestBadRequestError(t *testing.T) {
	srv, _ := newTestServer(t)
	err := srv.badRequestError("field is required")
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "field is required")
}
