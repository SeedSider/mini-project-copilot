package api

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProfile implements the gRPC CreateProfile RPC.
func (s *Server) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.ProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	modelReq := createReqToModel(req)
	profile, err := s.provider.CreateProfile(ctx, modelReq)
	if err != nil {
		log.Printf("gRPC CreateProfile error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create profile")
	}

	return profileToProto(profile), nil
}

// GetProfileByID implements the gRPC GetProfileByID RPC.
func (s *Server) GetProfileByID(ctx context.Context, req *pb.GetProfileByIDRequest) (*pb.ProfileResponse, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	profile, err := s.provider.GetProfileByID(ctx, req.GetId())
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
func (s *Server) GetProfileByUserID(ctx context.Context, req *pb.GetProfileByUserIDRequest) (*pb.ProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	profile, err := s.provider.GetProfileByUserID(ctx, req.GetUserId())
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
func (s *Server) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.StandardResponse, error) {
	id, modelReq := updateReqToModel(req)
	if id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	if err := s.provider.UpdateProfile(ctx, id, modelReq); err != nil {
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
