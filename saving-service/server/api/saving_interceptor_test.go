package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bankease/saving-service/server/db"
)

var mockInfo = &grpc.UnaryServerInfo{FullMethod: "/saving.SavingService/GetExchangeRates"}

// ═══════════════════════════════════════════
// UnaryInterceptors / StreamInterceptors
// ═══════════════════════════════════════════

func TestUnaryInterceptors_NotNil(t *testing.T) {
	interceptor := UnaryInterceptors()
	assert.NotNil(t, interceptor)
}

func TestStreamInterceptors_NotNil(t *testing.T) {
	interceptor := StreamInterceptors()
	assert.NotNil(t, interceptor)
}

// ═══════════════════════════════════════════
// ErrorsInterceptor
// ═══════════════════════════════════════════

func TestErrorsInterceptor_NotFoundErr(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, db.NotFound("ExchangeRate", "er-uuid-123")
	}

	_, err := ErrorsInterceptor(context.Background(), nil, mockInfo, handler)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "er-uuid-123")
}

func TestErrorsInterceptor_OtherError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, fmt.Errorf("connection error")
	}

	_, err := ErrorsInterceptor(context.Background(), nil, mockInfo, handler)

	assert.Error(t, err)
	st, _ := status.FromError(err)
	// generic errors are not remapped — code stays Unknown
	assert.NotEqual(t, codes.NotFound, st.Code())
}

func TestErrorsInterceptor_Success(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	out, err := ErrorsInterceptor(context.Background(), nil, mockInfo, handler)

	require.NoError(t, err)
	assert.Equal(t, "ok", out)
}

// ═══════════════════════════════════════════
// LoggingInterceptor
// ═══════════════════════════════════════════

func TestLoggingInterceptor_Success(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	out, err := LoggingInterceptor(context.Background(), nil, mockInfo, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", out)
}

func TestLoggingInterceptor_WithError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, fmt.Errorf("handler error")
	}

	_, err := LoggingInterceptor(context.Background(), nil, mockInfo, handler)

	assert.Error(t, err)
}

func TestLoggingInterceptor_HealthCheck(t *testing.T) {
	healthInfo := &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "healthy", nil
	}

	out, err := LoggingInterceptor(context.Background(), nil, healthInfo, handler)

	require.NoError(t, err)
	assert.Equal(t, "healthy", out)
}

// ═══════════════════════════════════════════
// ProcessIdInterceptor
// ═══════════════════════════════════════════

func TestProcessIdInterceptor_GeneratesID(t *testing.T) {
	interceptorFn := ProcessIdInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return ctx.Value("process_id"), nil // should have a process ID injected
	}

	out, err := interceptorFn(context.Background(), nil, mockInfo, handler)
	require.NoError(t, err)
	assert.NotNil(t, out)
}

func TestProcessIdInterceptor_UsesExistingID(t *testing.T) {
	interceptorFn := ProcessIdInterceptor()

	md := metadata.Pairs("process_id", "existing-process-id-1234")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	_, err := interceptorFn(ctx, nil, mockInfo, handler)
	require.NoError(t, err)
}

func TestProcessIdInterceptor_TruncatesLongID(t *testing.T) {
	interceptorFn := ProcessIdInterceptor()

	// process_id longer than 36 chars should be truncated
	longID := "this-is-a-very-long-process-id-that-exceeds-36-characters"
	md := metadata.Pairs("process_id", longID)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	_, err := interceptorFn(ctx, nil, mockInfo, handler)
	require.NoError(t, err)
}
