package api

import (
"context"
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"strings"
"testing"

"github.com/DATA-DOG/go-sqlmock"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"

pb "bitbucket.bri.co.id/scm/addons/addons-identity-service/protogen/identity-service"
)

// ===================================================================
// gRPC ValidateOtp
// ===================================================================

func TestValidateOtp_Success(t *testing.T) {
srv, mock := newTestServer(t)
ctx := context.Background()

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WithArgs("johndoe").
WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

resp, err := srv.ValidateOtp(ctx, &pb.ValidateOtpRequest{Username: "johndoe"})
require.NoError(t, err)
assert.GreaterOrEqual(t, resp.Otp, int32(100000))
assert.LessOrEqual(t, resp.Otp, int32(999999))
assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidateOtp_EmptyUsername(t *testing.T) {
srv, _ := newTestServer(t)
_, err := srv.ValidateOtp(context.Background(), &pb.ValidateOtpRequest{Username: ""})
st, _ := status.FromError(err)
assert.Equal(t, codes.InvalidArgument, st.Code())
assert.Contains(t, st.Message(), "username is required")
}

func TestValidateOtp_UsernameNotFound(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WithArgs("unknown").
WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

_, err := srv.ValidateOtp(context.Background(), &pb.ValidateOtpRequest{Username: "unknown"})
st, _ := status.FromError(err)
assert.Equal(t, codes.NotFound, st.Code())
}

func TestValidateOtp_DBError(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WillReturnError(fmt.Errorf("db error"))

_, err := srv.ValidateOtp(context.Background(), &pb.ValidateOtpRequest{Username: "johndoe"})
st, _ := status.FromError(err)
assert.Equal(t, codes.Internal, st.Code())
}

func TestValidateOtp_OTPGeneratorError(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WithArgs("johndoe").
WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

origGen := generateOTP
generateOTP = func() (int32, error) { return 0, fmt.Errorf("rng failure") }
defer func() { generateOTP = origGen }()

_, err := srv.ValidateOtp(context.Background(), &pb.ValidateOtpRequest{Username: "johndoe"})
st, _ := status.FromError(err)
assert.Equal(t, codes.Internal, st.Code())
}

func TestValidateOtp_RandomOTPRange(t *testing.T) {
srv, mock := newTestServer(t)

for i := 0; i < 10; i++ {
mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WithArgs("johndoe").
WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

resp, err := srv.ValidateOtp(context.Background(), &pb.ValidateOtpRequest{Username: "johndoe"})
require.NoError(t, err)
assert.GreaterOrEqual(t, resp.Otp, int32(100000), "OTP must be at least 100000")
assert.LessOrEqual(t, resp.Otp, int32(999999), "OTP must be at most 999999")
}
}

// ===================================================================
// gRPC UpdatePassword
// ===================================================================

func TestUpdatePassword_Success(t *testing.T) {
srv, mock := newTestServer(t)
ctx := context.Background()

mock.ExpectExec(`UPDATE users SET password_hash`).
WithArgs(sqlmock.AnyArg(), "johndoe").
WillReturnResult(sqlmock.NewResult(0, 1))

resp, err := srv.UpdatePassword(ctx, &pb.UpdatePasswordRequest{
Username:    "johndoe",
NewPassword: "newPassword123",
})
require.NoError(t, err)
assert.Equal(t, "berhasil ubah password", resp.Message)
assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePassword_EmptyUsername(t *testing.T) {
srv, _ := newTestServer(t)
_, err := srv.UpdatePassword(context.Background(), &pb.UpdatePasswordRequest{
Username:    "",
NewPassword: "newPassword123",
})
st, _ := status.FromError(err)
assert.Equal(t, codes.InvalidArgument, st.Code())
assert.Contains(t, st.Message(), "username is required")
}

func TestUpdatePassword_ShortPassword(t *testing.T) {
srv, _ := newTestServer(t)
_, err := srv.UpdatePassword(context.Background(), &pb.UpdatePasswordRequest{
Username:    "johndoe",
NewPassword: "12345",
})
st, _ := status.FromError(err)
assert.Equal(t, codes.InvalidArgument, st.Code())
assert.Contains(t, st.Message(), "at least 6 characters")
}

func TestUpdatePassword_DBError(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectExec(`UPDATE users SET password_hash`).
WillReturnError(fmt.Errorf("db error"))

_, err := srv.UpdatePassword(context.Background(), &pb.UpdatePasswordRequest{
Username:    "johndoe",
NewPassword: "newPassword123",
})
st, _ := status.FromError(err)
assert.Equal(t, codes.Internal, st.Code())
}

func TestUpdatePassword_UserNotFound(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectExec(`UPDATE users SET password_hash`).
WithArgs(sqlmock.AnyArg(), "unknown").
WillReturnResult(sqlmock.NewResult(0, 0))

_, err := srv.UpdatePassword(context.Background(), &pb.UpdatePasswordRequest{
Username:    "unknown",
NewPassword: "newPassword123",
})
st, _ := status.FromError(err)
assert.Equal(t, codes.Internal, st.Code())
}

// ===================================================================
// HTTP HandleValidateOtp
// ===================================================================

func TestHandleValidateOtp_Success(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

body := `{"username":"johndoe"}`
req := httptest.NewRequest("POST", "/api/auth/validate-otp", strings.NewReader(body))
w := httptest.NewRecorder()

srv.HandleValidateOtp(w, req)
assert.Equal(t, http.StatusOK, w.Code)

var resp ValidateOtpResponse
json.NewDecoder(w.Body).Decode(&resp)
assert.GreaterOrEqual(t, resp.OTP, int32(100000))
assert.LessOrEqual(t, resp.OTP, int32(999999))
}

func TestHandleValidateOtp_InvalidJSON(t *testing.T) {
srv, _ := newTestServer(t)

req := httptest.NewRequest("POST", "/api/auth/validate-otp", strings.NewReader("not-json"))
w := httptest.NewRecorder()

srv.HandleValidateOtp(w, req)
assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleValidateOtp_GrpcError(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

body := `{"username":"unknown"}`
req := httptest.NewRequest("POST", "/api/auth/validate-otp", strings.NewReader(body))
w := httptest.NewRecorder()

srv.HandleValidateOtp(w, req)
assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===================================================================
// HTTP HandleUpdatePassword (JWT required)
// ===================================================================

func TestHandleUpdatePassword_Success(t *testing.T) {
srv, mock := newTestServer(t)

token, err := srv.manager.Generate("uuid-123", "johndoe")
require.NoError(t, err)

mock.ExpectExec(`UPDATE users SET password_hash`).
WithArgs(sqlmock.AnyArg(), "johndoe").
WillReturnResult(sqlmock.NewResult(0, 1))

body := `{"newPassword":"newPassword123"}`
req := httptest.NewRequest("PUT", "/api/auth/update-password", strings.NewReader(body))
req.Header.Set("Authorization", "Bearer "+token)
w := httptest.NewRecorder()

srv.HandleUpdatePassword(w, req)
assert.Equal(t, http.StatusOK, w.Code)

var resp UpdatePasswordResponse
json.NewDecoder(w.Body).Decode(&resp)
assert.Equal(t, "berhasil ubah password", resp.Message)
}

func TestHandleUpdatePassword_NoAuthHeader(t *testing.T) {
srv, _ := newTestServer(t)

body := `{"newPassword":"newPassword123"}`
req := httptest.NewRequest("PUT", "/api/auth/update-password", strings.NewReader(body))
w := httptest.NewRecorder()

srv.HandleUpdatePassword(w, req)
assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleUpdatePassword_InvalidToken(t *testing.T) {
srv, _ := newTestServer(t)

body := `{"newPassword":"newPassword123"}`
req := httptest.NewRequest("PUT", "/api/auth/update-password", strings.NewReader(body))
req.Header.Set("Authorization", "Bearer invalid-token")
w := httptest.NewRecorder()

srv.HandleUpdatePassword(w, req)
assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleUpdatePassword_InvalidJSON(t *testing.T) {
srv, _ := newTestServer(t)

token, err := srv.manager.Generate("uuid-123", "johndoe")
require.NoError(t, err)

req := httptest.NewRequest("PUT", "/api/auth/update-password", strings.NewReader("not-json"))
req.Header.Set("Authorization", "Bearer "+token)
w := httptest.NewRecorder()

srv.HandleUpdatePassword(w, req)
assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUpdatePassword_ShortPassword(t *testing.T) {
srv, _ := newTestServer(t)

token, err := srv.manager.Generate("uuid-123", "johndoe")
require.NoError(t, err)

body := `{"newPassword":"123"}`
req := httptest.NewRequest("PUT", "/api/auth/update-password", strings.NewReader(body))
req.Header.Set("Authorization", "Bearer "+token)
w := httptest.NewRecorder()

srv.HandleUpdatePassword(w, req)
assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandleUpdatePassword_DBError(t *testing.T) {
srv, mock := newTestServer(t)

token, err := srv.manager.Generate("uuid-123", "johndoe")
require.NoError(t, err)

mock.ExpectExec(`UPDATE users SET password_hash`).
WillReturnError(fmt.Errorf("db error"))

body := `{"newPassword":"newPassword123"}`
req := httptest.NewRequest("PUT", "/api/auth/update-password", strings.NewReader(body))
req.Header.Set("Authorization", "Bearer "+token)
w := httptest.NewRecorder()

srv.HandleUpdatePassword(w, req)
assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===================================================================
// httpValidateOtp internal method
// ===================================================================

func TestHttpValidateOtp_EmptyUsername(t *testing.T) {
srv, _ := newTestServer(t)
_, err := srv.httpValidateOtp(context.Background(), &ValidateOtpRequest{Username: ""})
st, _ := status.FromError(err)
assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHttpValidateOtp_DBError(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WillReturnError(fmt.Errorf("db error"))

_, err := srv.httpValidateOtp(context.Background(), &ValidateOtpRequest{Username: "johndoe"})
st, _ := status.FromError(err)
assert.Equal(t, codes.Internal, st.Code())
}

func TestHttpValidateOtp_UsernameNotFound(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

_, err := srv.httpValidateOtp(context.Background(), &ValidateOtpRequest{Username: "unknown"})
st, _ := status.FromError(err)
assert.Equal(t, codes.NotFound, st.Code())
}

func TestHttpValidateOtp_GeneratorError(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectQuery(`SELECT COUNT\(1\) FROM users`).
WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

origGen := generateOTP
generateOTP = func() (int32, error) { return 0, fmt.Errorf("rng failure") }
defer func() { generateOTP = origGen }()

_, err := srv.httpValidateOtp(context.Background(), &ValidateOtpRequest{Username: "johndoe"})
st, _ := status.FromError(err)
assert.Equal(t, codes.Internal, st.Code())
}

// ===================================================================
// httpUpdatePassword internal method
// ===================================================================

func TestHttpUpdatePassword_EmptyUsername(t *testing.T) {
srv, _ := newTestServer(t)
_, err := srv.httpUpdatePassword(context.Background(), "", &UpdatePasswordRequest{NewPassword: "password123"})
st, _ := status.FromError(err)
assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHttpUpdatePassword_ShortPassword(t *testing.T) {
srv, _ := newTestServer(t)
_, err := srv.httpUpdatePassword(context.Background(), "johndoe", &UpdatePasswordRequest{NewPassword: "123"})
st, _ := status.FromError(err)
assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHttpUpdatePassword_DBError(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectExec(`UPDATE users SET password_hash`).
WillReturnError(fmt.Errorf("db error"))

_, err := srv.httpUpdatePassword(context.Background(), "johndoe", &UpdatePasswordRequest{
NewPassword: "newPassword123",
})
st, _ := status.FromError(err)
assert.Equal(t, codes.Internal, st.Code())
}

func TestHttpUpdatePassword_Success(t *testing.T) {
srv, mock := newTestServer(t)

mock.ExpectExec(`UPDATE users SET password_hash`).
WillReturnResult(sqlmock.NewResult(0, 1))

resp, err := srv.httpUpdatePassword(context.Background(), "johndoe", &UpdatePasswordRequest{
NewPassword: "newPassword123",
})
require.NoError(t, err)
assert.Equal(t, "berhasil ubah password", resp.Message)
}

// ===================================================================
// generateOTP
// ===================================================================

func TestGenerateOTP_Range(t *testing.T) {
for i := 0; i < 20; i++ {
otp, err := generateOTP()
require.NoError(t, err)
assert.GreaterOrEqual(t, otp, int32(100000))
assert.LessOrEqual(t, otp, int32(999999))
}
}
