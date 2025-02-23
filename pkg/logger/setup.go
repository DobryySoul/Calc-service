package logger

import (
	"go.uber.org/zap"
)

func SetupLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	logger.Sugar().Info("logger initialized")
	return logger
}
