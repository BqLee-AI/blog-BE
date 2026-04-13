package logger

import (
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logRef atomic.Pointer[zap.Logger]

var buildLogger = func(config zap.Config) (*zap.Logger, error) {
	return config.Build(zap.AddStacktrace(zap.ErrorLevel))
}

func init() {
	logRef.Store(zap.NewNop())
}

func L() *zap.Logger {
	if current := logRef.Load(); current != nil {
		return current
	}

	return zap.NewNop()
}

func Set(l *zap.Logger) {
	if l == nil {
		l = zap.NewNop()
	}

	logRef.Store(l)
}

func InitLogger(env string) {
	if current := L(); current != nil {
		_ = current.Sync()
	}

	config := zap.NewDevelopmentConfig()
	if env == "production" || env == "release" || env == "prod" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	instance, err := buildLogger(config)
	if err != nil {
		return
	}

	Set(instance)
}

func Sync() {
	if current := L(); current != nil {
		_ = current.Sync()
	}
}
