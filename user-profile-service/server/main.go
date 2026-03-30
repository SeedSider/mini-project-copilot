package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	_ "github.com/bankease/user-profile-service/docs"
	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"github.com/bankease/user-profile-service/server/api"
	"github.com/bankease/user-profile-service/server/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// @title        BankEase User Profile API
// @version      1.0
// @description  REST API for managing user profiles and homepage menus in BankEase mobile banking app.
// @host
// @BasePath     /

const serviceName = "user-profile"

var grpcServer *grpc.Server

func main() {
	initConfig()

	startDBConnection()
	runMigration()

	dbProvider := db.New(dbSql)
	apiServer := api.New(
		dbProvider,
		config.JWTSecret,
		config.AzureSASURL,
		config.AzureContainer,
	)

	// Start gRPC server
	go func() {
		if err := startGrpcServer(apiServer); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Start HTTP server
	go func() {
		if err := httpServer(apiServer); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	closeDBConnection()
	if grpcServer != nil {
		log.Println("Stopping gRPC server")
		grpcServer.GracefulStop()
		log.Println("gRPC server stopped")
	}
	log.Println("Server stopped")
}

func httpServer(apiServer *api.Server) error {
	log.Printf("Starting %s HTTP server on port %s...", serviceName, config.Port)

	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Profile routes
	r.Get("/api/profile", apiServer.HandleGetMyProfile)
	r.Post("/api/profile", apiServer.HandleCreateProfile)
	r.Get("/api/profile/user/{user_id}", apiServer.HandleGetProfileByUserID)
	r.Get("/api/profile/{id}", apiServer.HandleGetProfile)
	r.Put("/api/profile/{id}", apiServer.HandleUpdateProfile)

	// Menu routes
	r.Get("/api/menu", apiServer.HandleGetAllMenus)
	r.Get("/api/menu/{accountType}", apiServer.HandleGetMenusByAccountType)

	// Upload routes
	r.Post("/api/upload/image", apiServer.HandleUploadImage)

	// Search / rates / branches routes
	r.Get("/api/exchange-rates", apiServer.HandleGetExchangeRates)
	r.Get("/api/interest-rates", apiServer.HandleGetInterestRates)
	r.Get("/api/branches", apiServer.HandleGetBranches)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Port))
	if err != nil {
		return err
	}

	log.Printf("HTTP server listening on port %s", config.Port)
	return http.Serve(listener, r)
}

func startGrpcServer(apiServer *api.Server) error {
	log.Printf("Starting gRPC server on port %s...", config.GRPCPort)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		return err
	}

	grpcServer = grpc.NewServer()

	pb.RegisterUserProfileServiceServer(grpcServer, apiServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	log.Printf("gRPC server listening on port %s", config.GRPCPort)
	return grpcServer.Serve(listener)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Content-Security-Policy", "object-src 'none'; child-src 'none'; script-src 'unsafe-inline' https: http:")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "no-referrer")

		w.Header().Set("Access-Control-Allow-Origin", strings.Join([]string{"*"}, ", "))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
