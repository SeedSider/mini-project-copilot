package api

import (
	"context"
	"fmt"
	"strings"

	manager "github.com/bankease/payment-service/server/jwt"
	"github.com/bankease/payment-service/server/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const apiServicePath string = "/payment.PaymentService/"

type AuthInterceptor struct {
	jwtManager      *manager.JWTManager
	accessibleRoles map[string][]string
}

func NewAuthInterceptor(jwtManager *manager.JWTManager) *AuthInterceptor {
	return &AuthInterceptor{jwtManager, accessibleRoles()}
}

func accessibleRoles() map[string][]string {
	return map[string][]string{
		apiServicePath + "GetInternetBill": {},
	}
}

func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		if !interceptor.isRestricted(info.FullMethod) {
			return handler(ctx, req)
		}

		claims, err := interceptor.claimsToken(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, "user_claims", claims)

		return handler(ctx, req)
	}
}

func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		if !interceptor.isRestricted(info.FullMethod) {
			return handler(srv, stream)
		}

		_, err := interceptor.claimsToken(stream.Context())
		if err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (interceptor *AuthInterceptor) isRestricted(method string) bool {
	_, restricted := interceptor.accessibleRoles[method]
	return restricted
}

func (interceptor *AuthInterceptor) claimsToken(ctx context.Context) (*manager.UserClaims, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	log.Info(processId, "claimsToken", fmt.Sprint("md values : ", values), nil, nil, nil, nil)
	split := strings.Split(values[0], " ")
	accessToken := split[0]
	if len(split) > 1 {
		accessToken = split[1]
	}
	claims, err := interceptor.jwtManager.Verify(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	return claims, nil
}
