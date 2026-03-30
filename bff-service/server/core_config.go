package main

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BffGRPCPort string
	BffHTTPPort string

	IdentityServiceAddr string
	ProfileServiceAddr  string

	JWTSecret   string
	JWTDuration string

	AzureSASURL    string
	AzureContainer string

	Env          string
	AppName      string
	ProductName  string
	LoggerLevel  string
	LoggerOutput string

	CorsAllowedHeaders []string
	CorsAllowedMethods []string
	CorsAllowedOrigins []string
}

var config *Config

func initConfig() {
	godotenv.Load(".env")

	config = &Config{
		BffGRPCPort: GetEnv("BFF_GRPC_PORT", "9090"),
		BffHTTPPort: GetEnv("BFF_HTTP_PORT", "3000"),

		IdentityServiceAddr: GetEnv("BFF_IDENTITY_SERVICE_ADDR", "localhost:9301"),
		ProfileServiceAddr:  GetEnv("BFF_PROFILE_SERVICE_ADDR", "localhost:9302"),

		JWTSecret:   GetEnv("BFF_JWT_SECRET", "secret"),
		JWTDuration: GetEnv("JWT_DURATION", "24h"),

		AzureSASURL:    GetEnv("BFF_AZURE_SAS_URL", ""),
		AzureContainer: GetEnv("BFF_AZURE_CONTAINER", "images"),

		Env:          GetEnv("ENV", "DEV"),
		AppName:      GetEnv("APP_NAME", "bff-service"),
		ProductName:  GetEnv("PRODUCT_NAME", "BankEase"),
		LoggerLevel:  GetEnv("LOGGER_LEVEL", "debug"),
		LoggerOutput: GetEnv("LOGGER_OUTPUT", "stdout"),

		CorsAllowedHeaders: []string{
			"Connection", "User-Agent", "Referer",
			"Accept", "Accept-Language", "Content-Type",
			"Content-Language", "Content-Disposition", "Origin",
			"Content-Length", "Authorization", "ResponseType",
			"X-Requested-With", "X-Forwarded-For", "Idempotency-Key",
		},
		CorsAllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
		CorsAllowedOrigins: []string{"*"},
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
