package api

import (
	"context"
	"fmt"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	identityPB "github.com/bankease/bff-service/protogen/identity-service"
	profilePB "github.com/bankease/bff-service/protogen/user-profile-service"
	manager "github.com/bankease/bff-service/server/jwt"
	"github.com/bankease/bff-service/server/utils"
)

// SignUp orchestrates identity.SignUp → profile.CreateProfile (best-effort).
func (s *Server) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "SignUp", "Processing signup request", nil, nil, nil, nil)

	// 1. Call identity-service.SignUp
	identityReq := &identityPB.SignUpRequest{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Phone:    req.GetPhone(),
	}

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	identityResp, err := s.svcConn.IdentityClient().SignUp(outCtx, identityReq)
	if err != nil {
		log.Error(processId, "SignUp", fmt.Sprintf("identity-service.SignUp failed: %v", err), nil, nil, nil, err)
		return nil, err // propagate gRPC error as-is
	}

	// 2. Best-effort: create profile in user-profile-service
	profileReq := &profilePB.CreateProfileRequest{
		UserId:       identityResp.UserId,
		Bank:         "BRI",
		Branch:       "Jakarta",
		Name:         req.GetUsername(),
		CardNumber:   "",
		CardProvider: "",
		Balance:      0,
		Currency:     "IDR",
		AccountType:  "REGULAR",
	}

	_, profileErr := s.svcConn.UserProfileClient().CreateProfile(outCtx, profileReq)
	if profileErr != nil {
		// Best-effort: log error but don't fail the signup
		log.Error(processId, "SignUp", fmt.Sprintf("profile.CreateProfile best-effort failed: %v", profileErr), nil, nil, nil, profileErr)
	}

	log.Info(processId, "SignUp", fmt.Sprintf("User registered: %s", identityResp.UserId), nil, nil, nil, nil)
	return &pb.SignUpResponse{
		UserId:   identityResp.UserId,
		Username: identityResp.Username,
	}, nil
}

// SignIn proxies to identity-service.SignIn.
func (s *Server) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "SignIn", "Processing signin request", nil, nil, nil, nil)

	identityReq := &identityPB.SignInRequest{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	}

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	identityResp, err := s.svcConn.IdentityClient().SignIn(outCtx, identityReq)
	if err != nil {
		log.Error(processId, "SignIn", fmt.Sprintf("identity-service.SignIn failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return &pb.SignInResponse{
		UserId:   identityResp.UserId,
		Username: identityResp.Username,
		Token:    identityResp.Token,
	}, nil
}

// GetMe proxies to identity-service.GetMe using JWT claims from context.
func (s *Server) GetMe(ctx context.Context, _ *pb.GetMeRequest) (*pb.GetMeResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetMe", "Processing get me request", nil, nil, nil, nil)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, s.unauthorizedError()
	}

	return &pb.GetMeResponse{
		UserId:   claims.UserID,
		Username: claims.Username,
	}, nil
}
