package api

import (
	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"github.com/bankease/user-profile-service/server/db"
)

// Compile-time check that *Server implements pb.UserProfileServiceServer.
var _ pb.UserProfileServiceServer = (*Server)(nil)

// Server holds all dependencies and implements both HTTP and gRPC handlers.
// Pattern from: identity-service/server/api/api.go
type Server struct {
	pb.UnimplementedUserProfileServiceServer
	provider       *db.Provider
	jwtSecret      string
	azureSASURL    string
	azureContainer string
}

// New creates a new Server with all dependencies wired up.
func New(
	provider *db.Provider,
	jwtSecret string,
	azureSASURL string,
	azureContainer string,
) *Server {
	return &Server{
		provider:       provider,
		jwtSecret:      jwtSecret,
		azureSASURL:    azureSASURL,
		azureContainer: azureContainer,
	}
}
