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
	ListenAddress string

	CorsAllowedHeaders []string
	CorsAllowedMethods []string
	CorsAllowedOrigins []string
	ExposedHeaders     []string

	JWTSecret string

	LoggerTag    string
	LoggerOutput string
	LoggerLevel  string
	AppName      string

	GrpcMaxCallSendMessage int
	GrpcMaxCallRecvMessage int

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
	DbMaxRetry string
	DbTimeout  string
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

		JWTSecret: GetEnv("JWT_SECRET", "secret"),

		LoggerTag:    GetEnv("LOGGER_TAG", "payment.dev"),
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
		DbName:     GetEnv("DB_NAME", "payment"),
		DbPort:     GetEnv("DB_PORT", "5435"),
		DbSslmode:  GetEnv("DB_SSLMODE", "disable"),
		DbTimezone: GetEnv("DB_TIMEZONE", "Asia/Jakarta"),
		DbMaxRetry: GetEnv("DB_MAX_RETRY", "3"),
		DbTimeout:  GetEnv("DB_TIMEOUT", "300"),
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
