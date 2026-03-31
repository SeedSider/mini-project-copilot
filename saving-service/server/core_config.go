package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ListenAddress string

	CorsAllowedHeaders []string
	CorsAllowedMethods []string
	CorsAllowedOrigins []string

	LoggerTag    string
	LoggerOutput string
	LoggerLevel  string
	AppName      string

	Env           string
	ProductName   string
	FluentbitHost string
	FluentbitPort int
	TimeLocation  *time.Location

	DbHost     string
	DbUser     string
	DbPassword string
	DbName     string
	DbPort     string
	DbSslmode  string
	DbTimezone string

	HTTPPort string
	GRPCPort string
}

var config *Config

func initConfig() {
	godotenv.Load(".env")

	fluentbitPort, _ := strconv.Atoi(GetEnv("FLUENTBIT_PORT", "24223"))
	appName := GetEnv("APP_NAME", "saving-service")

	timeLocation, _ := time.LoadLocation(GetEnv("TIME_LOCATION", "Asia/Jakarta"))

	config = &Config{
		ListenAddress: fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")),

		CorsAllowedHeaders: []string{
			"Connection", "User-Agent", "Referer",
			"Accept", "Accept-Language", "Content-Type",
			"Content-Language", "Content-Disposition", "Origin",
			"Content-Length", "Authorization", "ResponseType",
			"X-Requested-With", "X-Forwarded-For",
		},
		CorsAllowedMethods: []string{"GET", "POST", "OPTIONS"},
		CorsAllowedOrigins: []string{"*"},

		LoggerTag:    GetEnv("LOGGER_TAG", "saving.dev"),
		LoggerOutput: GetEnv("LOGGER_OUTPUT", "stdout"),
		LoggerLevel:  GetEnv("LOGGER_LEVEL", "debug"),
		AppName:      appName,

		Env:           GetEnv("ENV", "DEV"),
		ProductName:   GetEnv("PRODUCT_NAME", "BRICaMS"),
		FluentbitHost: GetEnv("FLUENTBIT_HOST", "0.0.0.0"),
		FluentbitPort: fluentbitPort,
		TimeLocation:  timeLocation,

		DbHost:     GetEnv("DB_HOST", "localhost"),
		DbUser:     GetEnv("DB_USER", "postgres"),
		DbPassword: GetEnv("DB_PASSWORD", "postgres"),
		DbName:     GetEnv("DB_NAME", "saving"),
		DbPort:     GetEnv("DB_PORT", "5432"),
		DbSslmode:  GetEnv("DB_SSLMODE", "disable"),
		DbTimezone: GetEnv("DB_TIMEZONE", "Asia/Jakarta"),

		HTTPPort: GetEnv("HTTP_PORT", "8081"),
		GRPCPort: GetEnv("GRPC_PORT", "9303"),
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		config.DbHost, config.DbPort, config.DbUser, config.DbPassword,
		config.DbName, config.DbSslmode, config.DbTimezone)
}

func corsHeadersString() string {
	return strings.Join(config.CorsAllowedHeaders, ", ")
}
