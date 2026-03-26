package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	manager "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/jwt"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/utils"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//#region SignUp

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type SignUpResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

func (s *Server) SignUp(ctx context.Context, req *SignUpRequest) (*SignUpResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "SignUp", "Processing signup request", nil, nil, nil, nil)

	// 1. Validate input
	if err := validateSignUpRequest(req); err != nil {
		log.Error(processId, "SignUp", fmt.Sprintf("[error][api][func: SignUp] validation: %v", err), nil, nil, nil, err)
		return nil, s.badRequestError(err.Error())
	}

	// 2. Check email exists
	exists, err := s.provider.CheckEmailExists(ctx, req.Email)
	if err != nil {
		log.Error(processId, "SignUp", fmt.Sprintf("[error][api][func: SignUp] check email: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}
	if exists {
		log.Info(processId, "SignUp", "Email already registered", nil, nil, nil, nil)
		return nil, s.conflictError("Email already registered")
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(processId, "SignUp", fmt.Sprintf("[error][api][func: SignUp] hash password: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 4-5. Insert user + profile (transactional)
	result, err := s.provider.CreateUserWithProfile(ctx, req.Email, string(hashedPassword), req.FullName, req.Phone)
	if err != nil {
		log.Error(processId, "SignUp", fmt.Sprintf("[error][api][func: SignUp] create user: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 6. Return response
	log.Info(processId, "SignUp", fmt.Sprintf("User registered: %s", result.UserID), nil, nil, nil, nil)
	return &SignUpResponse{
		UserID:   result.UserID,
		Email:    result.Email,
		FullName: result.FullName,
	}, nil
}

//#endregion

//#region SignIn

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Token    string `json:"token"`
}

func (s *Server) SignIn(ctx context.Context, req *SignInRequest) (*SignInResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "SignIn", "Processing signin request", nil, nil, nil, nil)

	// 1. Validate input
	if req.Email == "" || req.Password == "" {
		return nil, s.badRequestError("Email and password are required")
	}

	// 2. Find user by email
	user, err := s.provider.GetUserByEmail(ctx, req.Email)
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
	profile, err := s.provider.GetProfileByUserID(ctx, user.ID)
	if err != nil {
		log.Error(processId, "SignIn", fmt.Sprintf("[error][api][func: SignIn] get profile: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 5. Generate JWT
	token, err := s.manager.Generate(user.ID, user.Email, profile.FullName)
	if err != nil {
		log.Error(processId, "SignIn", fmt.Sprintf("[error][api][func: SignIn] generate token: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	// 6. Return response
	log.Info(processId, "SignIn", fmt.Sprintf("User signed in: %s", user.ID), nil, nil, nil, nil)
	return &SignInResponse{
		UserID:   user.ID,
		Email:    user.Email,
		FullName: profile.FullName,
		Token:    token,
	}, nil
}

//#endregion

//#region GetMe

type GetMeResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone,omitempty"`
}

func (s *Server) GetMe(ctx context.Context) (*GetMeResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetMe", "Processing get me request", nil, nil, nil, nil)

	claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		return nil, s.unauthorizedError()
	}

	profile, err := s.provider.GetProfileByUserID(ctx, claims.UserID)
	if err != nil {
		log.Error(processId, "GetMe", fmt.Sprintf("[error][api][func: GetMe] get profile: %v", err), nil, nil, nil, err)
		return nil, s.serverError()
	}

	phone := ""
	if profile.Phone != nil {
		phone = *profile.Phone
	}

	return &GetMeResponse{
		UserID:   claims.UserID,
		Email:    claims.Email,
		FullName: profile.FullName,
		Phone:    phone,
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

	resp, err := s.SignUp(ctx, &req)
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

	resp, err := s.SignIn(ctx, &req)
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

	resp, svcErr := s.GetMe(ctx)
	if svcErr != nil {
		writeGrpcErrorResponse(w, svcErr)
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

//#endregion

//#region Helpers

func validateSignUpRequest(req *SignUpRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if req.FullName == "" {
		return fmt.Errorf("full_name is required")
	}
	return nil
}

type errorResponse struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeJSONResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func writeErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
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
