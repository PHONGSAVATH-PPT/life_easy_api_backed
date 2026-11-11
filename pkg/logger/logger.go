package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New() (*zap.Logger, error) {
	if err := os.MkdirAll("logs", 0750); err != nil {
		return nil, err
	}

	file := &lumberjack.Logger{
		Filename:   "logs/api.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
	}

	encoder := zap.NewProductionEncoderConfig()
	encoder.TimeKey = "timestamp"
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder

	output := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(file),
	)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		output,
		zap.InfoLevel,
	)

	return zap.New(core), nil
}
