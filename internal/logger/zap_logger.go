package logger

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.Logger
	ctx    context.Context
}

func NewZapLogger(debug bool, ctx context.Context) (*ZapLogger, error) {
	zapConfig := zap.NewProductionConfig()
	if debug {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.DisableCaller = true
	zapLog, err := zapConfig.Build()
	if err != nil {
		return nil, errors.Join(errors.New("building zap logger from the config"), err)
	}

	return &ZapLogger{logger: zapLog, ctx: ctx}, nil
}

func (l *ZapLogger) Debug(msg string, fields Fields) {
	l.addContextCommonFields(fields)

	l.logger.Debug(msg, zap.Any("args", fields))
}

func (l *ZapLogger) Info(msg string, fields Fields) {
	l.addContextCommonFields(fields)

	l.logger.Info(msg, zap.Any("args", fields))
}

func (l *ZapLogger) Warn(msg string, fields Fields) {
	l.addContextCommonFields(fields)

	l.logger.Warn(msg, zap.Any("args", fields))
}

func (l *ZapLogger) Error(msg string, fields Fields) {
	l.addContextCommonFields(fields)

	l.logger.Error(msg, zap.Any("args", fields))
}

func (l *ZapLogger) Fatal(msg string, fields Fields) {
	l.addContextCommonFields(fields)

	l.logger.Fatal(msg, zap.Any("args", fields))
}

func (l *ZapLogger) addContextCommonFields(fields Fields) {
	if l.ctx != nil && l.ctx.Value("commonFields") != nil && fields != nil {
		for k, v := range l.ctx.Value("commonFields").(map[string]interface{}) {
			if _, ok := fields[k]; !ok {
				fields[k] = v
			}
		}
	}
}
