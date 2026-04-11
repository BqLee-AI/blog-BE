package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log *zap.Logger = zap.NewNop()
	mu  sync.Mutex
)

func InitLogger(env string) {
	mu.Lock()
	defer mu.Unlock()

	if Log != nil {
		_ = Log.Sync()
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

	logger, err := config.Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		Log = zap.NewNop()
		return
	}

	Log = logger
}

func Sync() {
	mu.Lock()
	defer mu.Unlock()

	if Log != nil {
		_ = Log.Sync()
	}
}
