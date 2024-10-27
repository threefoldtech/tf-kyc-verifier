package logger

import (
	"example.com/tfgrid-kyc-service/internal/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

var log *Logger

func Init(config configs.Log) {
	debug := config.Debug
	zapConfig := zap.NewProductionConfig()
	if debug {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	var err error

	zapLog, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}
	log = &Logger{zapLog}
}

func GetLogger() *Logger {
	if log == nil {
		panic("logger not initialized")
	}
	return log
}
