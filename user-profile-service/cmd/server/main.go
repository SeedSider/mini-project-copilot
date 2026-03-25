package main

import (
	"log"
	"os"

	"github.com/bankease/user-profile-service/internal/db"
	"github.com/bankease/user-profile-service/internal/server"
	"github.com/joho/godotenv"
)

// GetEnv reads an environment variable with a fallback default.
// Pattern from: addons-issuance-lc-service/server/core_config.go
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Load .env file (ignore error if not present)
	godotenv.Load()

	databaseURL := GetEnv("DATABASE_URL", "")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	port := GetEnv("PORT", "8080")

	// Establish database connection
	database, err := db.NewDB(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create and start server
	srv := server.NewServer(database, port)

	log.Printf("Server started on :%s", port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
