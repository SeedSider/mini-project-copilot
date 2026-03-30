package server

import (
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/bankease/user-profile-service/internal/grpchandler"
	"github.com/bankease/user-profile-service/internal/handlers"
	"github.com/bankease/user-profile-service/internal/repository"
	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Server holds all dependencies and the HTTP router.
// Pattern from: addons-issuance-lc-service/server/main.go (DI via struct)
type Server struct {
	DB          *sql.DB
	Router      chi.Router
	Port        string
	GRPCPort    string
	profileRepo      *repository.ProfileRepository
	menuRepo         *repository.MenuRepository
	exchangeRateRepo *repository.ExchangeRateRepository
	interestRateRepo *repository.InterestRateRepository
	branchRepo       *repository.BranchRepository
}

// NewServer creates a new Server with all dependencies wired up.
func NewServer(db *sql.DB, port string, azureSASURL string, azureContainer string, jwtSecret string, grpcPort string) *Server {
	profileRepo := &repository.ProfileRepository{DB: db}
	menuRepo := &repository.MenuRepository{DB: db}
	exchangeRateRepo := &repository.ExchangeRateRepository{DB: db}
	interestRateRepo := &repository.InterestRateRepository{DB: db}
	branchRepo := &repository.BranchRepository{DB: db}

	profileHandler := &handlers.ProfileHandler{Repo: profileRepo, JWTSecret: jwtSecret}
	menuHandler := &handlers.MenuHandler{Repo: menuRepo}
	uploadHandler := &handlers.UploadHandler{
		AzureSASURL:    azureSASURL,
		AzureContainer: azureContainer,
	}
	searchHandler := &handlers.SearchHandler{
		ExchangeRateRepo: exchangeRateRepo,
		InterestRateRepo: interestRateRepo,
		BranchRepo:       branchRepo,
	}

	s := &Server{
		DB:               db,
		Port:             port,
		GRPCPort:         grpcPort,
		profileRepo:      profileRepo,
		menuRepo:         menuRepo,
		exchangeRateRepo: exchangeRateRepo,
		interestRateRepo: interestRateRepo,
		branchRepo:       branchRepo,
	}

	s.Router = setupRoutes(profileHandler, menuHandler, uploadHandler, searchHandler)
	return s
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.Port, s.Router)
}

// StartGRPC starts the gRPC server on GRPCPort.
func (s *Server) StartGRPC() error {
	listener, err := net.Listen("tcp", ":"+s.GRPCPort)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	grpcHandler := &grpchandler.GrpcServer{
		ProfileRepo:      s.profileRepo,
		MenuRepo:         s.menuRepo,
		ExchangeRateRepo: s.exchangeRateRepo,
		InterestRateRepo: s.interestRateRepo,
		BranchRepo:       s.branchRepo,
	}

	pb.RegisterUserProfileServiceServer(grpcServer, grpcHandler)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	log.Printf("gRPC server started on :%s", s.GRPCPort)
	return grpcServer.Serve(listener)
}
