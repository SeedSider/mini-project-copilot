package api

import (
	pb "github.com/bankease/payment-service/protogen/payment-service"
	"github.com/bankease/payment-service/server/db"
	manager "github.com/bankease/payment-service/server/jwt"
	"github.com/bankease/payment-service/server/lib/logger"
)

var log *logger.Logger

// Compile-time check that *Server implements pb.PaymentServiceServer.
var _ pb.PaymentServiceServer = (*Server)(nil)

type Server struct {
	pb.UnimplementedPaymentServiceServer
	provider *db.Provider
	manager  *manager.JWTManager
	logger   *logger.Logger
}

func New(
	jwtSecret string,
	provider *db.Provider,
	logger *logger.Logger,
) *Server {
	log = logger
	return &Server{
		provider: provider,
		manager:  manager.NewJWTManager(jwtSecret),
		logger:   logger,
	}
}

func (s *Server) GetManager() *manager.JWTManager {
	return s.manager
}
