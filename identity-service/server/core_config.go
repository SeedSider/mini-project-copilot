package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ListenAddress string `config:"LISTEN_ADDRESS"`

	CorsAllowedHeaders []string `config:"CORS_ALLOWED_HEADERS"`
	CorsAllowedMethods []string `config:"CORS_ALLOWED_METHODS"`
	CorsAllowedOrigins []string `config:"CORS_ALLOWED_ORIGINS"`
	ExposedHeaders     []string `config:"EXPOSED_HEADERS"`

	JWTSecret   string `config:"JWT_SECRET"`
	JWTDuration string `config:"JWT_DURATION"`

	LoggerTag    string `config:"LOGGER_TAG"`
	LoggerOutput string `config:"LOGGER_OUTPUT"`
	LoggerLevel  string `config:"LOGGER_LEVEL"`
	AppName      string `config:"APP_NAME"`

	GrpcMaxCallSendMessage int `config:"GRPC_MAX_CALL_SEND_MESSAGE"`
	GrpcMaxCallRecvMessage int `config:"GRPC_MAX_CALL_RECEIVE_MESSAGE"`

	Env           string         `config:"ENV"`
	ProductName   string         `config:"PRODUCT_NAME"`
	FluentbitHost string         `config:"FLUENTBIT_HOST"`
	FluentbitPort int            `config:"FLUENTBIT_PORT"`
	TimeLocation  *time.Location `config:"TIME_LOCATION"`

	DbHost     string `config:"DB_HOST"`
	DbUser     string `config:"DB_USER"`
	DbPassword string `config:"DB_PASSWORD"`
	DbName     string `config:"DB_NAME"`
	DbPort     string `config:"DB_PORT"`
	DbSslmode  string `config:"DB_SSLMODE"`
	DbTimezone string `config:"DB_TIMEZONE"`
	DbMaxRetry string `config:"DB_MAX_RETRY"`
	DbTimeout  string `config:"DB_TIMEOUT"`

	ProfileServiceURL string `config:"PROFILE_SERVICE_URL"`
}

var config *Config

func initConfig() {
	godotenv.Load(".env")

	fluentbitPort, _ := strconv.Atoi(GetEnv("FLUENTBIT_PORT", "24223"))
	appName := GetEnv("APP_NAME", "")
	if appName == "" {
		appName = GetEnv("ELASTIC_APM_SERVICE_NAME", "")
	}

	logrus.Infoln("GRPC_MAX_CALL_RECEIVE_MESSAGE: ", GetEnv("GRPC_MAX_CALL_RECEIVE_MESSAGE", "400000000"))
	maxCallRecvMsgSize, err := strconv.Atoi(GetEnv("GRPC_MAX_CALL_RECEIVE_MESSAGE", "400000000"))
	if err != nil {
		maxCallRecvMsgSize = 4
	}
	logrus.Infoln("GRPC_MAX_CALL_SEND_MESSAGE: ", GetEnv("GRPC_MAX_CALL_SEND_MESSAGE", "400000000"))
	maxCallSendMsgSize, err := strconv.Atoi(GetEnv("GRPC_MAX_CALL_SEND_MESSAGE", "400000000"))
	if err != nil {
		maxCallSendMsgSize = 4
	}

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
		ExposedHeaders:     []string{"Signature"},
		CorsAllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "PUT"},
		CorsAllowedOrigins: []string{"*"},

		JWTSecret:   GetEnv("JWT_SECRET", "secret"),
		JWTDuration: GetEnv("JWT_DURATION", "24h"),

		LoggerTag:    GetEnv("LOGGER_TAG", "identity.dev"),
		LoggerOutput: GetEnv("LOGGER_OUTPUT", "stdout"),
		LoggerLevel:  GetEnv("LOGGER_LEVEL", "debug"),
		AppName:      appName,

		GrpcMaxCallSendMessage: maxCallSendMsgSize,
		GrpcMaxCallRecvMessage: maxCallRecvMsgSize,

		Env:           GetEnv("ENV", "DEV"),
		ProductName:   GetEnv("PRODUCT_NAME", "BRICaMS"),
		FluentbitHost: GetEnv("FLUENTBIT_HOST", "0.0.0.0"),
		FluentbitPort: fluentbitPort,
		TimeLocation:  timeLocation,

		DbHost:     GetEnv("DB_HOST", "localhost"),
		DbUser:     GetEnv("DB_USER", "postgres"),
		DbPassword: GetEnv("DB_PASSWORD", "postgres"),
		DbName:     GetEnv("DB_NAME", "identity"),
		DbPort:     GetEnv("DB_PORT", "5432"),
		DbSslmode:  GetEnv("DB_SSLMODE", "disable"),
		DbTimezone: GetEnv("DB_TIMEZONE", "Asia/Jakarta"),
		DbMaxRetry: GetEnv("DB_MAX_RETRY", "3"),
		DbTimeout:  GetEnv("DB_TIMEOUT", "300"),

		ProfileServiceURL: GetEnv("PROFILE_SERVICE_URL", "http://localhost:8080"),
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
