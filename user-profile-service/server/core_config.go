package main

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL    string
	Port           string
	GRPCPort       string
	AzureSASURL    string
	AzureContainer string
	JWTSecret      string
}

var config *Config

func initConfig() {
	godotenv.Load(".env")

	config = &Config{
		DatabaseURL:    GetEnv("DATABASE_URL", ""),
		Port:           GetEnv("PORT", "8080"),
		GRPCPort:       GetEnv("GRPC_PORT", "9302"),
		AzureSASURL:    GetEnv("AZURE_STORAGE_SAS_URL", ""),
		AzureContainer: GetEnv("AZURE_STORAGE_CONTAINER", "images"),
		JWTSecret:      GetEnv("JWT_SECRET", "secret"),
	}
}

// GetEnv reads an environment variable with a fallback default.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
