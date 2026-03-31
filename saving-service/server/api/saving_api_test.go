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

const (
	testAPIExchangeRates = "/api/exchange-rates"
	testAPIInterestRates = "/api/interest-rates"
	testAPIBranches      = "/api/branches"
	testDBError          = "db error"
	testBRIJakarta       = "BRI Jakarta Pusat"
	testBRIBandung        = "BRI Bandung"
	testDistance1          = "1.2 km"
	testDistance2          = "5.0 km"
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

func TestHandleGetExchangeRatesSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnRows(sqlmock.NewRows(testExchangeRateCols).
			AddRow("er-1", "United States", "USD", "US", 16000.0, 16200.0).
			AddRow("er-2", "Japan", "JPY", "JP", 107.5, 108.0))

	r := httptest.NewRequest(http.MethodGet, testAPIExchangeRates, nil)
	w := httptest.NewRecorder()

	srv.HandleGetExchangeRates(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetExchangeRatesDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnError(fmt.Errorf(testDBError))

	r := httptest.NewRequest(http.MethodGet, testAPIExchangeRates, nil)
	w := httptest.NewRecorder()

	srv.HandleGetExchangeRates(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetInterestRates
// ═══════════════════════════════════════════

func TestHandleGetInterestRatesSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnRows(sqlmock.NewRows(testInterestRateCols).
			AddRow("ir-1", "Tabungan", "1 Bulan", 3.5).
			AddRow("ir-2", "Deposito", "3 Bulan", 4.25))

	r := httptest.NewRequest(http.MethodGet, testAPIInterestRates, nil)
	w := httptest.NewRecorder()

	srv.HandleGetInterestRates(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetInterestRatesDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnError(fmt.Errorf(testDBError))

	r := httptest.NewRequest(http.MethodGet, testAPIInterestRates, nil)
	w := httptest.NewRecorder()

	srv.HandleGetInterestRates(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetBranches
// ═══════════════════════════════════════════

func TestHandleGetBranchesNoQuerySuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch`).
		WillReturnRows(sqlmock.NewRows(testBranchCols).
			AddRow("br-1", testBRIJakarta, testDistance1, -6.2088, 106.8456).
			AddRow("br-2", testBRIBandung, testDistance2, -6.9175, 107.6191))

	r := httptest.NewRequest(http.MethodGet, testAPIBranches, nil)
	w := httptest.NewRecorder()

	srv.HandleGetBranches(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetBranchesWithQuerySuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WithArgs("jakarta").
		WillReturnRows(sqlmock.NewRows(testBranchCols).
			AddRow("br-1", testBRIJakarta, testDistance1, -6.2088, 106.8456))

	r := httptest.NewRequest(http.MethodGet, testAPIBranches+"?q=jakarta", nil)
	w := httptest.NewRecorder()

	srv.HandleGetBranches(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetBranchesDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch`).
		WillReturnError(fmt.Errorf(testDBError))

	r := httptest.NewRequest(http.MethodGet, testAPIBranches, nil)
	w := httptest.NewRecorder()

	srv.HandleGetBranches(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// gRPC: GetExchangeRates
// ═══════════════════════════════════════════

func TestGetExchangeRatesGRPCSuccess(t *testing.T) {
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

func TestGetExchangeRatesGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnError(fmt.Errorf(testDBError))

	_, err := srv.GetExchangeRates(context.Background(), &pb.GetExchangeRatesRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetInterestRates
// ═══════════════════════════════════════════

func TestGetInterestRatesGRPCSuccess(t *testing.T) {
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

func TestGetInterestRatesGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnError(fmt.Errorf(testDBError))

	_, err := srv.GetInterestRates(context.Background(), &pb.GetInterestRatesRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetBranches
// ═══════════════════════════════════════════

func TestGetBranchesGRPCSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch`).
		WillReturnRows(sqlmock.NewRows(testBranchCols).
			AddRow("br-1", testBRIJakarta, testDistance1, -6.2088, 106.8456).
			AddRow("br-2", testBRIBandung, testDistance2, -6.9175, 107.6191))

	resp, err := srv.GetBranches(context.Background(), &pb.GetBranchesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Branches, 2)
	assert.Equal(t, testBRIJakarta, resp.Branches[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBranchesGRPCWithQuery(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WithArgs("bandung").
		WillReturnRows(sqlmock.NewRows(testBranchCols).
			AddRow("br-2", testBRIBandung, testDistance2, -6.9175, 107.6191))

	resp, err := srv.GetBranches(context.Background(), &pb.GetBranchesRequest{Query: "bandung"})
	require.NoError(t, err)
	assert.Len(t, resp.Branches, 1)
	assert.Equal(t, testBRIBandung, resp.Branches[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBranchesGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch`).
		WillReturnError(fmt.Errorf(testDBError))

	_, err := srv.GetBranches(context.Background(), &pb.GetBranchesRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGetBranchesGRPCWithQueryDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WithArgs("jakarta").
		WillReturnError(fmt.Errorf(testDBError))

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
