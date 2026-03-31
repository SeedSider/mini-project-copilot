package logger

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Env           string
	ProductName   string
	ServiceName   string
	LogLevel      string
	LogOutput     string
	FluentbitHost string
	FluentbitPort int
}

type Logger struct {
	zapLog        *zap.Logger
	fluentBitHook *FluentBitHook
	loggerConfig  *LoggerConfig
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

	var core zapcore.Core
	var fluentBitHook *FluentBitHook
	var err error

	if strings.EqualFold(loggerConfig.LogOutput, "elastic") {
		fluentBitHook, err = NewFluentBitHook(loggerConfig.ServiceName, loggerConfig.FluentbitHost, loggerConfig.FluentbitPort)
		if err != nil {
			panic(err)
		}

		if logLevel.Level() == zap.DebugLevel {
			core = zapcore.NewTee(
				zapcore.NewCore(zapEncoder, fluentBitHook, logLevel),
				zapcore.NewCore(zapEncoder, os.Stdout, logLevel),
			)
		} else {
			core = zapcore.NewTee(
				zapcore.NewCore(zapEncoder, fluentBitHook, logLevel),
			)
		}
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(zapEncoder, os.Stdout, logLevel),
		)
	}

	zapLogger := zap.New(core)

	return &Logger{
		zapLogger,
		fluentBitHook,
		loggerConfig,
	}
}

func (l *Logger) Sync() error {
	if l.fluentBitHook != nil {
		err := l.fluentBitHook.Sync()
		if err != nil {
			return err
		}
	}
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
