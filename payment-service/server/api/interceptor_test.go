package api

import (
	"context"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bankease/payment-service/server/constant"
	"github.com/bankease/payment-service/server/db"
	manager "github.com/bankease/payment-service/server/jwt"
	"github.com/bankease/payment-service/server/lib/logger"
)

const testSecret = "test-secret"

func initTestLogger() {
	if log == nil {
		log = logger.New(&logger.LoggerConfig{
			Env:         "DEV",
			ServiceName: "test",
			ProductName: "test",
			LogLevel:    "error",
			LogOutput:   "stdout",
		})
	}
}

func generateTestToken(userID, username string) string {
	claims := &manager.UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		},
		UserID:   userID,
		Username: username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, _ := token.SignedString([]byte(testSecret))
	return str
}

// ═══════════════════════════════════════════════════════════
// NewAuthInterceptor + accessibleRoles
// ═══════════════════════════════════════════════════════════

func TestNewAuthInterceptor(t *testing.T) {
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	assert.NotNil(t, ai)
}

func TestAccessibleRoles(t *testing.T) {
	roles := accessibleRoles()
	assert.Contains(t, roles, apiServicePath+"GetInternetBill")
	assert.Contains(t, roles, apiServicePath+"GetBeneficiaries")
	assert.Contains(t, roles, apiServicePath+"PrepaidPay")
	// Public endpoints should NOT be in the restricted map
	assert.NotContains(t, roles, apiServicePath+"GetProviders")
	assert.NotContains(t, roles, apiServicePath+"GetCurrencyList")
}

// ═══════════════════════════════════════════════════════════
// AuthInterceptor.Unary
// ═══════════════════════════════════════════════════════════

func TestAuthInterceptor_Unary_Unrestricted(t *testing.T) {
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Unary()

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: apiServicePath + "GetProviders"}
	out, err := interceptor(context.Background(), nil, info, handler)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "ok", out)
}

func TestAuthInterceptor_Unary_Restricted_ValidToken(t *testing.T) {
	initTestLogger()
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Unary()

	token := generateTestToken("uid-1", "johndoe")
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		claims, ok := ctx.Value("user_claims").(*manager.UserClaims)
		assert.True(t, ok)
		assert.Equal(t, "uid-1", claims.UserID)
		called = true
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: apiServicePath + "GetInternetBill"}
	out, err := interceptor(ctx, nil, info, handler)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "ok", out)
}

