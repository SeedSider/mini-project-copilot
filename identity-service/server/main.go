package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/api"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/db"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/logger"

	"github.com/urfave/cli"
)

const serviceName = "identity"
const defaultPort = 9090

var (
	log *logger.Logger
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
}

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
		Usage: "Starts HTTP server",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "port1",
				Value: defaultPort,
			},
			cli.IntFlag{
				Name:  "port2",
				Value: 3000,
			},
			cli.StringFlag{
				Name:  "grpc-endpoint",
				Value: ":" + fmt.Sprint(defaultPort),
				Usage: "the address of the running gRPC server to transcode to",
			},
		},
		Action: func(c *cli.Context) error {
			httpPort := c.Int("port2")

			startDBConnection()
			runMigration()

			go func() {
				if err := httpServer(httpPort); err != nil {
					log.Fatal("", "grpcGatewayServerCmd", fmt.Sprintf("failed HTTP serve: %v", err), nil, nil, nil, err)
				}
			}()

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			<-ch

			closeDBConnections()
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
		config.JWTDuration,
		dbSql,
		log,
		dbProvider,
		config.ProfileServiceURL,
	)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/auth/signup", methodOnly("POST", apiServer.HandleSignUp))
	mux.HandleFunc("/api/auth/signin", methodOnly("POST", apiServer.HandleSignIn))
	mux.HandleFunc("/api/identity/me", methodOnly("GET", apiServer.HandleGetMe))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Swagger
	mux.HandleFunc("/api/identity/docs/swagger.json", serveSwagger)
	fs := http.FileServer(http.Dir("www/swagger-ui"))
	mux.Handle("/api/identity/docs/", http.StripPrefix("/api/identity/docs/", fs))

	// Serve /swagger/ as static files from www/swagger
	swaggerFs := http.FileServer(http.Dir("www/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", swaggerFs))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	return http.Serve(listener, cors(mux))
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

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "www/swagger.json")
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

func runMigration() {
	log.Info("", "runMigration", "Running database migration...", nil, nil, nil, nil)

	migrations := []string{
		// 001: Create users table
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			phone VARCHAR(50),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// 002: Rename email→username if old schema exists (safe: skips if already renamed)
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'email'
			) THEN
				ALTER TABLE users RENAME COLUMN email TO username;
			END IF;
		END $$;`,
		// 002b: Add phone column if missing (old schema may not have it)
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'phone'
			) THEN
				ALTER TABLE users ADD COLUMN phone VARCHAR(50);
			END IF;
		END $$;`,
	}

	for _, m := range migrations {
		_, err := dbSql.GetPmConnection().Exec(m)
		if err != nil {
			log.Fatal("", "runMigration", fmt.Sprintf("Migration failed: %v", err), nil, nil, nil, err)
		}
	}

	log.Info("", "runMigration", "Migration completed", nil, nil, nil, nil)
}
