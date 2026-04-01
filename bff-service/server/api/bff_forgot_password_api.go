package api

import (
	"context"
	"fmt"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	identityPB "github.com/bankease/bff-service/protogen/identity-service"
	manager "github.com/bankease/bff-service/server/jwt"
	"github.com/bankease/bff-service/server/utils"
)

// ValidateOtp proxies to identity-service.ValidateOtp (public, no JWT required).
func (s *Server) ValidateOtp(ctx context.Context, req *pb.ValidateOtpRequest) (*pb.ValidateOtpResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "ValidateOtp", "Processing validate OTP request", nil, nil, nil, nil)

	identityReq := &identityPB.ValidateOtpRequest{
		Username: req.GetUsername(),
	}

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	identityResp, err := s.svcConn.IdentityClient().ValidateOtp(outCtx, identityReq)
	if err != nil {
		log.Error(processId, "ValidateOtp", fmt.Sprintf("identity-service.ValidateOtp failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return &pb.ValidateOtpResponse{
		Otp: identityResp.Otp,
	}, nil
}

// UpdatePassword extracts username from JWT claims and proxies to identity-service.UpdatePassword.
func (s *Server) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "UpdatePassword", "Processing update password request", nil, nil, nil, nil)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, s.unauthorizedError()
	}

	identityReq := &identityPB.UpdatePasswordRequest{
		Username:    claims.Username,
		NewPassword: req.GetNewPassword(),
	}

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	identityResp, err := s.svcConn.IdentityClient().UpdatePassword(outCtx, identityReq)
	if err != nil {
		log.Error(processId, "UpdatePassword", fmt.Sprintf("identity-service.UpdatePassword failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return &pb.UpdatePasswordResponse{
		Message: identityResp.Message,
	}, nil
}
