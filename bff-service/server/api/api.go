package api

import (
	"time"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	manager "github.com/bankease/bff-service/server/jwt"
	"github.com/bankease/bff-service/server/lib/logger"
	svc "github.com/bankease/bff-service/server/services"
)

var log *logger.Logger

// Server implements BffServiceServer and orchestrates calls to downstream services.
type Server struct {
	manager *manager.JWTManager
	svcConn *svc.ServiceConnection
	logger  *logger.Logger

	pb.UnimplementedBffServiceServer
}

// Ensure Server implements BffServiceServer at compile time.
var _ pb.BffServiceServer = (*Server)(nil)

func New(jwtSecret, jwtDuration string, svcConn *svc.ServiceConnection, logger *logger.Logger) *Server {
	log = logger
	tokenDuration, err := time.ParseDuration(jwtDuration)
	if err != nil {
		log.Fatal("", "New", "Failed when Parsing Duration", nil, nil, nil, err)
	}

	return &Server{
		manager: manager.NewJWTManager(jwtSecret, tokenDuration),
		svcConn: svcConn,
		logger:  logger,
	}
}

func (s *Server) GetManager() *manager.JWTManager {
	return s.manager
}
