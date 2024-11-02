package logger

import (
	"context"

	"example.com/tfgrid-kyc-service/internal/configs"
)

type LoggerW struct {
	logger Logger
}

type Fields map[string]interface{}

var log *LoggerW

func Init(config configs.Log) {
	zapLogger, err := NewZapLogger(config.Debug, context.Background())
	if err != nil {
		panic(err)
	}

	log = &LoggerW{logger: zapLogger}
}

func GetLogger() *LoggerW {
	if log == nil {
		panic("logger not initialized")
	}
	return log
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
