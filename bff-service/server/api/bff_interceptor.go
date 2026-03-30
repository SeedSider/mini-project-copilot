package api

import (
	"context"
	"time"

	"github.com/bankease/bff-service/server/constant"
	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryInterceptors(
	authI *AuthInterceptor,
) grpc.UnaryServerInterceptor {
	return grpc_middleware.ChainUnaryServer(
		ProcessIdInterceptor(),
		LoggingInterceptor,
		authI.Unary(),
	)
}

func StreamInterceptors(
	authI *AuthInterceptor,
) grpc.StreamServerInterceptor {
	return grpc_middleware.ChainStreamServer(
		authI.Stream(),
	)
}

func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (out interface{}, err error) {
	if info.FullMethod == "/grpc.health.v1.Health/Check" {
		out, err = handler(ctx, req)
		return out, err
	}

	entry := logrus.WithField("method", info.FullMethod)
	start := time.Now()
	out, err = handler(ctx, req)
	duration := time.Since(start)

	if err != nil {
		entry = entry.WithError(err)
	}

	entry.WithField("duration", duration.String()).Info("finished RPC")
	return out, err
}

func ProcessIdInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		processIds := md.Get(constant.ProcessIdCtx)
		var processId string
		if len(processIds) > 0 {
			processId = processIds[0]
			if len(processId) > 36 {
				processId = processId[:36]
			}
		} else {
			processId = uuid.New().String()
			md.Set(constant.ProcessIdCtx, processId)
			ctx = metadata.NewIncomingContext(ctx, md)
		}

		ctx = context.WithValue(ctx, constant.ProcessIdCtx, processId)
		ctx = metadata.AppendToOutgoingContext(ctx, constant.ProcessIdCtx, processId)

		return handler(ctx, req)
	}
}
