package logger

import (
	"log/slog"
	"os"
)

const (
	envDebug = "debug"
	envProd  = "prod"
)

func SetupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envDebug:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo.Level()}))
	}

	return logger
}