package db

import (
	"context"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bankease/saving-service/server/lib/logger"
)

const (
	testDBError    = "db error"
	testBRIJakarta = "BRI Jakarta Pusat"
)

func newTestProvider(t *testing.T) (*Provider, sqlmock.Sqlmock) {
	t.Helper()
	dbConn, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { dbConn.Close() })

	testLogger := logger.New(&logger.LoggerConfig{
		Env:         "DEV",
		ServiceName: "test",
		ProductName: "test",
		LogLevel:    "error",
		LogOutput:   "stdout",
	})

	return New(dbConn, testLogger), mock
}

var exchangeRateCols = []string{"id", "country", "currency", "country_code", "buy", "sell"}
var interestRateCols = []string{"id", "kind", "deposit", "rate"}
var branchCols = []string{"id", "name", "distance", "latitude", "longitude"}

// ── GetAllExchangeRates ──

func TestGetAllExchangeRatesSuccess(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnRows(sqlmock.NewRows(exchangeRateCols).
			AddRow("er-1", "United States", "USD", "US", 16000.0, 16200.0).
			AddRow("er-2", "Japan", "JPY", "JP", 107.5, 108.0))

	rates, err := p.GetAllExchangeRates(context.Background())
	require.NoError(t, err)
	assert.Len(t, rates, 2)
	assert.Equal(t, "USD", rates[0].Currency)
	assert.Equal(t, "JPY", rates[1].Currency)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllExchangeRatesEmpty(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnRows(sqlmock.NewRows(exchangeRateCols))

	rates, err := p.GetAllExchangeRates(context.Background())
	require.NoError(t, err)
	assert.Empty(t, rates)
}

func TestGetAllExchangeRatesDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM exchange_rate`).
		WillReturnError(fmt.Errorf(testDBError))

	rates, err := p.GetAllExchangeRates(context.Background())
	assert.Error(t, err)
	assert.Nil(t, rates)
}

// ── GetAllInterestRates ──

func TestGetAllInterestRatesSuccess(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnRows(sqlmock.NewRows(interestRateCols).
			AddRow("ir-1", "Tabungan", "1 Bulan", 3.5).
			AddRow("ir-2", "Deposito", "3 Bulan", 4.25))

	rates, err := p.GetAllInterestRates(context.Background())
	require.NoError(t, err)
	assert.Len(t, rates, 2)
	assert.Equal(t, "Tabungan", rates[0].Kind)
	assert.Equal(t, 4.25, rates[1].Rate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllInterestRatesEmpty(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnRows(sqlmock.NewRows(interestRateCols))

	rates, err := p.GetAllInterestRates(context.Background())
	require.NoError(t, err)
	assert.Empty(t, rates)
}

func TestGetAllInterestRatesDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM interest_rate`).
		WillReturnError(fmt.Errorf(testDBError))

	rates, err := p.GetAllInterestRates(context.Background())
	assert.Error(t, err)
	assert.Nil(t, rates)
}

// ── GetAllBranches ──

func TestGetAllBranchesSuccess(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM branch`).
		WillReturnRows(sqlmock.NewRows(branchCols).
			AddRow("br-1", testBRIJakarta, "1.2 km", -6.2088, 106.8456).
			AddRow("br-2", "BRI Bandung", "5.0 km", -6.9175, 107.6191))

	branches, err := p.GetAllBranches(context.Background())
	require.NoError(t, err)
	assert.Len(t, branches, 2)
	assert.Equal(t, testBRIJakarta, branches[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllBranchesEmpty(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM branch`).
		WillReturnRows(sqlmock.NewRows(branchCols))

	branches, err := p.GetAllBranches(context.Background())
	require.NoError(t, err)
	assert.Empty(t, branches)
}

func TestGetAllBranchesDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM branch`).
		WillReturnError(fmt.Errorf(testDBError))

	branches, err := p.GetAllBranches(context.Background())
	assert.Error(t, err)
	assert.Nil(t, branches)
}

// ── SearchBranchesByName ──

func TestSearchBranchesByNameFound(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WithArgs("jakarta").
		WillReturnRows(sqlmock.NewRows(branchCols).
			AddRow("br-1", testBRIJakarta, "1.2 km", -6.2088, 106.8456))

	branches, err := p.SearchBranchesByName(context.Background(), "jakarta")
	require.NoError(t, err)
	assert.Len(t, branches, 1)
	assert.Equal(t, testBRIJakarta, branches[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBranchesByNameNotFound(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows(branchCols))

	branches, err := p.SearchBranchesByName(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, branches)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchBranchesByNameDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM branch WHERE name ILIKE`).
		WillReturnError(fmt.Errorf(testDBError))

	branches, err := p.SearchBranchesByName(context.Background(), "any")
	assert.Error(t, err)
	assert.Nil(t, branches)
}
