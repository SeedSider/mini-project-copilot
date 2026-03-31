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

const (
	testJWTSecret    		= "test-secret-key-for-unit-tests"
	testProfileUUID  		= "profile-uuid-1"
	testUserUUID     		= "user-uuid-1"
	testProfileName  		= "John Doe"
	testAPIProfile   		= "/api/profile"
	testBearerPrefix 		= "Bearer "
	testAnyUUID      		= "any-uuid"
	testHeaderContentType	= "Content-Type"
	testApplicationJSON		= "application/json"
	testDbError				= "db error"
	testNewUUID				= "new-uuid"
)

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
		id, &uid, "BRI", "Jakarta", testProfileName,
		"4111111111111111", "VISA", int64(1_000_000), "IDR", "REGULAR", "",
	)
}

// ═══════════════════════════════════════════
// HTTP HandleGetMyProfile
// ═══════════════════════════════════════════

func TestHandleGetMyProfileSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	userID := testUserUUID
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(userID).
		WillReturnRows(testProfileRow(testProfileUUID, userID))

	token := makeTestJWT(t, userID)
	r := httptest.NewRequest(http.MethodGet, testAPIProfile, nil)
	r.Header.Set("Authorization", testBearerPrefix+token)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetMyProfileMissingAuthHeader(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, testAPIProfile, nil)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGetMyProfileInvalidToken(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, testAPIProfile, nil)
	r.Header.Set("Authorization", testBearerPrefix+"invalid.token.here")
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGetMyProfileProfileNotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	userID := "user-uuid-missing"
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	token := makeTestJWT(t, userID)
	r := httptest.NewRequest(http.MethodGet, testAPIProfile, nil)
	r.Header.Set("Authorization", testBearerPrefix+token)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetMyProfileInternalError(t *testing.T) {
	srv, mock := newTestServer(t)
	userID := testUserUUID
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("connection refused"))

	token := makeTestJWT(t, userID)
	r := httptest.NewRequest(http.MethodGet, testAPIProfile, nil)
	r.Header.Set("Authorization", testBearerPrefix+token)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetMyProfileMissingUserIDClaim(t *testing.T) {
	srv, _ := newTestServer(t)
	// JWT with no user_id claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, testAPIProfile, nil)
	r.Header.Set("Authorization", testBearerPrefix+tokenStr)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGetMyProfileEmptyUserIDClaim(t *testing.T) {
	srv, _ := newTestServer(t)
	// JWT with user_id = "" (empty string)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, testAPIProfile, nil)
	r.Header.Set("Authorization", testBearerPrefix+tokenStr)
	w := httptest.NewRecorder()

	srv.HandleGetMyProfile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetProfile
// ═══════════════════════════════════════════

func TestHandleGetProfileSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testProfileUUID).
		WillReturnRows(testProfileRow(testProfileUUID, testUserUUID))

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "id", testProfileUUID)
	w := httptest.NewRecorder()

	srv.HandleGetProfile(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetProfileNotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "id", "nonexistent")
	w := httptest.NewRecorder()

	srv.HandleGetProfile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetProfileMissingID(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, testAPIProfile+"/", nil)
	w := httptest.NewRecorder()

	srv.HandleGetProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetProfileDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testAnyUUID).
		WillReturnError(fmt.Errorf("connection refused"))

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "id", testAnyUUID)
	w := httptest.NewRecorder()

	srv.HandleGetProfile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleUpdateProfile
// ═══════════════════════════════════════════

func TestHandleUpdateProfileSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(db.EditProfileRequest{Bank: "BRI", Branch: "Jakarta", Name: "Jane", CardNumber: "1234"})
	r := withURLParam(httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)), "id", testProfileUUID)
	r.Header.Set(testHeaderContentType, testApplicationJSON)
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUpdateProfileNotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	body, _ := json.Marshal(db.EditProfileRequest{Bank: "BRI", Branch: "Jakarta", Name: "Jane", CardNumber: "1234"})
	r := withURLParam(httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)), "id", "nonexistent")
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleUpdateProfileInvalidBody(t *testing.T) {
	srv, _ := newTestServer(t)
	r := withURLParam(httptest.NewRequest(http.MethodPut, "/", strings.NewReader("not json")), "id", testProfileUUID)
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUpdateProfileMissingID(t *testing.T) {
	srv, _ := newTestServer(t)
	body, _ := json.Marshal(db.EditProfileRequest{})
	r := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUpdateProfileDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnError(fmt.Errorf(testDbError))

	body, _ := json.Marshal(db.EditProfileRequest{Bank: "BRI", Branch: "Jakarta", Name: "Test", CardNumber: "1234"})
	r := withURLParam(httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)), "id", testProfileUUID)
	w := httptest.NewRecorder()

	srv.HandleUpdateProfile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetProfileByUserID
// ═══════════════════════════════════════════

func TestHandleGetProfileByUserIDSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testUserUUID).
		WillReturnRows(testProfileRow(testProfileUUID, testUserUUID))

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "user_id", testUserUUID)
	w := httptest.NewRecorder()

	srv.HandleGetProfileByUserID(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetProfileByUserIDMissingUserID(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.HandleGetProfileByUserID(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetProfileByUserIDNotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("user-nonexistent").
		WillReturnError(sql.ErrNoRows)

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "user_id", "user-nonexistent")
	w := httptest.NewRecorder()

	srv.HandleGetProfileByUserID(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetProfileByUserIDDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testUserUUID).
		WillReturnError(fmt.Errorf(testDbError))

	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "user_id", testUserUUID)
	w := httptest.NewRecorder()

	srv.HandleGetProfileByUserID(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleCreateProfile
// ═══════════════════════════════════════════

func TestHandleCreateProfileSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	uid := testUserUUID
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnRows(sqlmock.NewRows(testProfileCols).AddRow(
			testNewUUID, &uid, "BRI", "Jakarta", testProfileName,
			"1234", "VISA", int64(0), "IDR", "REGULAR", "",
		))

	body, _ := json.Marshal(db.CreateProfileRequest{
		UserID: testUserUUID, Bank: "BRI", Branch: "Jakarta", Name: testProfileName,
		CardNumber: "1234", CardProvider: "VISA", Currency: "IDR", AccountType: "REGULAR",
	})
	r := httptest.NewRequest(http.MethodPost, testAPIProfile, bytes.NewReader(body))
	r.Header.Set(testHeaderContentType, testApplicationJSON)
	w := httptest.NewRecorder()

	srv.HandleCreateProfile(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCreateProfileMissingUserID(t *testing.T) {
	srv, _ := newTestServer(t)
	body, _ := json.Marshal(db.CreateProfileRequest{Bank: "BRI"})
	r := httptest.NewRequest(http.MethodPost, testAPIProfile, bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.HandleCreateProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateProfileInvalidBody(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodPost, testAPIProfile, strings.NewReader("not json"))
	w := httptest.NewRecorder()

	srv.HandleCreateProfile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateProfileDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnError(fmt.Errorf(testDbError))

	body, _ := json.Marshal(db.CreateProfileRequest{UserID: testUserUUID, Bank: "BRI"})
	r := httptest.NewRequest(http.MethodPost, testAPIProfile, bytes.NewReader(body))
	r.Header.Set(testHeaderContentType, testApplicationJSON)
	w := httptest.NewRecorder()

	srv.HandleCreateProfile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// gRPC: CreateProfile
// ═══════════════════════════════════════════

func TestCreateProfileGRPCSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	uid := testUserUUID
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnRows(sqlmock.NewRows(testProfileCols).AddRow(
			testNewUUID, &uid, "BRI", "Jakarta", testProfileName,
			"1234", "VISA", int64(0), "IDR", "REGULAR", "",
		))

	resp, err := srv.CreateProfile(context.Background(), &pb.CreateProfileRequest{
		UserId: testUserUUID, Bank: "BRI", Branch: "Jakarta", Name: testProfileName,
		CardNumber: "1234", CardProvider: "VISA", Currency: "IDR", AccountType: "REGULAR",
	})
	require.NoError(t, err)
	assert.Equal(t, testNewUUID, resp.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProfileGRPCEmptyUserID(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.CreateProfile(context.Background(), &pb.CreateProfileRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateProfileGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnError(fmt.Errorf(testDbError))

	_, err := srv.CreateProfile(context.Background(), &pb.CreateProfileRequest{UserId: testUserUUID})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetProfileByID
// ═══════════════════════════════════════════

func TestGetProfileByIDGRPCSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testProfileUUID).
		WillReturnRows(testProfileRow(testProfileUUID, testUserUUID))

	resp, err := srv.GetProfileByID(context.Background(), &pb.GetProfileByIDRequest{Id: testProfileUUID})
	require.NoError(t, err)
	assert.Equal(t, testProfileUUID, resp.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileByIDGRPCEmptyID(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetProfileByID(context.Background(), &pb.GetProfileByIDRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetProfileByIDGRPCNotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err := srv.GetProfileByID(context.Background(), &pb.GetProfileByIDRequest{Id: "nonexistent"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGetProfileByIDGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testAnyUUID).
		WillReturnError(fmt.Errorf(testDbError))

	_, err := srv.GetProfileByID(context.Background(), &pb.GetProfileByIDRequest{Id: testAnyUUID})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetProfileByUserID
// ═══════════════════════════════════════════

func TestGetProfileByUserIDGRPCSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testUserUUID).
		WillReturnRows(testProfileRow(testProfileUUID, testUserUUID))

	resp, err := srv.GetProfileByUserID(context.Background(), &pb.GetProfileByUserIDRequest{UserId: testUserUUID})
	require.NoError(t, err)
	assert.Equal(t, testProfileUUID, resp.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileByUserIDGRPCEmptyUserID(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetProfileByUserID(context.Background(), &pb.GetProfileByUserIDRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetProfileByUserIDGRPCNotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err := srv.GetProfileByUserID(context.Background(), &pb.GetProfileByUserIDRequest{UserId: "nonexistent"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGetProfileByUserIDGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testAnyUUID).
		WillReturnError(fmt.Errorf(testDbError))

	_, err := srv.GetProfileByUserID(context.Background(), &pb.GetProfileByUserIDRequest{UserId: testAnyUUID})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: UpdateProfile
// ═══════════════════════════════════════════

func TestUpdateProfileGRPCSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
		Id: testProfileUUID, Bank: "BRI", Branch: "Jakarta", Name: "Jane Doe", CardNumber: "1234",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(http.StatusOK), resp.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProfileGRPCEmptyID(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestUpdateProfileGRPCNotFound(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
		Id: "nonexistent", Bank: "BRI", Branch: "Jakarta", Name: "Test", CardNumber: "1234",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestUpdateProfileGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectExec(`UPDATE profile`).
		WillReturnError(fmt.Errorf(testDbError))

	_, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
		Id: testProfileUUID, Bank: "BRI", Branch: "Jakarta", Name: "Test", CardNumber: "1234",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}
