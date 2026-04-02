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

// ═══════════════════════════════════════════════════════════
// gRPC AddBeneficiary
// ═══════════════════════════════════════════════════════════

func TestGRPC_AddBeneficiary_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery("INSERT INTO beneficiary").
		WithArgs("acc-1", "Rina", "085678901234", "").
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "name", "phone", "avatar"}).
			AddRow("ben-003", "acc-1", "Rina", "085678901234", ""))

	resp, err := srv.AddBeneficiary(ctxWithClaims(), &pb.AddBeneficiaryRequest{
		AccountId: "acc-1", Name: "Rina", Phone: "085678901234",
	})
	require.NoError(t, err)
	assert.Equal(t, "Rina", resp.Name)
	assert.Equal(t, "ben-003", resp.Id)
}

func TestGRPC_AddBeneficiary_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.AddBeneficiary(context.Background(), &pb.AddBeneficiaryRequest{
		AccountId: "acc-1", Name: "X", Phone: "081234567890",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPC_AddBeneficiary_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *pb.AddBeneficiaryRequest
	}{
		{"no accountId", &pb.AddBeneficiaryRequest{Name: "X", Phone: "081234567890"}},
		{"no name", &pb.AddBeneficiaryRequest{AccountId: "a", Phone: "081234567890"}},
		{"no phone", &pb.AddBeneficiaryRequest{AccountId: "a", Name: "X"}},
		{"bad phone", &pb.AddBeneficiaryRequest{AccountId: "a", Name: "X", Phone: "12"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			_, err := srv.AddBeneficiary(ctxWithClaims(), tc.req)
			st, _ := status.FromError(err)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		})
	}
}

