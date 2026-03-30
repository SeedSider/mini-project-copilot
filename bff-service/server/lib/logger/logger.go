package logger

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Env         string
	ProductName string
	ServiceName string
	LogLevel    string
	LogOutput   string
}

type Logger struct {
	zapLog       *zap.Logger
	loggerConfig *LoggerConfig
}

var hostname string

func New(loggerConfig *LoggerConfig) *Logger {
	logLevel, levelErr := zap.ParseAtomicLevel(loggerConfig.LogLevel)
	if levelErr != nil {
		panic(levelErr)
	}

	hostname, _ = os.Hostname()

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.MessageKey = "message"

	zapEncoder := zapcore.NewJSONEncoder(encoderCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(zapEncoder, os.Stdout, logLevel),
	)

	zapLogger := zap.New(core)

	return &Logger{
		zapLogger,
		loggerConfig,
	}
}

func (l *Logger) Sync() error {
	return l.zapLog.Sync()
}

func (l *Logger) Info(processId, functionName, message string,
	startTime, endTime, responseTime *time.Time, metadata interface{}) {
	l.zapLog.Info(message,
		zap.String("hostname", hostname),
		zap.String("product_name", l.loggerConfig.ProductName),
		zap.String("service_name", l.loggerConfig.ServiceName),
		zap.String("process_id", processId),
		zap.String("function_name", functionName),
		zap.String("log_type", "application"),
		zap.Timep("start_time", startTime),
		zap.Timep("end_time", endTime),
		zap.Timep("response_time", responseTime),
		zap.Any("metadata", ParseMetadata(metadata)))
}

func (l *Logger) Warn(processId, functionName, message string,
	startTime, endTime, responseTime *time.Time, metadata interface{}) {
	l.zapLog.Warn(message,
		zap.String("hostname", hostname),
		zap.String("product_name", l.loggerConfig.ProductName),
		zap.String("service_name", l.loggerConfig.ServiceName),
		zap.String("process_id", processId),
		zap.String("function_name", functionName),
		zap.String("log_type", "application"),
		zap.Timep("start_time", startTime),
		zap.Timep("end_time", endTime),
		zap.Timep("response_time", responseTime),
		zap.Any("metadata", ParseMetadata(metadata)))
}

func (l *Logger) Error(processId, functionName, message string,
	startTime, endTime, responseTime *time.Time, metadata interface{}) {
	l.zapLog.Error(message,
		zap.String("hostname", hostname),
		zap.String("product_name", l.loggerConfig.ProductName),
		zap.String("service_name", l.loggerConfig.ServiceName),
		zap.String("process_id", processId),
		zap.String("function_name", functionName),
		zap.String("log_type", "application"),
		zap.Timep("start_time", startTime),
		zap.Timep("end_time", endTime),
		zap.Timep("response_time", responseTime),
		zap.Any("metadata", ParseMetadata(metadata)))
}

func (l *Logger) Debug(processId, functionName, message string,
	startTime, endTime, responseTime *time.Time, metadata interface{}) {
	l.zapLog.Debug(message,
		zap.String("hostname", hostname),
		zap.String("product_name", l.loggerConfig.ProductName),
		zap.String("service_name", l.loggerConfig.ServiceName),
		zap.String("process_id", processId),
		zap.String("function_name", functionName),
		zap.String("log_type", "application"),
		zap.Timep("start_time", startTime),
		zap.Timep("end_time", endTime),
		zap.Timep("response_time", responseTime),
		zap.Any("metadata", ParseMetadata(metadata)))
}

func (l *Logger) Fatal(processId, functionName, message string,
	startTime, endTime, responseTime *time.Time, metadata interface{}) {
	l.zapLog.Fatal(message,
		zap.String("hostname", hostname),
		zap.String("product_name", l.loggerConfig.ProductName),
		zap.String("service_name", l.loggerConfig.ServiceName),
		zap.String("process_id", processId),
		zap.String("function_name", functionName),
		zap.String("log_type", "application"),
		zap.Timep("start_time", startTime),
		zap.Timep("end_time", endTime),
		zap.Timep("response_time", responseTime),
		zap.Any("metadata", ParseMetadata(metadata)))
}

func ParseMetadata(metadata interface{}) interface{} {
	if metadata == nil {
		return nil
	}

	switch reflect.TypeOf(metadata).Kind() {
	case reflect.Map, reflect.Struct:
		return metadata
	default:
		return map[string]string{"default_value": fmt.Sprintf("%v", metadata)}
	}
}
