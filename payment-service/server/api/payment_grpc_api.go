package api

import (
	"context"

	manager "github.com/bankease/payment-service/server/jwt"

	pb "github.com/bankease/payment-service/protogen/payment-service"
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
