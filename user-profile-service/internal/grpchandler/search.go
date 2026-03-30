package grpchandler

import (
	"context"
	"log"

	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetExchangeRates implements the gRPC GetExchangeRates RPC.
func (s *GrpcServer) GetExchangeRates(ctx context.Context, _ *pb.GetExchangeRatesRequest) (*pb.ExchangeRateListResponse, error) {
	rates, err := s.ExchangeRateRepo.GetAll(ctx)
	if err != nil {
		log.Printf("gRPC GetExchangeRates error: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.ExchangeRateListResponse{
		ExchangeRates: exchangeRatesToProto(rates),
	}, nil
}

// GetInterestRates implements the gRPC GetInterestRates RPC.
func (s *GrpcServer) GetInterestRates(ctx context.Context, _ *pb.GetInterestRatesRequest) (*pb.InterestRateListResponse, error) {
	rates, err := s.InterestRateRepo.GetAll(ctx)
	if err != nil {
		log.Printf("gRPC GetInterestRates error: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.InterestRateListResponse{
		InterestRates: interestRatesToProto(rates),
	}, nil
}

// GetBranches implements the gRPC GetBranches RPC.
// If req.Query is empty, returns all branches; otherwise filters by name.
func (s *GrpcServer) GetBranches(ctx context.Context, req *pb.GetBranchesRequest) (*pb.BranchListResponse, error) {
	q := req.GetQuery()

	if q == "" {
		items, err := s.BranchRepo.GetAll(ctx)
		if err != nil {
			log.Printf("gRPC GetBranches error: %v", err)
			return nil, status.Errorf(codes.Internal, "internal server error")
		}
		return &pb.BranchListResponse{Branches: branchesToProto(items)}, nil
	}

	items, err := s.BranchRepo.SearchByName(ctx, q)
	if err != nil {
		log.Printf("gRPC GetBranches error: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
	return &pb.BranchListResponse{Branches: branchesToProto(items)}, nil
}
