package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ── Auth endpoints ──

// handleAuthSignUp godoc
// @Summary Register a new user
// @Description Orchestrates identity-service SignUp followed by best-effort profile creation in user-profile-service
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body pb.SignUpRequest true "SignUp request body"
// @Success 201 {object} pb.SignUpResponse
// @Failure 400 {object} pb.ErrorBodyResponse
// @Failure 409 {object} pb.ErrorBodyResponse "Username already exists"
// @Failure 500 {object} pb.ErrorBodyResponse
// @Router /api/auth/signup [post]
func (s *gatewayServer) handleAuthSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req pb.SignUpRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := contextFromHTTPRequest(r)
	resp, err := s.apiServer.SignUp(ctx, &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// handleAuthSignIn godoc
// @Summary Sign in user
// @Description Authenticate user with username and password, returns JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body pb.SignInRequest true "SignIn request body"
// @Success 200 {object} pb.SignInResponse
// @Failure 400 {object} pb.ErrorBodyResponse
// @Failure 401 {object} pb.ErrorBodyResponse "Invalid credentials"
// @Failure 500 {object} pb.ErrorBodyResponse
// @Router /api/auth/signin [post]
func (s *gatewayServer) handleAuthSignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req pb.SignInRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := contextFromHTTPRequest(r)
	resp, err := s.apiServer.SignIn(ctx, &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleAuthGetMe godoc
// @Summary Get current user
// @Description Retrieve current authenticated user information from JWT token
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} pb.GetMeResponse
// @Failure 401 {object} pb.ErrorBodyResponse
// @Failure 500 {object} pb.ErrorBodyResponse
// @Router /api/auth/me [get]
func (s *gatewayServer) handleAuthGetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "GET" {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := contextFromHTTPRequest(r)
	resp, err := s.apiServer.GetMe(ctx, &pb.GetMeRequest{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ── Profile endpoints ──

func (s *gatewayServer) handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}

	ctx := contextFromHTTPRequest(r)

	switch r.Method {
	case "GET":
		// GET /api/profile → GetMyProfile (JWT required)
		resp, err := s.apiServer.GetMyProfile(ctx, &pb.GetMyProfileRequest{})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)

	case "POST":
		// POST /api/profile → CreateProfile
		var req pb.CreateProfileRequest
		if err := decodeJSONBody(r, &req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		resp, err := s.apiServer.CreateProfile(ctx, &req)
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, resp)

	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *gatewayServer) handleProfileByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}

	// Extract ID from path: /api/profile/{id}
	id := strings.TrimPrefix(r.URL.Path, "/api/profile/")
	if id == "" || strings.Contains(id, "/") {
		writeJSONError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	ctx := contextFromHTTPRequest(r)

	switch r.Method {
	case "GET":
		// GET /api/profile/{id} → GetProfileByID
		resp, err := s.apiServer.GetProfileByID(ctx, &pb.GetProfileByIDRequest{Id: id})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)

	case "PUT":
		// PUT /api/profile/{id} → UpdateProfile
		var req pb.UpdateProfileRequest
		if err := decodeJSONBody(r, &req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		req.Id = id
		resp, err := s.apiServer.UpdateProfile(ctx, &req)
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)

	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProfileByUserID godoc
// @Summary Get profile by user ID
// @Description Retrieve a user profile by the user's UUID
// @Tags Profile
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {object} pb.ProfileResponse
// @Failure 400 {object} pb.ErrorBodyResponse
// @Failure 404 {object} pb.ErrorBodyResponse
// @Failure 500 {object} pb.ErrorBodyResponse
// @Router /api/profile/user/{user_id} [get]
func (s *gatewayServer) handleProfileByUserID(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "GET" {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract user_id from path: /api/profile/user/{user_id}
	userID := strings.TrimPrefix(r.URL.Path, "/api/profile/user/")
	if userID == "" {
		writeJSONError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	ctx := contextFromHTTPRequest(r)
	resp, err := s.apiServer.GetProfileByUserID(ctx, &pb.GetProfileByUserIDRequest{UserId: userID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ── Menu endpoints ──

// handleMenuAll godoc
// @Summary Get all menus
// @Description Retrieve all homepage menu items
// @Tags Menu
// @Produce json
// @Success 200 {object} pb.MenuListResponse
// @Failure 500 {object} pb.ErrorBodyResponse
// @Router /api/menu [get]
func (s *gatewayServer) handleMenuAll(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "GET" {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := contextFromHTTPRequest(r)
	resp, err := s.apiServer.GetAllMenus(ctx, &pb.GetAllMenusRequest{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleMenuByAccountType godoc
// @Summary Get menus by account type
// @Description Retrieve menus filtered by account type. PREMIUM gets all menus, REGULAR gets only REGULAR menus.
// @Tags Menu
// @Produce json
// @Param account_type path string true "Account type (REGULAR or PREMIUM)"
// @Success 200 {object} pb.MenuListResponse
// @Failure 400 {object} pb.ErrorBodyResponse
// @Failure 500 {object} pb.ErrorBodyResponse
// @Router /api/menu/{account_type} [get]
func (s *gatewayServer) handleMenuByAccountType(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "GET" {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract account_type from path: /api/menu/{account_type}
	accountType := strings.TrimPrefix(r.URL.Path, "/api/menu/")
	if accountType == "" {
		writeJSONError(w, http.StatusBadRequest, "account_type is required")
		return
	}

	ctx := contextFromHTTPRequest(r)
	resp, err := s.apiServer.GetMenusByAccountType(ctx, &pb.GetMenusByAccountTypeRequest{AccountType: accountType})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ── Helpers ──

func contextFromHTTPRequest(r *http.Request) context.Context {
	md := metadata.New(nil)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		md.Set("authorization", authHeader)
	}
	ctx := metadata.NewIncomingContext(r.Context(), md)

	// Inject user_claims into context so protected gRPC handlers can read them.
	// The gRPC auth interceptor only fires for gRPC transport connections, not
	// for direct handler calls made by this HTTP gateway.
	if jwtMgr != nil && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if claims, err := jwtMgr.Verify(token); err == nil {
			ctx = context.WithValue(ctx, "user_claims", claims)
		}
	}

	return ctx
}

func decodeJSONBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, _ := status.FromError(err)
	httpCode := grpcToHTTPCode(int(st.Code()))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(pb.ErrorBodyResponse{
		Error:   true,
		Code:    int32(httpCode),
		Message: st.Message(),
	})
}
