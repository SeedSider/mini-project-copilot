package api

import (
	"context"
	"fmt"
	"regexp"
	"time"

	manager "github.com/bankease/payment-service/server/jwt"

	pb "github.com/bankease/payment-service/protogen/payment-service"
	"github.com/bankease/payment-service/server/db"
	"github.com/bankease/payment-service/server/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetProviders implements the gRPC GetProviders RPC (public).
func (s *Server) GetProviders(ctx context.Context, _ *pb.GetProvidersRequest) (*pb.ProviderListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)

	providers, err := s.provider.GetAllProviders(ctx)
	if err != nil {
		log.Error(processId, "GetProviders", err.Error(), nil, nil, nil, err)
		return nil, status.Errorf(codes.Internal, InternalServerErrorMessage)
	}

	return &pb.ProviderListResponse{
		Providers: providersToProto(providers),
	}, nil
}

// GetInternetBill implements the gRPC GetInternetBill RPC (JWT protected).
func (s *Server) GetInternetBill(ctx context.Context, _ *pb.GetInternetBillRequest) (*pb.InternetBillResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
	}

	bill, err := s.provider.GetInternetBillByUserID(ctx, claims.UserID)
	if err != nil {
		log.Error(processId, "GetInternetBill", err.Error(), nil, nil, nil, err)
		return nil, status.Errorf(codes.NotFound, "Internet bill not found")
	}

	return &pb.InternetBillResponse{
		Bill: internetBillToProto(bill),
	}, nil
}

// GetCurrencyList implements the gRPC GetCurrencyList RPC (public).
func (s *Server) GetCurrencyList(ctx context.Context, _ *pb.GetCurrencyListRequest) (*pb.CurrencyListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)

	currencies, err := s.provider.GetAllCurrencies(ctx)
	if err != nil {
		log.Error(processId, "GetCurrencyList", err.Error(), nil, nil, nil, err)
		return nil, status.Errorf(codes.Internal, InternalServerErrorMessage)
	}

	return &pb.CurrencyListResponse{
		Currencies: currenciesToProto(currencies),
	}, nil
}

var grpcPhoneRegex = regexp.MustCompile(`^\d{10,13}$`)

// GetBeneficiaries implements the gRPC GetBeneficiaries RPC (JWT protected).
func (s *Server) GetBeneficiaries(ctx context.Context, req *pb.GetBeneficiariesRequest) (*pb.BeneficiaryListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
	}

	accountID := req.GetAccountId()
	if accountID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "accountId is required")
	}

	beneficiaries, err := s.provider.GetBeneficiariesByAccountID(ctx, accountID)
	if err != nil {
		log.Error(processId, "GetBeneficiaries", err.Error(), nil, nil, nil, err)
		return nil, status.Errorf(codes.Internal, InternalServerErrorMessage)
	}

	if beneficiaries == nil {
		beneficiaries = []db.Beneficiary{}
	}

	return &pb.BeneficiaryListResponse{
		Beneficiaries: beneficiariesToProto(beneficiaries),
	}, nil
}

// PrepaidPay implements the gRPC PrepaidPay RPC (JWT protected).
func (s *Server) PrepaidPay(ctx context.Context, req *pb.PrepaidPayRequest) (*pb.PrepaidPayResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
	}

	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		return nil, status.Errorf(codes.InvalidArgument, "idempotencyKey is required")
	}
	if req.GetCardId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "cardId is required")
	}
	if req.GetPhone() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "phone is required")
	}
	if !grpcPhoneRegex.MatchString(req.GetPhone()) {
		return nil, status.Errorf(codes.InvalidArgument, "phone must be 10-13 digits")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be greater than 0")
	}

	// Idempotency check
	existing, err := s.provider.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		log.Error(processId, "PrepaidPay", err.Error(), nil, nil, nil, err)
		return nil, status.Errorf(codes.Internal, InternalServerErrorMessage)
	}
	if existing != nil {
		return transactionToProto(existing), nil
	}

	txnID := fmt.Sprintf("txn-%d", time.Now().UnixMilli())
	txn := db.PrepaidTransaction{
		ID:             txnID,
		CardID:         req.GetCardId(),
		Phone:          req.GetPhone(),
		Amount:         req.GetAmount(),
		Status:         "SUCCESS",
		Message:        "Top-up successful",
		IdempotencyKey: idempotencyKey,
	}

	result, err := s.provider.CreatePrepaidTransaction(ctx, txn)
	if err != nil {
		log.Error(processId, "PrepaidPay", err.Error(), nil, nil, nil, err)
		return nil, status.Errorf(codes.Internal, InternalServerErrorMessage)
	}

	return transactionToProto(result), nil
}
