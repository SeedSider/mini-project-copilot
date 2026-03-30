package api

import (
	"context"
	"fmt"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	profilePB "github.com/bankease/bff-service/protogen/user-profile-service"
	manager "github.com/bankease/bff-service/server/jwt"
	"github.com/bankease/bff-service/server/utils"
)

// GetMyProfile gets profile by user_id extracted from JWT claims.
func (s *Server) GetMyProfile(ctx context.Context, _ *pb.GetMyProfileRequest) (*pb.ProfileResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetMyProfile", "Processing get my profile request", nil, nil, nil, nil)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, s.unauthorizedError()
	}

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	profileResp, err := s.svcConn.UserProfileClient().GetProfileByUserID(outCtx, &profilePB.GetProfileByUserIDRequest{
		UserId: claims.UserID,
	})
	if err != nil {
		log.Error(processId, "GetMyProfile", fmt.Sprintf("profile.GetProfileByUserID failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return toProfileResponse(profileResp), nil
}

// GetProfileByID proxies to user-profile-service.
func (s *Server) GetProfileByID(ctx context.Context, req *pb.GetProfileByIDRequest) (*pb.ProfileResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetProfileByID", "Processing get profile by ID request", nil, nil, nil, nil)

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	profileResp, err := s.svcConn.UserProfileClient().GetProfileByID(outCtx, &profilePB.GetProfileByIDRequest{
		Id: req.GetId(),
	})
	if err != nil {
		log.Error(processId, "GetProfileByID", fmt.Sprintf("profile.GetProfileByID failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return toProfileResponse(profileResp), nil
}

// GetProfileByUserID proxies to user-profile-service.
func (s *Server) GetProfileByUserID(ctx context.Context, req *pb.GetProfileByUserIDRequest) (*pb.ProfileResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetProfileByUserID", "Processing get profile by user ID request", nil, nil, nil, nil)

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	profileResp, err := s.svcConn.UserProfileClient().GetProfileByUserID(outCtx, &profilePB.GetProfileByUserIDRequest{
		UserId: req.GetUserId(),
	})
	if err != nil {
		log.Error(processId, "GetProfileByUserID", fmt.Sprintf("profile.GetProfileByUserID failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return toProfileResponse(profileResp), nil
}

// CreateProfile proxies to user-profile-service.
func (s *Server) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.ProfileResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "CreateProfile", "Processing create profile request", nil, nil, nil, nil)

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	profileResp, err := s.svcConn.UserProfileClient().CreateProfile(outCtx, &profilePB.CreateProfileRequest{
		UserId:       req.GetUserId(),
		Bank:         req.GetBank(),
		Branch:       req.GetBranch(),
		Name:         req.GetName(),
		CardNumber:   req.GetCardNumber(),
		CardProvider: req.GetCardProvider(),
		Balance:      req.GetBalance(),
		Currency:     req.GetCurrency(),
		AccountType:  req.GetAccountType(),
	})
	if err != nil {
		log.Error(processId, "CreateProfile", fmt.Sprintf("profile.CreateProfile failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return toProfileResponse(profileResp), nil
}

// UpdateProfile proxies to user-profile-service.
func (s *Server) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.StandardResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "UpdateProfile", "Processing update profile request", nil, nil, nil, nil)

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	resp, err := s.svcConn.UserProfileClient().UpdateProfile(outCtx, &profilePB.UpdateProfileRequest{
		Id:         req.GetId(),
		Bank:       req.GetBank(),
		Branch:     req.GetBranch(),
		Name:       req.GetName(),
		CardNumber: req.GetCardNumber(),
	})
	if err != nil {
		log.Error(processId, "UpdateProfile", fmt.Sprintf("profile.UpdateProfile failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return &pb.StandardResponse{
		Code:        resp.Code,
		Description: resp.Description,
	}, nil
}

// toProfileResponse converts user-profile-service ProfileResponse to BFF ProfileResponse.
func toProfileResponse(p *profilePB.ProfileResponse) *pb.ProfileResponse {
	return &pb.ProfileResponse{
		Id:           p.Id,
		UserId:       p.UserId,
		Bank:         p.Bank,
		Branch:       p.Branch,
		Name:         p.Name,
		CardNumber:   p.CardNumber,
		CardProvider: p.CardProvider,
		Balance:      p.Balance,
		Currency:     p.Currency,
		AccountType:  p.AccountType,
		Image:        p.Image,
	}
}
