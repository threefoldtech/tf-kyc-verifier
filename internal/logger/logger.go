/*
Package logger contains a Logger Wrapper to enable support for multiple logging libraries.
This is a layer between the application code and the underlying logging library.
It provides a simplified API that abstracts away the complexity of different logging libraries, making it easier to switch between them or add new ones.
*/
package logger

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
)

type LoggerW struct {
	logger Logger
}

type Fields map[string]interface{}

var log *LoggerW

func Init(config config.Log) error {
	zapLogger, err := NewZapLogger(config.Debug, context.Background())
	if err != nil {
		return fmt.Errorf("initializing zap logger: %w", err)
	}

	log = &LoggerW{logger: zapLogger}
	return nil
}

func GetLogger() (*LoggerW, error) {
	if log == nil {
		return nil, fmt.Errorf("logger not initialized")
	}
	return log, nil
}

func (lw *LoggerW) Debug(msg string, fields Fields) {
	lw.logger.Debug(msg, fields)
}

func (lw *LoggerW) Info(msg string, fields Fields) {
	lw.logger.Info(msg, fields)
}

func (lw *LoggerW) Warn(msg string, fields Fields) {
	lw.logger.Warn(msg, fields)
}

func (lw *LoggerW) Error(msg string, fields Fields) {
	lw.logger.Error(msg, fields)
}

func (lw *LoggerW) Fatal(msg string, fields Fields) {
	lw.logger.Fatal(msg, fields)
}
