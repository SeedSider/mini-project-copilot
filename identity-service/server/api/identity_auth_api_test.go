package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "bitbucket.bri.co.id/scm/addons/addons-identity-service/protogen/identity-service"
	manager "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/jwt"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/database"
	databasemock "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/database/mock"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/logger"
)

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dbSql := &database.DbSql{
		SqlDb: db,
		Dbw:   &databasemock.DatabaseMock{DbPq: db},
	}
	dbSql.Conn = db

	testLogger := logger.New(&logger.LoggerConfig{
		Env:         "DEV",
		ServiceName: "test",
		ProductName: "test",
		LogLevel:    "error",
		LogOutput:   "stdout",
	})

	srv := New("test-secret", "24h", dbSql, testLogger, nil, "")
	return srv, mock
}

func hashedPassword(plain string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	return string(h)
}

// ═══════════════════════════════════════════════════════════
// gRPC SignUp
// ═══════════════════════════════════════════════════════════

func TestSignUp_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WithArgs("johndoe").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("uuid-123"))

	resp, err := srv.SignUp(ctx, &pb.SignUpRequest{
		Username: "johndoe",
		Password: "password123",
		Phone:    "08123",
	})
	require.NoError(t, err)
	assert.Equal(t, "uuid-123", resp.UserId)
	assert.Equal(t, "johndoe", resp.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignUp_EmptyUsername(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.SignUp(context.Background(), &pb.SignUpRequest{
		Username: "",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "username is required")
}

func TestSignUp_ShortPassword(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.SignUp(context.Background(), &pb.SignUpRequest{
		Username: "johndoe",
		Password: "12345",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "at least 6 characters")
}

func TestSignUp_UsernameExists(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WithArgs("johndoe").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := srv.SignUp(context.Background(), &pb.SignUpRequest{
		Username: "johndoe",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

func TestSignUp_CheckUsernameDBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.SignUp(context.Background(), &pb.SignUpRequest{
		Username: "johndoe",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestSignUp_CreateUserDBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WithArgs("johndoe").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(fmt.Errorf("insert error"))

	_, err := srv.SignUp(context.Background(), &pb.SignUpRequest{
		Username: "johndoe",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════════════════════
// gRPC SignIn
// ═══════════════════════════════════════════════════════════

func TestSignIn_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	ctx := context.Background()
	pw := hashedPassword("password123")

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WithArgs("johndoe").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("uid-1", "johndoe", pw, "2026-03-27"))

	resp, err := srv.SignIn(ctx, &pb.SignInRequest{
		Username: "johndoe",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, "uid-1", resp.UserId)
	assert.Equal(t, "johndoe", resp.Username)
	assert.NotEmpty(t, resp.Token)
}

func TestSignIn_EmptyFields(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.SignIn(context.Background(), &pb.SignInRequest{Username: "", Password: ""})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestSignIn_UserNotFound(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	_, err := srv.SignIn(context.Background(), &pb.SignInRequest{
		Username: "unknown",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestSignIn_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnError(fmt.Errorf("connection error"))

	_, err := srv.SignIn(context.Background(), &pb.SignInRequest{
		Username: "johndoe",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestSignIn_WrongPassword(t *testing.T) {
	srv, mock := newTestServer(t)
	pw := hashedPassword("correct-password")

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WithArgs("johndoe").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("uid-1", "johndoe", pw, "2026-03-27"))

	_, err := srv.SignIn(context.Background(), &pb.SignInRequest{
		Username: "johndoe",
		Password: "wrong-password",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

// ═══════════════════════════════════════════════════════════
// gRPC GetMe
// ═══════════════════════════════════════════════════════════

func TestGetMe_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	claims := &manager.UserClaims{UserID: "uid-1", Username: "johndoe"}
	ctx := context.WithValue(context.Background(), "user_claims", claims)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WithArgs("johndoe").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("uid-1", "johndoe", "hash", "2026-03-27"))

	resp, err := srv.GetMe(ctx, &pb.GetMeRequest{})
	require.NoError(t, err)
	assert.Equal(t, "uid-1", resp.UserId)
	assert.Equal(t, "johndoe", resp.Username)
}

func TestGetMe_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetMe(context.Background(), &pb.GetMeRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGetMe_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	claims := &manager.UserClaims{UserID: "uid-1", Username: "johndoe"}
	ctx := context.WithValue(context.Background(), "user_claims", claims)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetMe(ctx, &pb.GetMeRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════════════════════
// HTTP Handlers
// ═══════════════════════════════════════════════════════════

func TestHandleSignUp_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("uuid-http"))

	body := `{"username":"httpuser","password":"password123","phone":"08123"}`
	req := httptest.NewRequest("POST", "/api/auth/signup", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.HandleSignUp(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp SignUpResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "uuid-http", resp.UserID)
}

func TestHandleSignUp_InvalidJSON(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest("POST", "/api/auth/signup", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	srv.HandleSignUp(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleSignIn_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	pw := hashedPassword("password123")

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("uid-http", "httpuser", pw, "2026-03-27"))

	body := `{"username":"httpuser","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/signin", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.HandleSignIn(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp SignInResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "uid-http", resp.UserID)
	assert.NotEmpty(t, resp.Token)
}

func TestHandleSignIn_InvalidJSON(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest("POST", "/api/auth/signin", strings.NewReader("bad"))
	w := httptest.NewRecorder()

	srv.HandleSignIn(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetMe_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	// Generate a valid token
	token, _ := srv.manager.Generate("uid-me", "meuser")

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WithArgs("meuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("uid-me", "meuser", "hash", "2026-03-27"))

	req := httptest.NewRequest("GET", "/api/identity/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	srv.HandleGetMe(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp GetMeResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "uid-me", resp.UserID)
}

func TestHandleGetMe_NoAuth(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/identity/me", nil)
	w := httptest.NewRecorder()

	srv.HandleGetMe(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGetMe_InvalidToken(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/identity/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	srv.HandleGetMe(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ═══════════════════════════════════════════════════════════
// Error helpers
// ═══════════════════════════════════════════════════════════

func TestErrorHelpers(t *testing.T) {
	srv, _ := newTestServer(t)

	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{"notImplemented", srv.notImplementedError(), codes.Unimplemented},
		{"serverError", srv.serverError(), codes.Internal},
		{"serverErrorWithDetail", srv.serverErrorWithDetail(fmt.Errorf("detail")), codes.Internal},
		{"badRequest", srv.badRequestError("bad"), codes.InvalidArgument},
		{"unauthorized", srv.unauthorizedError(), codes.Unauthenticated},
		{"conflict", srv.conflictError("exists"), codes.AlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, _ := status.FromError(tt.err)
			assert.Equal(t, tt.wantCode, st.Code())
		})
	}
}

// ═══════════════════════════════════════════════════════════
// grpcToHTTPCode + helpers
// ═══════════════════════════════════════════════════════════

func TestGrpcToHTTPCode(t *testing.T) {
	tests := []struct {
		grpcCode codes.Code
		httpCode int
	}{
		{codes.InvalidArgument, http.StatusUnprocessableEntity},
		{codes.Unauthenticated, http.StatusUnauthorized},
		{codes.AlreadyExists, http.StatusConflict},
		{codes.NotFound, http.StatusNotFound},
		{codes.PermissionDenied, http.StatusForbidden},
		{codes.Internal, http.StatusInternalServerError},
		{codes.Canceled, http.StatusInternalServerError}, // default
	}
	for _, tt := range tests {
		t.Run(tt.grpcCode.String(), func(t *testing.T) {
			assert.Equal(t, tt.httpCode, grpcToHTTPCode(tt.grpcCode))
		})
	}
}

func TestWriteGrpcErrorResponse_NonGrpcError(t *testing.T) {
	w := httptest.NewRecorder()
	writeGrpcErrorResponse(w, fmt.Errorf("plain error"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestValidateSignUpRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *SignUpRequest
		wantErr bool
		errMsg  string
	}{
		{"valid", &SignUpRequest{Username: "user", Password: "123456"}, false, ""},
		{"empty username", &SignUpRequest{Username: "", Password: "123456"}, true, "username is required"},
		{"short password", &SignUpRequest{Username: "user", Password: "12345"}, true, "at least 6 characters"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSignUpRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSONResponse(w, http.StatusOK, map[string]string{"key": "value"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestWriteErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	writeErrorResponse(w, http.StatusBadRequest, "bad request")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp errorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.True(t, resp.Error)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, "bad request", resp.Message)
}

func TestGetManager(t *testing.T) {
	srv, _ := newTestServer(t)
	m := srv.GetManager()
	assert.NotNil(t, m)
}

// ═══════════════════════════════════════════════════════════
// httpSignUp / httpSignIn / httpGetMe (internal methods)
// ═══════════════════════════════════════════════════════════

func TestHttpSignUp_ValidationError(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.httpSignUp(context.Background(), &SignUpRequest{Username: "", Password: "123456"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHttpSignIn_EmptyFields(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.httpSignIn(context.Background(), &SignInRequest{Username: "", Password: ""})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHandleSignIn_GrpcError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnError(sql.ErrNoRows)

	body := `{"username":"unknown","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/signin", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.HandleSignIn(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ═══════════════════════════════════════════════════════════
// httpSignUp - additional branches
// ═══════════════════════════════════════════════════════════

func TestHttpSignUp_CheckUsernameDBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.httpSignUp(context.Background(), &SignUpRequest{
		Username: "johndoe",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestHttpSignUp_UsernameExists(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := srv.httpSignUp(context.Background(), &SignUpRequest{
		Username: "johndoe",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

func TestHttpSignUp_CreateUserDBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(fmt.Errorf("insert error"))

	_, err := srv.httpSignUp(context.Background(), &SignUpRequest{
		Username: "johndoe",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestHttpSignUp_Success(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("uuid-http"))

	resp, err := srv.httpSignUp(context.Background(), &SignUpRequest{
		Username: "johndoe",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, "uuid-http", resp.UserID)
}

// ═══════════════════════════════════════════════════════════
// httpSignIn - additional branches
// ═══════════════════════════════════════════════════════════

func TestHttpSignIn_UserNotFound(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnError(sql.ErrNoRows)

	_, err := srv.httpSignIn(context.Background(), &SignInRequest{
		Username: "unknown",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestHttpSignIn_DBError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnError(fmt.Errorf("connection error"))

	_, err := srv.httpSignIn(context.Background(), &SignInRequest{
		Username: "johndoe",
		Password: "password123",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestHttpSignIn_WrongPassword(t *testing.T) {
	srv, mock := newTestServer(t)
	pw := hashedPassword("correct")

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("uid-1", "johndoe", pw, "2026-03-27"))

	_, err := srv.httpSignIn(context.Background(), &SignInRequest{
		Username: "johndoe",
		Password: "wrong",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestHttpSignIn_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	pw := hashedPassword("password123")

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("uid-1", "johndoe", pw, "2026-03-27"))

	resp, err := srv.httpSignIn(context.Background(), &SignInRequest{
		Username: "johndoe",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, "uid-1", resp.UserID)
	assert.NotEmpty(t, resp.Token)
}

// ═══════════════════════════════════════════════════════════
// httpGetMe - additional branches
// ═══════════════════════════════════════════════════════════

func TestHttpGetMe_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	claims := &manager.UserClaims{UserID: "uid-1", Username: "johndoe"}
	ctx := context.WithValue(context.Background(), "user_claims", claims)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("uid-1", "johndoe", "hash", "2026-03-27"))

	resp, err := srv.httpGetMe(ctx)
	require.NoError(t, err)
	assert.Equal(t, "uid-1", resp.UserID)
}

func TestHttpGetMe_NoClaims(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.httpGetMe(context.Background())
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestHttpGetMe_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	claims := &manager.UserClaims{UserID: "uid-1", Username: "johndoe"}
	ctx := context.WithValue(context.Background(), "user_claims", claims)

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.httpGetMe(ctx)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════════════════════
// HandleSignUp - GrpcError branch
// ═══════════════════════════════════════════════════════════

func TestHandleSignUp_GrpcError(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	body := `{"username":"johndoe","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/signup", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.HandleSignUp(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// ═══════════════════════════════════════════════════════════
// HandleGetMe - additional branches
// ═══════════════════════════════════════════════════════════

func TestHandleGetMe_GetMeServiceError(t *testing.T) {
	srv, mock := newTestServer(t)

	token, _ := srv.manager.Generate("uid-me", "meuser")

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users`).
		WillReturnError(fmt.Errorf("db error"))

	req := httptest.NewRequest("GET", "/api/identity/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	srv.HandleGetMe(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
