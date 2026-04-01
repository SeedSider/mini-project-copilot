package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/bankease/payment-service/protogen/payment-service"
	"github.com/bankease/payment-service/server/db"
)

// ═══════════════════════════════════════════════════════════
// gRPC GetProviders
// ═══════════════════════════════════════════════════════════

func TestGRPC_GetProviders_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("1", "Biznet").
		AddRow("2", "Indihome")
	mock.ExpectQuery("SELECT id, name FROM provider").WillReturnRows(rows)

	resp, err := srv.GetProviders(context.Background(), &pb.GetProvidersRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Providers, 2)
	assert.Equal(t, "Biznet", resp.Providers[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGRPC_GetProviders_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, name FROM provider").WillReturnError(errors.New("db err"))

	_, err := srv.GetProviders(context.Background(), &pb.GetProvidersRequest{})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// gRPC GetInternetBill
// ═══════════════════════════════════════════════════════════

func TestGRPC_GetInternetBill_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "customer_id", "name", "address", "phone_number", "code", "bill_from", "bill_to", "internet_fee", "tax", "total"}).
		AddRow("bill-1", "user-001", "CUST-001", "John", "Jl. Test", "08123", "BZN-001", "2026-01", "2026-02", "250000", "25000", "275000")
	mock.ExpectQuery("SELECT id, user_id, customer_id, name, address").
		WithArgs("user-001").
		WillReturnRows(rows)

	resp, err := srv.GetInternetBill(ctxWithClaims(), &pb.GetInternetBillRequest{})
	require.NoError(t, err)
	assert.Equal(t, "CUST-001", resp.Bill.CustomerId)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGRPC_GetInternetBill_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)

	_, err := srv.GetInternetBill(context.Background(), &pb.GetInternetBillRequest{})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPC_GetInternetBill_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, user_id, customer_id, name, address").
		WithArgs("user-001").
		WillReturnError(errors.New("db error"))

	_, err := srv.GetInternetBill(ctxWithClaims(), &pb.GetInternetBillRequest{})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// gRPC GetCurrencyList
// ═══════════════════════════════════════════════════════════

func TestGRPC_GetCurrencyList_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"code", "label", "rate"}).
		AddRow("USD", "US Dollar", 15800.50)
	mock.ExpectQuery("SELECT code, label, rate FROM currency").WillReturnRows(rows)

	resp, err := srv.GetCurrencyList(context.Background(), &pb.GetCurrencyListRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Currencies, 1)
	assert.Equal(t, "USD", resp.Currencies[0].Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGRPC_GetCurrencyList_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT code, label, rate FROM currency").WillReturnError(errors.New("db err"))

	_, err := srv.GetCurrencyList(context.Background(), &pb.GetCurrencyListRequest{})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// gRPC GetBeneficiaries
// ═══════════════════════════════════════════════════════════

func TestGRPC_GetBeneficiaries_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"id", "account_id", "name", "phone", "avatar"}).
		AddRow("b-1", "acc-1", "Alice", "081234567890", "")
	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WithArgs("acc-1").
		WillReturnRows(rows)

	resp, err := srv.GetBeneficiaries(ctxWithClaims(), &pb.GetBeneficiariesRequest{AccountId: "acc-1"})
	require.NoError(t, err)
	assert.Len(t, resp.Beneficiaries, 1)
	assert.Equal(t, "Alice", resp.Beneficiaries[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGRPC_GetBeneficiaries_EmptyAccountId(t *testing.T) {
	srv, _ := newTestServer(t)

	_, err := srv.GetBeneficiaries(ctxWithClaims(), &pb.GetBeneficiariesRequest{AccountId: ""})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPC_GetBeneficiaries_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)

	_, err := srv.GetBeneficiaries(context.Background(), &pb.GetBeneficiariesRequest{AccountId: "acc-1"})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPC_GetBeneficiaries_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WithArgs("acc-1").
		WillReturnError(errors.New("db error"))

	_, err := srv.GetBeneficiaries(ctxWithClaims(), &pb.GetBeneficiariesRequest{AccountId: "acc-1"})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGRPC_GetBeneficiaries_EmptyResult(t *testing.T) {
	srv, mock := newTestServer(t)

	rows := sqlmock.NewRows([]string{"id", "account_id", "name", "phone", "avatar"})
	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WithArgs("acc-1").
		WillReturnRows(rows)

	resp, err := srv.GetBeneficiaries(ctxWithClaims(), &pb.GetBeneficiariesRequest{AccountId: "acc-1"})
	require.NoError(t, err)
	assert.Empty(t, resp.Beneficiaries)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// gRPC PrepaidPay
// ═══════════════════════════════════════════════════════════

func TestGRPC_PrepaidPay_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	// No existing txn
	mock.ExpectQuery("SELECT id, status, message, created_at FROM prepaid_transaction").
		WithArgs("idem-001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at"}))

	mock.ExpectExec("INSERT INTO prepaid_transaction").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.PrepaidPay(ctxWithClaims(), &pb.PrepaidPayRequest{
		IdempotencyKey: "idem-001",
		CardId:         "card-001",
		Phone:          "081234567890",
		Amount:         2000,
	})
	require.NoError(t, err)
	assert.Equal(t, "SUCCESS", resp.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGRPC_PrepaidPay_IdempotencyHit(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, status, message, created_at FROM prepaid_transaction").
		WithArgs("idem-dup").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at"}).
			AddRow("txn-existing", "SUCCESS", "Top-up successful", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)))

	resp, err := srv.PrepaidPay(ctxWithClaims(), &pb.PrepaidPayRequest{
		IdempotencyKey: "idem-dup",
		CardId:         "card-001",
		Phone:          "081234567890",
		Amount:         2000,
	})
	require.NoError(t, err)
	assert.Equal(t, "txn-existing", resp.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGRPC_PrepaidPay_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)

	_, err := srv.PrepaidPay(context.Background(), &pb.PrepaidPayRequest{
		IdempotencyKey: "idem-001",
		CardId:         "card-001",
		Phone:          "081234567890",
		Amount:         2000,
	})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPC_PrepaidPay_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *pb.PrepaidPayRequest
	}{
		{"missing idempotencyKey", &pb.PrepaidPayRequest{CardId: "c1", Phone: "081234567890", Amount: 1000}},
		{"missing cardId", &pb.PrepaidPayRequest{IdempotencyKey: "k1", Phone: "081234567890", Amount: 1000}},
		{"missing phone", &pb.PrepaidPayRequest{IdempotencyKey: "k1", CardId: "c1", Amount: 1000}},
		{"short phone", &pb.PrepaidPayRequest{IdempotencyKey: "k1", CardId: "c1", Phone: "0812345", Amount: 1000}},
		{"zero amount", &pb.PrepaidPayRequest{IdempotencyKey: "k1", CardId: "c1", Phone: "081234567890", Amount: 0}},
		{"negative amount", &pb.PrepaidPayRequest{IdempotencyKey: "k1", CardId: "c1", Phone: "081234567890", Amount: -5}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			_, err := srv.PrepaidPay(ctxWithClaims(), tc.req)
			assert.Error(t, err)
			st, _ := status.FromError(err)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		})
	}
}

