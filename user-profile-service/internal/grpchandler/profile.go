package grpchandler

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"github.com/bankease/user-profile-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcServer implements the UserProfileServiceServer interface.
type GrpcServer struct {
	pb.UnimplementedUserProfileServiceServer
	ProfileRepo      *repository.ProfileRepository
	MenuRepo         *repository.MenuRepository
	ExchangeRateRepo *repository.ExchangeRateRepository
	InterestRateRepo *repository.InterestRateRepository
	BranchRepo       *repository.BranchRepository
}

// CreateProfile implements the gRPC CreateProfile RPC.
func (s *GrpcServer) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.ProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	modelReq := createReqToModel(req)
	profile, err := s.ProfileRepo.CreateProfile(ctx, modelReq)
	if err != nil {
		log.Printf("gRPC CreateProfile error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create profile")
	}

	return profileToProto(profile), nil
}

// GetProfileByID implements the gRPC GetProfileByID RPC.
func (s *GrpcServer) GetProfileByID(ctx context.Context, req *pb.GetProfileByIDRequest) (*pb.ProfileResponse, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	profile, err := s.ProfileRepo.GetProfileByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "profile not found")
		}
		log.Printf("gRPC GetProfileByID error: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return profileToProto(profile), nil
}

// GetProfileByUserID implements the gRPC GetProfileByUserID RPC.
func (s *GrpcServer) GetProfileByUserID(ctx context.Context, req *pb.GetProfileByUserIDRequest) (*pb.ProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	profile, err := s.ProfileRepo.GetProfileByUserID(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "profile not found")
		}
		log.Printf("gRPC GetProfileByUserID error: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return profileToProto(profile), nil
}

// UpdateProfile implements the gRPC UpdateProfile RPC.
func (s *GrpcServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.StandardResponse, error) {
	id, modelReq := updateReqToModel(req)
	if id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	if err := s.ProfileRepo.UpdateProfile(ctx, id, modelReq); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "profile not found")
		}
		log.Printf("gRPC UpdateProfile error: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.StandardResponse{
		Code:        int32(http.StatusOK),
		Description: "Data pengguna berhasil diubah.",
	}, nil
}