func TestGRPC_AddBeneficiary_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery("INSERT INTO beneficiary").WillReturnError(errors.New("db error"))

	_, err := srv.AddBeneficiary(ctxWithClaims(), &pb.AddBeneficiaryRequest{
		AccountId: "acc-1", Name: "X", Phone: "081234567890",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════════════════════
// gRPC SearchBeneficiaries
// ═══════════════════════════════════════════════════════════

func TestGRPC_SearchBeneficiaries_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	rows := sqlmock.NewRows([]string{"id", "account_id", "name", "phone", "avatar"}).
		AddRow("b-1", "acc-1", "Emma", "081234567890", "")
	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WithArgs("acc-1", "%Emma%").
		WillReturnRows(rows)

	resp, err := srv.SearchBeneficiaries(ctxWithClaims(), &pb.SearchBeneficiariesRequest{
		AccountId: "acc-1", Query: "Emma",
	})
	require.NoError(t, err)
	assert.Len(t, resp.Beneficiaries, 1)
	assert.Equal(t, "Emma", resp.Beneficiaries[0].Name)
}

func TestGRPC_SearchBeneficiaries_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.SearchBeneficiaries(context.Background(), &pb.SearchBeneficiariesRequest{
		AccountId: "a", Query: "q",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPC_SearchBeneficiaries_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *pb.SearchBeneficiariesRequest
	}{
		{"no accountId", &pb.SearchBeneficiariesRequest{Query: "q"}},
		{"no query", &pb.SearchBeneficiariesRequest{AccountId: "a"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			_, err := srv.SearchBeneficiaries(ctxWithClaims(), tc.req)
			st, _ := status.FromError(err)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		})
	}
}

func TestGRPC_SearchBeneficiaries_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WillReturnError(errors.New("db error"))

	_, err := srv.SearchBeneficiaries(ctxWithClaims(), &pb.SearchBeneficiariesRequest{
		AccountId: "a", Query: "q",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGRPC_SearchBeneficiaries_EmptyResult(t *testing.T) {
	srv, mock := newTestServer(t)
	rows := sqlmock.NewRows([]string{"id", "account_id", "name", "phone", "avatar"})
	mock.ExpectQuery("SELECT id, account_id, name, phone, avatar FROM beneficiary").
		WithArgs("acc-1", "%zzz%").
		WillReturnRows(rows)

	resp, err := srv.SearchBeneficiaries(ctxWithClaims(), &pb.SearchBeneficiariesRequest{
		AccountId: "acc-1", Query: "zzz",
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Beneficiaries)
}

// ═══════════════════════════════════════════════════════════
// gRPC GetPaymentCards
// ═══════════════════════════════════════════════════════════

func TestGRPC_GetPaymentCards_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	rows := sqlmock.NewRows([]string{"id", "account_id", "holder_name", "card_label", "masked_number", "balance", "currency", "brand", "gradient_colors"}).
		AddRow("card-001", "acc-1", "John", "Primary", "****1234", int64(500000), "USD", "VISA", "{#aaa,#bbb}")
	mock.ExpectQuery("SELECT id, account_id, holder_name").
		WithArgs("acc-1").
		WillReturnRows(rows)

	resp, err := srv.GetPaymentCards(ctxWithClaims(), &pb.GetPaymentCardsRequest{AccountId: "acc-1"})
	require.NoError(t, err)
	assert.Len(t, resp.Cards, 1)
	assert.Equal(t, "John", resp.Cards[0].HolderName)
}

func TestGRPC_GetPaymentCards_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetPaymentCards(context.Background(), &pb.GetPaymentCardsRequest{AccountId: "a"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPC_GetPaymentCards_MissingAccountId(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetPaymentCards(ctxWithClaims(), &pb.GetPaymentCardsRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPC_GetPaymentCards_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery("SELECT id, account_id, holder_name").
		WillReturnError(errors.New("db error"))

	_, err := srv.GetPaymentCards(ctxWithClaims(), &pb.GetPaymentCardsRequest{AccountId: "a"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGRPC_GetPaymentCards_EmptyList(t *testing.T) {
	srv, mock := newTestServer(t)
	rows := sqlmock.NewRows([]string{"id", "account_id", "holder_name", "card_label", "masked_number", "balance", "currency", "brand", "gradient_colors"})
	mock.ExpectQuery("SELECT id, account_id, holder_name").
		WithArgs("acc-1").
		WillReturnRows(rows)

	resp, err := srv.GetPaymentCards(ctxWithClaims(), &pb.GetPaymentCardsRequest{AccountId: "acc-1"})
	require.NoError(t, err)
	assert.Empty(t, resp.Cards)
}

// ═══════════════════════════════════════════════════════════
// gRPC CreatePaymentCard
// ═══════════════════════════════════════════════════════════

func TestGRPC_CreatePaymentCard_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery("INSERT INTO payment_card").
		WithArgs("acc-1", "John", "Main", "****1234", int64(100000), "USD", "VISA", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "holder_name", "card_label", "masked_number", "balance", "currency", "brand", "gradient_colors"}).
			AddRow("card-new", "acc-1", "John", "Main", "****1234", int64(100000), "USD", "VISA", "{#111,#222}"))

	resp, err := srv.CreatePaymentCard(ctxWithClaims(), &pb.CreatePaymentCardRequest{
		AccountId: "acc-1", HolderName: "John", CardLabel: "Main",
		MaskedNumber: "****1234", Balance: 100000, Currency: "USD",
		Brand: "VISA", GradientColors: []string{"#111", "#222"},
	})
	require.NoError(t, err)
	assert.Equal(t, "card-new", resp.Id)
	assert.Equal(t, "John", resp.HolderName)
}

func TestGRPC_CreatePaymentCard_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.CreatePaymentCard(context.Background(), &pb.CreatePaymentCardRequest{
		AccountId: "a", HolderName: "J",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPC_CreatePaymentCard_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *pb.CreatePaymentCardRequest
	}{
		{"no accountId", &pb.CreatePaymentCardRequest{HolderName: "J"}},
		{"no holderName", &pb.CreatePaymentCardRequest{AccountId: "a"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			_, err := srv.CreatePaymentCard(ctxWithClaims(), tc.req)
			st, _ := status.FromError(err)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		})
	}
}

func TestGRPC_CreatePaymentCard_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery("INSERT INTO payment_card").
		WillReturnError(errors.New("db error"))

	_, err := srv.CreatePaymentCard(ctxWithClaims(), &pb.CreatePaymentCardRequest{
		AccountId: "a", HolderName: "J",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════════════════════
// Converter: cardsToProto + cardToProto
// ═══════════════════════════════════════════════════════════

func TestCardsToProto(t *testing.T) {
	cards := []db.PaymentCard{
		{ID: "c1", AccountID: "a1", HolderName: "H1", Balance: 100, Currency: "USD", Brand: "VISA", GradientColors: []string{"#a"}},
		{ID: "c2", AccountID: "a2", HolderName: "H2", Balance: 200, Currency: "EUR", Brand: "MASTERCARD"},
	}
	result := cardsToProto(cards)
	assert.Len(t, result, 2)
	assert.Equal(t, "c1", result[0].Id)
	assert.Equal(t, int64(100), result[0].Balance)
	assert.Equal(t, []string{"#a"}, result[0].GradientColors)
	assert.Equal(t, "c2", result[1].Id)
}

func TestCardsToProto_Empty(t *testing.T) {
	result := cardsToProto([]db.PaymentCard{})
	assert.Empty(t, result)
}

func TestCardToProto_AllFields(t *testing.T) {
	card := &db.PaymentCard{
		ID: "c1", AccountID: "a1", HolderName: "H", CardLabel: "L",
		MaskedNumber: "****1234", Balance: 999, Currency: "USD",
		Brand: "VISA", GradientColors: []string{"#x", "#y"},
	}
	result := cardToProto(card)
	assert.Equal(t, "c1", result.Id)
	assert.Equal(t, "a1", result.AccountId)
	assert.Equal(t, "H", result.HolderName)
	assert.Equal(t, "L", result.CardLabel)
	assert.Equal(t, "****1234", result.MaskedNumber)
	assert.Equal(t, int64(999), result.Balance)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, "VISA", result.Brand)
	assert.Equal(t, []string{"#x", "#y"}, result.GradientColors)
}
