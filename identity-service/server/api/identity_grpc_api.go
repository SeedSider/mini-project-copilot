package api

import (
	"context"
	"database/sql"
	"fmt"

	pb "bitbucket.bri.co.id/scm/addons/addons-identity-service/protogen/identity-service"
	manager "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/jwt"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/utils"
	"golang.org/x/crypto/bcrypt"
)

// Ensure Server implements IdentityServiceServer
var _ pb.IdentityServiceServer = (*Server)(nil)

// SignUp implements the gRPC SignUp RPC method.
func (s *Server) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "SignUp", "Processing gRPC signup request", nil, nil, nil, nil)

	// 1. Validate input
	if req.GetUsername() == "" {
		return nil, s.badRequestError("username is required")
	}
	if len(req.GetPassword()) < 6 {
		return nil, s.badRequestError("password must be at least 6 characters")
	}

	// 2. Check username exists
	exists, err := s.provider.CheckUsernameExists(ctx, req.GetUsername())
	if err != nil {
		log.Error(processId, "GrpcSignUp", fmt.Sprintf("[error][api][func: GrpcSignUp] check username: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}
	if exists {
		log.Info(processId, "GrpcSignUp", "Username already registered", nil, nil, nil, nil)
		return nil, s.conflictError("Username already registered")
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		log.Error(processId, "GrpcSignUp", fmt.Sprintf("[error][api][func: GrpcSignUp] hash password: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 4. Insert user
	result, err := s.provider.CreateUser(ctx, req.GetUsername(), string(hashedPassword), req.GetPhone())
	if err != nil {
		log.Error(processId, "GrpcSignUp", fmt.Sprintf("[error][api][func: GrpcSignUp] create user: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 5. Best-effort: create profile for the new user with a randomized card number
	go s.createProfileBestEffort(result.UserID, result.Username)

	// 6. Return response
	log.Info(processId, "GrpcSignUp", fmt.Sprintf("User registered: %s", result.UserID), nil, nil, nil, nil)
	return &pb.SignUpResponse{
		UserId:   result.UserID,
		Username: result.Username,
	}, nil
}

// SignIn implements the gRPC SignIn RPC method.
func (s *Server) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "SignIn", "Processing gRPC signin request", nil, nil, nil, nil)

	// 1. Validate input
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, s.badRequestError("Username and password are required")
	}

	// 2. Find user by username
	user, err := s.provider.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info(processId, "GrpcSignIn", "Invalid credentials - user not found", nil, nil, nil, nil)
			return nil, s.unauthorizedError()
		}
		log.Error(processId, "GrpcSignIn", fmt.Sprintf("[error][api][func: GrpcSignIn] get user: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 3. Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.GetPassword()))
	if err != nil {
		log.Info(processId, "GrpcSignIn", "Invalid credentials - password mismatch", nil, nil, nil, nil)
		return nil, s.unauthorizedError()
	}

	// 4. Generate JWT
	token, err := s.manager.Generate(user.ID, user.Username)
	if err != nil {
		log.Error(processId, "GrpcSignIn", fmt.Sprintf("[error][api][func: GrpcSignIn] generate token: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 5. Return response
	log.Info(processId, "GrpcSignIn", fmt.Sprintf("User signed in: %s", user.ID), nil, nil, nil, nil)
	return &pb.SignInResponse{
		UserId:   user.ID,
		Username: user.Username,
		Token:    token,
	}, nil
}

// GetMe implements the gRPC GetMe RPC method.
func (s *Server) GetMe(ctx context.Context, _ *pb.GetMeRequest) (*pb.GetMeResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetMe", "Processing gRPC get me request", nil, nil, nil, nil)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, s.unauthorizedError()
	}

	user, err := s.provider.GetUserByUsername(ctx, claims.Username)
	if err != nil {
		log.Error(processId, "GrpcGetMe", fmt.Sprintf("[error][api][func: GrpcGetMe] get user: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	return &pb.GetMeResponse{
		UserId:   claims.UserID,
		Username: user.Username,
	}, nil
}

// ValidateOtp implements the gRPC ValidateOtp RPC method.
func (s *Server) ValidateOtp(ctx context.Context, req *pb.ValidateOtpRequest) (*pb.ValidateOtpResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "ValidateOtp", "Processing gRPC validate OTP request", nil, nil, nil, nil)

	if req.GetUsername() == "" {
		return nil, s.badRequestError("username is required")
	}

	exists, err := s.provider.CheckUsernameExists(ctx, req.GetUsername())
	if err != nil {
		log.Error(processId, "GrpcValidateOtp", fmt.Sprintf("[error][api][func: GrpcValidateOtp] check username: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}
	if !exists {
		return nil, s.notFoundError("Username not found")
	}

	otp, err := generateOTP()
	if err != nil {
		log.Error(processId, "GrpcValidateOtp", fmt.Sprintf("[error][api][func: GrpcValidateOtp] generate OTP: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	log.Info(processId, "GrpcValidateOtp", fmt.Sprintf("OTP generated for user: %s", req.GetUsername()), nil, nil, nil, nil)
	return &pb.ValidateOtpResponse{Otp: otp}, nil
}

// UpdatePassword implements the gRPC UpdatePassword RPC method.
func (s *Server) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "UpdatePassword", "Processing gRPC update password request", nil, nil, nil, nil)

	if req.GetUsername() == "" {
		return nil, s.badRequestError("username is required")
	}
	if len(req.GetNewPassword()) < 6 {
		return nil, s.badRequestError("password must be at least 6 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetNewPassword()), bcrypt.DefaultCost)
	if err != nil {
		log.Error(processId, "GrpcUpdatePassword", fmt.Sprintf("[error][api][func: GrpcUpdatePassword] hash password: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	err = s.provider.UpdatePasswordByUsername(ctx, req.GetUsername(), string(hashedPassword))
	if err != nil {
		log.Error(processId, "GrpcUpdatePassword", fmt.Sprintf("[error][api][func: GrpcUpdatePassword] update password: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	log.Info(processId, "GrpcUpdatePassword", fmt.Sprintf("Password updated for user: %s", req.GetUsername()), nil, nil, nil, nil)
	return &pb.UpdatePasswordResponse{Message: "berhasil ubah password"}, nil
}
