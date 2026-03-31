package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"github.com/bankease/user-profile-service/server/db"
)

const testJWTSecret = "test-secret-key-for-unit-tests"

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })
	provider := db.New(sqlDB)
	return New(provider, testJWTSecret, "", ""), mock
}

// makeTestJWT generates a valid HS256 JWT with the given user_id claim.
func makeTestJWT(t *testing.T, userID string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return tokenStr
}

// withURLParam injects chi URL params into the request context.
func withURLParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

var testProfileCols = []string{
	"id", "user_id", "bank", "branch", "name",
	"card_number", "card_provider", "balance", "currency", "account_type", "image",
}

func testProfileRow(id, userID string) *sqlmock.Rows {
	uid := userID
	return sqlmock.NewRows(testProfileCols).AddRow(
		id, &uid, "BRI", "Jakarta", "John Doe",
		"4111111111111111", "VISA", int64(1_000_000), "IDR", "REGULAR", "",
	)
}

// ═══════════════════════════════════════════
// HTTP HandleGetMyProfile
// ═══════════════════════════════════════════

func TestHandleGetMyProfile_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	userID := "user-uuid-1"
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(userID).
		WillReturnRows(testProfileRow("profile-uuid-1", userID))

	token := makeTestJWT(t, userID)
	r := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetMyProfile_MissingAuthHeader(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGetMyProfile_InvalidToken(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	r.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGetMyProfile_ProfileNotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	userID := "user-uuid-missing"
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	token := makeTestJWT(t, userID)
	r := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetMyProfile_InternalError(t *testing.T) {
	srv, mock := newTestServer(t)
	userID := "user-uuid-1"
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("connection refused"))

	token := makeTestJWT(t, userID)
	r := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetMyProfile_MissingUserIDClaim(t *testing.T) {
	srv, _ := newTestServer(t)
	// JWT with no user_id claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	r.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGetMyProfile_EmptyUserIDClaim(t *testing.T) {
	srv, _ := newTestServer(t)
	// JWT with user_id = "" (empty string)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	r.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetProfile
// ═══════════════════════════════════════════

func TestHandleGetProfile_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("profile-uuid-1").
		WillReturnRows(testProfileRow("profile-uuid-1", "user-uuid-1"))

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "id", "profile-uuid-1")
	w := httptest.NewRecorder()

	srv.HandleGetProfile(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetProfile_NotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "id", "nonexistent")
	w := httptest.NewRecorder()

	srv.HandleGetProfile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetProfile_MissingID(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/api/profile/", nil)
	w := httptest.NewRecorder()

	srv.HandleGetProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetProfile_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("any-uuid").
		WillReturnError(fmt.Errorf("connection refused"))

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "id", "any-uuid")
	w := httptest.NewRecorder()

	srv.HandleGetProfile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleUpdateProfile
// ═══════════════════════════════════════════

func TestHandleUpdateProfile_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(db.EditProfileRequest{Bank: "BRI", Branch: "Jakarta", Name: "Jane", CardNumber: "1234"})
	r := withURLParam(httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)), "id", "profile-uuid-1")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUpdateProfile_NotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	body, _ := json.Marshal(db.EditProfileRequest{Bank: "BRI", Branch: "Jakarta", Name: "Jane", CardNumber: "1234"})
	r := withURLParam(httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)), "id", "nonexistent")
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleUpdateProfile_InvalidBody(t *testing.T) {
	srv, _ := newTestServer(t)
	r := withURLParam(httptest.NewRequest(http.MethodPut, "/", strings.NewReader("not json")), "id", "profile-uuid-1")
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUpdateProfile_MissingID(t *testing.T) {
	srv, _ := newTestServer(t)
	body, _ := json.Marshal(db.EditProfileRequest{})
	r := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUpdateProfile_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnError(fmt.Errorf("db error"))

	body, _ := json.Marshal(db.EditProfileRequest{Bank: "BRI", Branch: "Jakarta", Name: "Test", CardNumber: "1234"})
	r := withURLParam(httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)), "id", "profile-uuid-1")
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetProfileByUserID
// ═══════════════════════════════════════════

