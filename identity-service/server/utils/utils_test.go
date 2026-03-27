package utils

import (
	"context"
	"testing"

	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/constant"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestGetEnv_Exists(t *testing.T) {
	t.Setenv("TEST_KEY", "test_value")
	assert.Equal(t, "test_value", GetEnv("TEST_KEY", "fallback"))
}

func TestGetEnv_Fallback(t *testing.T) {
	assert.Equal(t, "fallback", GetEnv("NON_EXISTENT_KEY_12345", "fallback"))
}

func TestGenerateProcessId(t *testing.T) {
	id := GenerateProcessId()
	assert.NotEmpty(t, id)
	assert.Len(t, id, 36) // UUID format

	id2 := GenerateProcessId()
	assert.NotEqual(t, id, id2)
}

func TestGetProcessIdFromCtx_WithValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), constant.ProcessIdCtx, "test-process-id")
	assert.Equal(t, "test-process-id", GetProcessIdFromCtx(ctx))
}

func TestGetProcessIdFromCtx_Empty(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", GetProcessIdFromCtx(ctx))
}

func TestGetProcessIdFromCtx_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), constant.ProcessIdCtx, 12345)
	assert.Equal(t, "", GetProcessIdFromCtx(ctx))
}

func TestCreateNewContextWithProcessId(t *testing.T) {
	ctx := context.WithValue(context.Background(), constant.ProcessIdCtx, "pid-123")
	newCtx := CreateNewContextWithProcessId(ctx, nil)

	// Verify process_id is in outgoing metadata
	md, ok := metadata.FromOutgoingContext(newCtx)
	assert.True(t, ok)
	vals := md.Get(constant.ProcessIdCtx)
	assert.Contains(t, vals, "pid-123")
}

func TestCreateNewContextWithProcessId_WithExistingMD(t *testing.T) {
	ctx := context.WithValue(context.Background(), constant.ProcessIdCtx, "pid-456")
	existingMD := metadata.New(map[string]string{"other": "value"})
	newCtx := CreateNewContextWithProcessId(ctx, existingMD)

	md, ok := metadata.FromOutgoingContext(newCtx)
	assert.True(t, ok)
	vals := md.Get(constant.ProcessIdCtx)
	assert.Contains(t, vals, "pid-456")
}