func TestAuthInterceptor_Unary_Restricted_NoMetadata(t *testing.T) {
	initTestLogger()
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Unary()

	info := &grpc.UnaryServerInfo{FullMethod: apiServicePath + "GetInternetBill"}
	_, err := interceptor(context.Background(), nil, info, nil)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptor_Unary_Restricted_NoAuthHeader(t *testing.T) {
	initTestLogger()
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Unary()

	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: apiServicePath + "GetInternetBill"}
	_, err := interceptor(ctx, nil, info, nil)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptor_Unary_Restricted_InvalidToken(t *testing.T) {
	initTestLogger()
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Unary()

	md := metadata.New(map[string]string{"authorization": "Bearer invalid-token"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: apiServicePath + "GetInternetBill"}
	_, err := interceptor(ctx, nil, info, nil)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptor_Unary_TokenWithoutBearer(t *testing.T) {
	initTestLogger()
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Unary()

	token := generateTestToken("uid-1", "johndoe")
	md := metadata.New(map[string]string{"authorization": token})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: apiServicePath + "GetInternetBill"}
	_, err := interceptor(ctx, nil, info, handler)
	require.NoError(t, err)
	assert.True(t, called)
}

// ═══════════════════════════════════════════════════════════
// AuthInterceptor.Stream
// ═══════════════════════════════════════════════════════════

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context { return m.ctx }

func TestAuthInterceptor_Stream_Unrestricted(t *testing.T) {
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Stream()

	called := false
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		called = true
		return nil
	}

	info := &grpc.StreamServerInfo{FullMethod: apiServicePath + "GetProviders"}
	err := interceptor(nil, &mockServerStream{ctx: context.Background()}, info, handler)
	require.NoError(t, err)
	assert.True(t, called)
}

func TestAuthInterceptor_Stream_Restricted_NoMetadata(t *testing.T) {
	initTestLogger()
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Stream()

	info := &grpc.StreamServerInfo{FullMethod: apiServicePath + "GetInternetBill"}
	err := interceptor(nil, &mockServerStream{ctx: context.Background()}, info, nil)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptor_Stream_Restricted_ValidToken(t *testing.T) {
	initTestLogger()
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	interceptor := ai.Stream()

	token := generateTestToken("uid-1", "johndoe")
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	called := false
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		called = true
		return nil
	}

	info := &grpc.StreamServerInfo{FullMethod: apiServicePath + "GetInternetBill"}
	err := interceptor(nil, &mockServerStream{ctx: ctx}, info, handler)
	require.NoError(t, err)
	assert.True(t, called)
}

// ═══════════════════════════════════════════════════════════
// ProcessIdInterceptor
// ═══════════════════════════════════════════════════════════

func TestProcessIdInterceptor_NoExistingProcessId(t *testing.T) {
	interceptor := ProcessIdInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		pid := ctx.Value(constant.ProcessIdCtx).(string)
		assert.NotEmpty(t, pid)
		assert.Len(t, pid, 36) // UUID length
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test"}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(nil))
	_, err := interceptor(ctx, nil, info, handler)
	require.NoError(t, err)
}

func TestProcessIdInterceptor_WithExistingProcessId(t *testing.T) {
	interceptor := ProcessIdInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		pid := ctx.Value(constant.ProcessIdCtx).(string)
		assert.Equal(t, "existing-pid", pid)
		return "ok", nil
	}

	md := metadata.New(map[string]string{constant.ProcessIdCtx: "existing-pid"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/test"}
	_, err := interceptor(ctx, nil, info, handler)
	require.NoError(t, err)
}

func TestProcessIdInterceptor_TruncateLongProcessId(t *testing.T) {
	interceptor := ProcessIdInterceptor()

	longPid := "12345678-1234-1234-1234-1234567890ab-extra-long-suffix"

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		pid := ctx.Value(constant.ProcessIdCtx).(string)
		assert.Len(t, pid, 36)
		return "ok", nil
	}

	md := metadata.New(map[string]string{constant.ProcessIdCtx: longPid})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/test"}
	_, err := interceptor(ctx, nil, info, handler)
	require.NoError(t, err)
}

func TestProcessIdInterceptor_NoMetadata(t *testing.T) {
	interceptor := ProcessIdInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		pid := ctx.Value(constant.ProcessIdCtx).(string)
		assert.NotEmpty(t, pid)
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test"}
	_, err := interceptor(context.Background(), nil, info, handler)
	require.NoError(t, err)
}

// ═══════════════════════════════════════════════════════════
// ErrorsInterceptor
// ═══════════════════════════════════════════════════════════

func TestErrorsInterceptor_NoError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test"}
	out, err := ErrorsInterceptor(context.Background(), nil, info, handler)
	require.NoError(t, err)
	assert.Equal(t, "ok", out)
}

func TestErrorsInterceptor_NotFoundError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, db.NotFound("User", "123")
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test"}
	_, err := ErrorsInterceptor(context.Background(), nil, info, handler)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestErrorsInterceptor_OtherError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, assert.AnError
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test"}
	_, err := ErrorsInterceptor(context.Background(), nil, info, handler)
	assert.Error(t, err)
}

// ═══════════════════════════════════════════════════════════
// LoggingInterceptor
// ═══════════════════════════════════════════════════════════

func TestLoggingInterceptor_HealthCheck(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}
	out, err := LoggingInterceptor(context.Background(), nil, info, handler)
	require.NoError(t, err)
	assert.Equal(t, "ok", out)
}

func TestLoggingInterceptor_NormalCall(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "result", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: apiServicePath + "GetProviders"}
	out, err := LoggingInterceptor(context.Background(), nil, info, handler)
	require.NoError(t, err)
	assert.Equal(t, "result", out)
}

func TestLoggingInterceptor_WithError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, assert.AnError
	}
	info := &grpc.UnaryServerInfo{FullMethod: apiServicePath + "GetProviders"}
	_, err := LoggingInterceptor(context.Background(), nil, info, handler)
	assert.Error(t, err)
}

// ═══════════════════════════════════════════════════════════
// UnaryInterceptors / StreamInterceptors
// ═══════════════════════════════════════════════════════════

func TestUnaryInterceptors(t *testing.T) {
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	chain := UnaryInterceptors(ai)
	assert.NotNil(t, chain)
}

func TestStreamInterceptors(t *testing.T) {
	m := manager.NewJWTManager(testSecret)
	ai := NewAuthInterceptor(m)
	chain := StreamInterceptors(ai)
	assert.NotNil(t, chain)
}
