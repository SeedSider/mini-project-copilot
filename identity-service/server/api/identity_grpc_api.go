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

	// 5. Return response (no best-effort profile creation — BFF handles orchestration)
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
