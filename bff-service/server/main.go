package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/bankease/bff-service/docs"
	pb "github.com/bankease/bff-service/protogen/bff-service"
	"github.com/bankease/bff-service/server/api"
	jwts "github.com/bankease/bff-service/server/jwt"
	"github.com/bankease/bff-service/server/lib/logger"
	"github.com/bankease/bff-service/server/services"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const serviceName = "bff"

var (
	log        *logger.Logger
	grpcServer *grpc.Server
	jwtMgr     *jwts.JWTManager
)

type gatewayServer struct {
	apiServer *api.Server
}

func init() {
	initConfig()

	log = logger.New(&logger.LoggerConfig{
		Env:         config.Env,
		ServiceName: config.AppName,
		ProductName: config.ProductName,
		LogLevel:    config.LoggerLevel,
		LogOutput:   config.LoggerOutput,
	})
}

// @title BFF Service API
// @version 1.0
// @description Backend for Frontend (BFF) service for BankEase mobile banking application. Single entry point for the mobile app, orchestrating calls to identity-service and user-profile-service.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token with the `Bearer ` prefix, e.g. "Bearer eyJhbGciOi..."
func main() {
	initConfig()

	app := cli.NewApp()
	app.Name = "bff-service"
	app.Commands = []cli.Command{
		grpcGatewayServerCmd(),
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err.Error())
		os.Exit(1)
	}
}

func grpcGatewayServerCmd() cli.Command {
	return cli.Command{
		Name:  "grpc-gw-server",
		Usage: "Starts gRPC and HTTP gateway server",
		Action: func(c *cli.Context) error {
			grpcPort := config.BffGRPCPort
			httpPort := config.BffHTTPPort

			// Initialize service connections to downstream services
			svcConn := services.InitServicesConn(
				config.IdentityServiceAddr,
				config.ProfileServiceAddr,
				config.SavingServiceAddr,
				config.PaymentServiceAddr,
			)
			defer svcConn.Close()

			apiServer := api.New(
				config.JWTSecret,
				config.JWTDuration,
				svcConn,
				log,
			)

			// Start gRPC server
			go func() {
				if err := startGrpcServer(grpcPort, apiServer); err != nil {
					log.Fatal("", "grpcGatewayServerCmd", fmt.Sprintf("failed gRPC serve: %v", err), nil, nil, nil, err)
				}
			}()

			// Start HTTP server
			go func() {
				if err := startHTTPServer(httpPort, apiServer); err != nil {
					log.Fatal("", "grpcGatewayServerCmd", fmt.Sprintf("failed HTTP serve: %v", err), nil, nil, nil, err)
				}
			}()

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			<-ch

			if grpcServer != nil {
				log.Info("", "grpcGatewayServerCmd", "Stopping gRPC server", nil, nil, nil, nil)
				grpcServer.GracefulStop()
				log.Info("", "grpcGatewayServerCmd", "gRPC server stopped", nil, nil, nil, nil)
			}
			log.Info("", "grpcGatewayServerCmd", "Server stopped", nil, nil, nil, nil)

			return nil
		},
	}
}

func startGrpcServer(port string, apiServer *api.Server) error {
	log.Info("", "startGrpcServer", fmt.Sprintf("Starting gRPC server on port %s...", port), nil, nil, nil, nil)

	list, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	authInterceptor := api.NewAuthInterceptor(apiServer.GetManager())

	grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(api.UnaryInterceptors(authInterceptor)),
		grpc.StreamInterceptor(api.StreamInterceptors(authInterceptor)),
	)

	pb.RegisterBffServiceServer(grpcServer, apiServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	log.Info("", "startGrpcServer", fmt.Sprintf("gRPC server listening on port %s", port), nil, nil, nil, nil)
	return grpcServer.Serve(list)
}

func startHTTPServer(port string, apiServer *api.Server) error {
	log.Info("", "startHTTPServer", fmt.Sprintf("Starting %s HTTP gateway on port %s...", serviceName, port), nil, nil, nil, nil)

	jwtMgr = apiServer.GetManager()
	gwServer := &gatewayServer{apiServer: apiServer}

	mux := http.NewServeMux()

	// Custom HTTP handler for upload (multipart/form-data, not through grpc-gateway)
	mux.HandleFunc("/api/upload/image", gwServer.HandleUploadImage)

	// Swagger UI
	mux.Handle("/swagger/bff/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/bff/doc.json"),
	))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// gRPC-gateway style REST endpoints (manually implemented via HTTP → gRPC bridge)
	mux.HandleFunc("/api/auth/signup", gwServer.handleAuthSignUp)
	mux.HandleFunc("/api/auth/signin", gwServer.handleAuthSignIn)
	mux.HandleFunc("/api/auth/me", gwServer.handleAuthGetMe)
	mux.HandleFunc("/api/auth/validate-otp", gwServer.handleValidateOtp)
	mux.HandleFunc("/api/auth/update-password", gwServer.handleUpdatePassword)

	// Profile endpoints
	mux.HandleFunc("/api/profile/user/", gwServer.handleProfileByUserID)
	mux.HandleFunc("/api/profile/", gwServer.handleProfileByID)
	mux.HandleFunc("/api/profile", gwServer.handleProfile)

	// Menu endpoints
	mux.HandleFunc("/api/menu/", gwServer.handleMenuByAccountType)
	mux.HandleFunc("/api/menu", gwServer.handleMenuAll)

	// Search / Saving endpoints
	mux.HandleFunc("/api/exchange-rates", gwServer.handleExchangeRates)
	mux.HandleFunc("/api/interest-rates", gwServer.handleInterestRates)
	mux.HandleFunc("/api/branches", gwServer.handleBranches)

	// Payment endpoints
	mux.HandleFunc("/api/pay-the-bill/providers", gwServer.handleProviders)
	mux.HandleFunc("/api/pay-the-bill/internet-bill", gwServer.handleInternetBill)
	mux.HandleFunc("/api/currency-list", gwServer.handleCurrencyList)

	// Mobile Prepaid endpoints
	mux.HandleFunc("/api/mobile-prepaid/beneficiaries", gwServer.handleGetBeneficiaries)
	mux.HandleFunc("/api/mobile-prepaid/pay", gwServer.handlePrepaidPay)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	log.Info("", "startHTTPServer", fmt.Sprintf("HTTP gateway listening on port %s", port), nil, nil, nil, nil)
	return http.Serve(listener, corsMiddleware(mux))
}
