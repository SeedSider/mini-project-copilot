package grpchandler

import (
	"context"
	"log"

	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetAllMenus implements the gRPC GetAllMenus RPC.
func (s *GrpcServer) GetAllMenus(ctx context.Context, _ *pb.GetAllMenusRequest) (*pb.MenuListResponse, error) {
	menus, err := s.MenuRepo.GetAllMenus(ctx)
	if err != nil {
		log.Printf("gRPC GetAllMenus error: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.MenuListResponse{
		Menus: menusToProto(menus),
	}, nil
}

// GetMenusByAccountType implements the gRPC GetMenusByAccountType RPC.
func (s *GrpcServer) GetMenusByAccountType(ctx context.Context, req *pb.GetMenusByAccountTypeRequest) (*pb.MenuListResponse, error) {
	if req.GetAccountType() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "account_type is required")
	}

	menus, err := s.MenuRepo.GetMenusByAccountType(ctx, req.GetAccountType())
	if err != nil {
		log.Printf("gRPC GetMenusByAccountType error: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.MenuListResponse{
		Menus: menusToProto(menus),
	}, nil
}
