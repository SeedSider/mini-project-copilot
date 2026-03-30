package utils

import (
	"context"
	"os"

	"github.com/bankease/bff-service/server/constant"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GenerateProcessId() string {
	return uuid.New().String()
}

func GetProcessIdFromCtx(ctx context.Context) string {
	processId, ok := ctx.Value(constant.ProcessIdCtx).(string)
	if !ok {
		return ""
	}
	return processId
}

func CreateNewContextWithProcessId(ctx context.Context, md metadata.MD) context.Context {
	processId := GetProcessIdFromCtx(ctx)

	if md == nil {
		md = make(metadata.MD)
	}
	md[constant.ProcessIdCtx] = []string{processId}

	newCtx := context.Background()
	newCtx = context.WithValue(newCtx, constant.ProcessIdCtx, processId)
	newCtx = metadata.NewIncomingContext(newCtx, md)
	newCtx = metadata.AppendToOutgoingContext(newCtx, constant.ProcessIdCtx, processId)

	return metadata.NewOutgoingContext(newCtx, md)
}
