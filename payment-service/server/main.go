package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	pb "github.com/bankease/payment-service/protogen/payment-service"
	"github.com/bankease/payment-service/server/api"
	"github.com/bankease/payment-service/server/db"
	manager "github.com/bankease/payment-service/server/jwt"
	"github.com/bankease/payment-service/server/lib/logger"

	_ "github.com/bankease/payment-service/docs"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const serviceName = "payment"
const defaultPort = 9304

var (
	log        *logger.Logger
	grpcServer *grpc.Server
	jwtMgr     *manager.JWTManager
)

func init() {
	initConfig()

	time.Local = config.TimeLocation

	log = logger.New(&logger.LoggerConfig{
		Env:           config.Env,
		ServiceName:   config.LoggerTag,
		ProductName:   config.ProductName,
		LogLevel:      config.LoggerLevel,
		LogOutput:     config.LoggerOutput,
		FluentbitHost: config.FluentbitHost,
		FluentbitPort: config.FluentbitPort,
	})

	jwtMgr = manager.NewJWTManager(config.JWTSecret)
}

// @title           Payment Service API
// @version         1.0
// @description     BankEase Payment Service — providers, internet-bill, currency-list
// @host            localhost:8082
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	initConfig()

	app := cli.NewApp()
	app.Name = ""
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
		Usage: "Starts gRPC and HTTP server",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "port1",
				Value: defaultPort,
			},
			cli.IntFlag{
				Name:  "port2",
				Value: 8082,
			},
			cli.StringFlag{
				Name:  "grpc-endpoint",
				Value: ":" + fmt.Sprint(defaultPort),
				Usage: "the address of the running gRPC server to transcode to",
			},
		},
		Action: func(c *cli.Context) error {
			grpcPort := c.Int("port1")
			httpPort := c.Int("port2")

			startDBConnection()
			runMigration()

			// Start gRPC server
			go func() {
				if err := startGrpcServer(grpcPort); err != nil {
					log.Fatal("", "grpcGatewayServerCmd", fmt.Sprintf("failed gRPC serve: %v", err), nil, nil, nil, err)
				}
			}()

			// Start HTTP server
			go func() {
				if err := httpServer(httpPort); err != nil {
					log.Fatal("", "grpcGatewayServerCmd", fmt.Sprintf("failed HTTP serve: %v", err), nil, nil, nil, err)
				}
			}()

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			<-ch

			closeDBConnection()
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

func httpServer(port int) error {
	log.Info("", "httpServer", fmt.Sprintf("Starting %s Service ................", serviceName), nil, nil, nil, nil)
	log.Info("", "httpServer", fmt.Sprintf("Starting HTTP server on port %d...", port), nil, nil, nil, nil)

	dbProvider := db.New(dbSql, log)
	apiServer := api.New(
		config.JWTSecret,
		dbProvider,
		log,
	)

	mux := http.NewServeMux()

	// Public API routes
	mux.HandleFunc("/api/pay-the-bill/providers", methodOnly("GET", apiServer.HandleGetProviders))
	mux.HandleFunc("/api/currency-list", methodOnly("GET", apiServer.HandleGetCurrencyList))

	// Protected API route (JWT required)
	mux.HandleFunc("/api/pay-the-bill/internet-bill", methodOnly("GET", jwtMiddleware(apiServer.HandleGetInternetBill)))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	return http.Serve(listener, cors(mux))
}

func startGrpcServer(port int) error {
	log.Info("", "startGrpcServer", fmt.Sprintf("Starting gRPC server on port %d...", port), nil, nil, nil, nil)

	list, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	dbProvider := db.New(dbSql, log)
	apiServer := api.New(
		config.JWTSecret,
		dbProvider,
		log,
	)

	authInterceptor := api.NewAuthInterceptor(apiServer.GetManager())

	grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(api.UnaryInterceptors(authInterceptor)),
		grpc.StreamInterceptor(api.StreamInterceptors(authInterceptor)),
	)

	pb.RegisterPaymentServiceServer(grpcServer, apiServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	log.Info("", "startGrpcServer", fmt.Sprintf("gRPC server listening on port %d", port), nil, nil, nil, nil)
	return grpcServer.Serve(list)
}

// jwtMiddleware verifies JWT for HTTP endpoints and injects user_claims into context.
func jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   true,
				"code":    401,
				"message": "Unauthorized",
			})
			return
		}

		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		claims, err := jwtMgr.Verify(token)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   true,
				"code":    401,
				"message": "Unauthorized",
			})
			return
		}

		ctx := context.WithValue(r.Context(), "user_claims", claims)
		next(w, r.WithContext(ctx))
	}
}

func methodOnly(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method != method {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}
		next(w, r)
	}
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Content-Security-Policy", "object-src 'none'; child-src 'none'; script-src 'unsafe-inline' https: http:")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "no-referrer")

		w.Header().Set("Access-Control-Allow-Origin", strings.Join(config.CorsAllowedOrigins, ", "))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.CorsAllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.CorsAllowedHeaders, ", "))

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
