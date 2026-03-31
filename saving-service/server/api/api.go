package api

import (
	pb "github.com/bankease/saving-service/protogen/saving-service"
	"github.com/bankease/saving-service/server/db"
	"github.com/bankease/saving-service/server/lib/logger"
)

var log *logger.Logger

// Compile-time check that *Server implements pb.SavingServiceServer.
var _ pb.SavingServiceServer = (*Server)(nil)

// Server implements both HTTP and gRPC handlers for the saving service.
type Server struct {
	pb.UnimplementedSavingServiceServer
	provider *db.Provider
	logger   *logger.Logger
}

// New creates a new Server with all dependencies wired up.
func New(provider *db.Provider, logger *logger.Logger) *Server {
	log = logger
	return &Server{
		provider: provider,
		logger:   logger,
	}
}