func TestGRPC_PrepaidPay_IdempotencyCheckDBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, status, message, created_at FROM prepaid_transaction").
		WithArgs("idem-err").
		WillReturnError(errors.New("db error"))

	_, err := srv.PrepaidPay(ctxWithClaims(), &pb.PrepaidPayRequest{
		IdempotencyKey: "idem-err",
		CardId:         "card-001",
		Phone:          "081234567890",
		Amount:         2000,
	})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGRPC_PrepaidPay_InsertDBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery("SELECT id, status, message, created_at FROM prepaid_transaction").
		WithArgs("idem-ins-err").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at"}))

	mock.ExpectExec("INSERT INTO prepaid_transaction").
		WillReturnError(errors.New("insert error"))

	_, err := srv.PrepaidPay(ctxWithClaims(), &pb.PrepaidPayRequest{
		IdempotencyKey: "idem-ins-err",
		CardId:         "card-001",
		Phone:          "081234567890",
		Amount:         2000,
	})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ═══════════════════════════════════════════════════════════
// converter functions
// ═══════════════════════════════════════════════════════════

func TestProvidersToProto(t *testing.T) {
	providers := []db.ServiceProvider{
		{ID: "1", Name: "Biznet"},
		{ID: "2", Name: "Indihome"},
	}
	result := providersToProto(providers)
	assert.Len(t, result, 2)
	assert.Equal(t, "Biznet", result[0].Name)
}

func TestInternetBillToProto(t *testing.T) {
	bill := &db.InternetBill{
		CustomerID: "CUST-001", Name: "John", Address: "Jl. Test",
		PhoneNumber: "08123", Code: "BZN", From: "Jan", To: "Feb",
		InternetFee: "250000", Tax: "25000", Total: "275000",
	}
	result := internetBillToProto(bill)
	assert.Equal(t, "CUST-001", result.CustomerId)
	assert.Equal(t, "275000", result.Total)
}

func TestCurrenciesToProto(t *testing.T) {
	currencies := []db.Currency{{Code: "USD", Label: "US Dollar", Rate: 15800.50}}
	result := currenciesToProto(currencies)
	assert.Len(t, result, 1)
	assert.Equal(t, "USD", result[0].Code)
}

func TestBeneficiariesToProto(t *testing.T) {
	beneficiaries := []db.Beneficiary{
		{ID: "b-1", Name: "Alice", Phone: "08123", Avatar: "url"},
	}
	result := beneficiariesToProto(beneficiaries)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice", result[0].Name)
}

func TestTransactionToProto(t *testing.T) {
	txn := &db.PrepaidTransaction{
		ID: "txn-1", Status: "SUCCESS", Message: "OK", Timestamp: "2026-04-01",
	}
	result := transactionToProto(txn)
	assert.Equal(t, "txn-1", result.Id)
	assert.Equal(t, "SUCCESS", result.Status)
}

func TestProvidersToProto_Empty(t *testing.T) {
	result := providersToProto([]db.ServiceProvider{})
	assert.Empty(t, result)
}

func TestCurrenciesToProto_Empty(t *testing.T) {
	result := currenciesToProto([]db.Currency{})
	assert.Empty(t, result)
}

func TestBeneficiariesToProto_Empty(t *testing.T) {
	result := beneficiariesToProto([]db.Beneficiary{})
	assert.Empty(t, result)
}