func TestHandleGetProfileByUserID_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("user-uuid-1").
		WillReturnRows(testProfileRow("profile-uuid-1", "user-uuid-1"))

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "user_id", "user-uuid-1")
	w := httptest.NewRecorder()

	srv.HandleGetProfileByUserID(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetProfileByUserID_MissingUserID(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.HandleGetProfileByUserID(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetProfileByUserID_NotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("user-nonexistent").
		WillReturnError(sql.ErrNoRows)

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "user_id", "user-nonexistent")
	w := httptest.NewRecorder()

	srv.HandleGetProfileByUserID(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetProfileByUserID_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("user-uuid-1").
		WillReturnError(fmt.Errorf("db error"))

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "user_id", "user-uuid-1")
	w := httptest.NewRecorder()

	srv.HandleGetProfileByUserID(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleCreateProfile
// ═══════════════════════════════════════════

func TestHandleCreateProfile_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	uid := "user-uuid-1"
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnRows(sqlmock.NewRows(testProfileCols).AddRow(
			"new-uuid", &uid, "BRI", "Jakarta", "John Doe",
			"1234", "VISA", int64(0), "IDR", "REGULAR", "",
		))

	body, _ := json.Marshal(db.CreateProfileRequest{
		UserID: "user-uuid-1", Bank: "BRI", Branch: "Jakarta", Name: "John Doe",
		CardNumber: "1234", CardProvider: "VISA", Currency: "IDR", AccountType: "REGULAR",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/profile", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.HandleCreateProfile(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCreateProfile_MissingUserID(t *testing.T) {
	srv, _ := newTestServer(t)
	body, _ := json.Marshal(db.CreateProfileRequest{Bank: "BRI"})
	r := httptest.NewRequest(http.MethodPost, "/api/profile", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.HandleCreateProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateProfile_InvalidBody(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodPost, "/api/profile", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	srv.HandleCreateProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateProfile_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnError(fmt.Errorf("db error"))

	body, _ := json.Marshal(db.CreateProfileRequest{UserID: "user-uuid-1", Bank: "BRI"})
	r := httptest.NewRequest(http.MethodPost, "/api/profile", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.HandleCreateProfile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// gRPC: CreateProfile
// ═══════════════════════════════════════════

func TestCreateProfileGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	uid := "user-uuid-1"
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnRows(sqlmock.NewRows(testProfileCols).AddRow(
			"new-uuid", &uid, "BRI", "Jakarta", "John Doe",
			"1234", "VISA", int64(0), "IDR", "REGULAR", "",
		))

	resp, err := srv.CreateProfile(context.Background(), &pb.CreateProfileRequest{
		UserId: "user-uuid-1", Bank: "BRI", Branch: "Jakarta", Name: "John Doe",
		CardNumber: "1234", CardProvider: "VISA", Currency: "IDR", AccountType: "REGULAR",
	})
	require.NoError(t, err)
	assert.Equal(t, "new-uuid", resp.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProfileGRPC_EmptyUserID(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.CreateProfile(context.Background(), &pb.CreateProfileRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateProfileGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.CreateProfile(context.Background(), &pb.CreateProfileRequest{UserId: "user-uuid-1"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetProfileByID
// ═══════════════════════════════════════════

func TestGetProfileByIDGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("profile-uuid-1").
		WillReturnRows(testProfileRow("profile-uuid-1", "user-uuid-1"))

	resp, err := srv.GetProfileByID(context.Background(), &pb.GetProfileByIDRequest{Id: "profile-uuid-1"})
	require.NoError(t, err)
	assert.Equal(t, "profile-uuid-1", resp.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileByIDGRPC_EmptyID(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetProfileByID(context.Background(), &pb.GetProfileByIDRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetProfileByIDGRPC_NotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err := srv.GetProfileByID(context.Background(), &pb.GetProfileByIDRequest{Id: "nonexistent"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGetProfileByIDGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("any-uuid").
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetProfileByID(context.Background(), &pb.GetProfileByIDRequest{Id: "any-uuid"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetProfileByUserID
// ═══════════════════════════════════════════

func TestGetProfileByUserIDGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("user-uuid-1").
		WillReturnRows(testProfileRow("profile-uuid-1", "user-uuid-1"))

	resp, err := srv.GetProfileByUserID(context.Background(), &pb.GetProfileByUserIDRequest{UserId: "user-uuid-1"})
	require.NoError(t, err)
	assert.Equal(t, "profile-uuid-1", resp.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileByUserIDGRPC_EmptyUserID(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetProfileByUserID(context.Background(), &pb.GetProfileByUserIDRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetProfileByUserIDGRPC_NotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err := srv.GetProfileByUserID(context.Background(), &pb.GetProfileByUserIDRequest{UserId: "nonexistent"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGetProfileByUserIDGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("any-uuid").
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetProfileByUserID(context.Background(), &pb.GetProfileByUserIDRequest{UserId: "any-uuid"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: UpdateProfile
// ═══════════════════════════════════════════

func TestUpdateProfileGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
		Id: "profile-uuid-1", Bank: "BRI", Branch: "Jakarta", Name: "Jane Doe", CardNumber: "1234",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(http.StatusOK), resp.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProfileGRPC_EmptyID(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestUpdateProfileGRPC_NotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
		Id: "nonexistent", Bank: "BRI", Branch: "Jakarta", Name: "Test", CardNumber: "1234",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestUpdateProfileGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
		Id: "profile-uuid-1", Bank: "BRI", Branch: "Jakarta", Name: "Test", CardNumber: "1234",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}
