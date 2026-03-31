package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	manager "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/jwt"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/utils"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//#region SignUp

type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type SignUpResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type createProfileRequest struct {
	UserID       string `json:"user_id"`
	Bank         string `json:"bank"`
	Branch       string `json:"branch"`
	Name         string `json:"name"`
	CardNumber   string `json:"card_number"`
	CardProvider string `json:"card_provider"`
	Balance      int64  `json:"balance"`
	Currency     string `json:"currency"`
	AccountType  string `json:"accountType"`
	Image        string `json:"image"`
}

// createProfileBestEffort calls the profile service to create a profile for a newly registered user.
// Errors are logged but do not affect the sign-up response.
func (s *Server) createProfileBestEffort(userID, username string) {
	if s.profileServiceURL == "" {
		return
	}

	reqBody := createProfileRequest{
		UserID:       userID,
		Bank:         "BRI",
		Branch:       "PUSAT",
		Name:         username,
		CardNumber:   utils.GenerateCardNumber(),
		CardProvider: utils.GenerateCardProvider(),
		Balance:      0,
		Currency:     "IDR",
		AccountType:  "REGULAR",
		Image:        "",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		log.Error("", "createProfileBestEffort", fmt.Sprintf("failed to marshal profile request: %v", err), nil, nil, nil, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.profileServiceURL+"/api/profile", bytes.NewReader(body))
	if err != nil {
		log.Error("", "createProfileBestEffort", fmt.Sprintf("failed to build profile request: %v", err), nil, nil, nil, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("", "createProfileBestEffort", fmt.Sprintf("failed to call profile service: %v", err), nil, nil, nil, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Error("", "createProfileBestEffort", fmt.Sprintf("profile service returned status %d", resp.StatusCode), nil, nil, nil, nil)
	}
}

func (s *Server) httpSignUp(ctx context.Context, req *SignUpRequest) (*SignUpResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "httpSignUp", "Processing signup request", nil, nil, nil, nil)

	// 1. Validate input
	if err := validateSignUpRequest(req); err != nil {
		log.Error(processId, "SignUp", fmt.Sprintf("[error][api][func: SignUp] validation: %v", err), nil, nil, nil, err)
		return nil, s.badRequestError(err.Error())
	}

	// 2. Check username exists
	exists, err := s.provider.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		log.Error(processId, "SignUp", fmt.Sprintf("[error][api][func: SignUp] check username: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}
	if exists {
		log.Info(processId, "SignUp", "Username already registered", nil, nil, nil, nil)
		return nil, s.conflictError("Username already registered")
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(processId, "SignUp", fmt.Sprintf("[error][api][func: SignUp] hash password: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 4. Insert user
	result, err := s.provider.CreateUser(ctx, req.Username, string(hashedPassword), req.Phone)
	if err != nil {
		log.Error(processId, "httpSignUp", fmt.Sprintf("[error][api][func: httpSignUp] create user: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 5. Best-effort: create profile for the new user with a randomized card number
	go s.createProfileBestEffort(result.UserID, result.Username)

	log.Info(processId, "httpSignUp", fmt.Sprintf("User registered: %s", result.UserID), nil, nil, nil, nil)
	return &SignUpResponse{
		UserID:   result.UserID,
		Username: result.Username,
	}, nil
}

//#endregion

//#region SignIn

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignInResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (s *Server) httpSignIn(ctx context.Context, req *SignInRequest) (*SignInResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "httpSignIn", "Processing signin request", nil, nil, nil, nil)

	// 1. Validate input
	if req.Username == "" || req.Password == "" {
		return nil, s.badRequestError("Username and password are required")
	}

	// 2. Find user by username
	user, err := s.provider.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info(processId, "SignIn", "Invalid credentials - user not found", nil, nil, nil, nil)
			return nil, s.unauthorizedError()
		}
		log.Error(processId, "SignIn", fmt.Sprintf("[error][api][func: SignIn] get user: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 3. Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		log.Info(processId, "SignIn", "Invalid credentials - password mismatch", nil, nil, nil, nil)
		return nil, s.unauthorizedError()
	}

	// 4. Get profile
	// (profiles table removed — user data is in users table only)

	// 5. Generate JWT
	token, err := s.manager.Generate(user.ID, user.Username)
	if err != nil {
		log.Error(processId, "SignIn", fmt.Sprintf("[error][api][func: SignIn] generate token: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 6. Return response
	log.Info(processId, "SignIn", fmt.Sprintf("User signed in: %s", user.ID), nil, nil, nil, nil)
	return &SignInResponse{
		UserID:   user.ID,
		Username: user.Username,
		Token:    token,
	}, nil
}

//#endregion

//#region GetMe

type GetMeResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func (s *Server) httpGetMe(ctx context.Context) (*GetMeResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "httpGetMe", "Processing get me request", nil, nil, nil, nil)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, s.unauthorizedError()
	}

	profile, err := s.provider.GetUserByUsername(ctx, claims.Username)
	if err != nil {
		log.Error(processId, "GetMe", fmt.Sprintf("[error][api][func: GetMe] get user: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	return &GetMeResponse{
		UserID:   claims.UserID,
		Username: profile.Username,
	}, nil
}

//#endregion

//#region HTTP Handlers

func (s *Server) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	processId := utils.GenerateProcessId()
	ctx := context.WithValue(r.Context(), "process_id", processId)

	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := s.httpSignUp(ctx, &req)
	if err != nil {
		writeGrpcErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusCreated, resp)
}

func (s *Server) HandleSignIn(w http.ResponseWriter, r *http.Request) {
	processId := utils.GenerateProcessId()
	ctx := context.WithValue(r.Context(), "process_id", processId)

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := s.httpSignIn(ctx, &req)
	if err != nil {
		writeGrpcErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

func (s *Server) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	processId := utils.GenerateProcessId()
	ctx := context.WithValue(r.Context(), "process_id", processId)

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "Authorization token is required")
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := s.manager.Verify(token)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	ctx = context.WithValue(ctx, "user_claims", claims)

	resp, svcErr := s.httpGetMe(ctx)
	if svcErr != nil {
		writeGrpcErrorResponse(w, svcErr)
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

//#endregion

//#region Helpers

func validateSignUpRequest(req *SignUpRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	return nil
}

type errorResponse struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	contentTypeJSON = "application/json"
	contentTypeHeader = "Content-Type"
)
func writeJSONResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func writeErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&errorResponse{
		Error:   true,
		Code:    code,
		Message: message,
	})
}

func writeGrpcErrorResponse(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeErrorResponse(w, http.StatusInternalServerError, "Internal error")
		return
	}
	httpCode := grpcToHTTPCode(st.Code())
	writeErrorResponse(w, httpCode, st.Message())
}

func grpcToHTTPCode(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusUnprocessableEntity
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.NotFound:
		return http.StatusNotFound
	case codes.PermissionDenied:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

//#endregion
