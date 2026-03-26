package logger

import "github.com/fluent/fluent-logger-golang/fluent"

type FluentBitHook struct {
	logger *fluent.Fluent
}

func NewFluentBitHook(serviceName string, hostname string, port int) (*FluentBitHook, error) {
	fluentConfig := fluent.Config{
		FluentPort: port,
		FluentHost: hostname,
		TagPrefix:  serviceName,
	}
	fluentLogger, err := fluent.New(fluentConfig)
	if err != nil {
		return nil, err
	}
	return &FluentBitHook{logger: fluentLogger}, nil
}

func (hook *FluentBitHook) Write(p []byte) (n int, err error) {
	logEntry := map[string]string{"message": string(p)}
	err = hook.logger.Post("zap", logEntry)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (hook *FluentBitHook) Sync() error {
	return hook.logger.Close()
}
