package utils

import (
	"context"
	"math/rand"
	"os"

	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/constant"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func GenerateProcessId() string {
	return uuid.New().String()
}

// GenerateCardNumber generates a random 16-digit card number string.
func GenerateCardNumber() string {
	digits := make([]byte, 16)
	for i := range digits {
		digits[i] = byte('0' + rand.Intn(10))
	}
	return string(digits)
}

// GenerateCardProvider returns a random card provider from the supported options.
func GenerateCardProvider() string {
	providers := []string{"Mastercard Platinum", "VISA", "GPN"}
	return providers[rand.Intn(len(providers))]
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
