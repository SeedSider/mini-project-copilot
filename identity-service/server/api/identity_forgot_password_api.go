package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	manager "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/jwt"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/utils"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// generateOTP generates a cryptographically random 6-digit OTP (100000–999999).
var generateOTP = func() (int32, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return 0, err
	}
	return int32(n.Int64()) + 100000, nil
}

//#region ValidateOTP

type ValidateOtpRequest struct {
	Username string `json:"username"`
}

type ValidateOtpResponse struct {
	OTP int32 `json:"otp"`
}

func (s *Server) httpValidateOtp(ctx context.Context, req *ValidateOtpRequest) (*ValidateOtpResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "httpValidateOtp", "Processing validate OTP request", nil, nil, nil, nil)

	if req.Username == "" {
		return nil, s.badRequestError("username is required")
	}

	exists, err := s.provider.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		log.Error(processId, "httpValidateOtp", fmt.Sprintf("[error][api][func: httpValidateOtp] check username: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}
	if !exists {
		return nil, status.New(codes.NotFound, "Username not found").Err()
	}

	otp, err := generateOTP()
	if err != nil {
		log.Error(processId, "httpValidateOtp", fmt.Sprintf("[error][api][func: httpValidateOtp] generate OTP: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	log.Info(processId, "httpValidateOtp", fmt.Sprintf("OTP generated for user: %s", req.Username), nil, nil, nil, nil)
	return &ValidateOtpResponse{OTP: otp}, nil
}

//#endregion

//#region UpdatePassword

type UpdatePasswordRequest struct {
	NewPassword string `json:"newPassword"`
}

type UpdatePasswordResponse struct {
	Message string `json:"message"`
}

func (s *Server) httpUpdatePassword(ctx context.Context, username string, req *UpdatePasswordRequest) (*UpdatePasswordResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "httpUpdatePassword", "Processing update password request", nil, nil, nil, nil)

	if username == "" {
		return nil, s.badRequestError("username is required")
	}
	if len(req.NewPassword) < 6 {
		return nil, s.badRequestError("password must be at least 6 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error(processId, "httpUpdatePassword", fmt.Sprintf("[error][api][func: httpUpdatePassword] hash password: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	err = s.provider.UpdatePasswordByUsername(ctx, username, string(hashedPassword))
	if err != nil {
		log.Error(processId, "httpUpdatePassword", fmt.Sprintf("[error][api][func: httpUpdatePassword] update password: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	log.Info(processId, "httpUpdatePassword", fmt.Sprintf("Password updated for user: %s", username), nil, nil, nil, nil)
	return &UpdatePasswordResponse{Message: "berhasil ubah password"}, nil
}

//#endregion

//#region HTTP Handlers

func (s *Server) HandleValidateOtp(w http.ResponseWriter, r *http.Request) {
	processId := utils.GenerateProcessId()
	ctx := context.WithValue(r.Context(), "process_id", processId)

	var req ValidateOtpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := s.httpValidateOtp(ctx, &req)
	if err != nil {
		writeGrpcErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

func (s *Server) HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	processId := utils.GenerateProcessId()
	ctx := context.WithValue(r.Context(), "process_id", processId)

	// Extract username from JWT claims
	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		// Try from Authorization header (direct HTTP access)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 8 {
			writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		token := authHeader[7:] // strip "Bearer "
		var err error
		claims, err = s.manager.Verify(token)
		if err != nil {
			writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := s.httpUpdatePassword(ctx, claims.Username, &req)
	if err != nil {
		writeGrpcErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

//#endregion
