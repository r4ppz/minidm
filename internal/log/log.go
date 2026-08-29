package log

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	level := slog.LevelInfo
	if os.Getenv("MINIDM_DEBUG") == "1" {
		level = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

func Debug(msg string, args ...any) { logger.Debug(msg, args...) }
func Info(msg string, args ...any)  { logger.Info(msg, args...) }
func Error(msg string, args ...any) { logger.Error(msg, args...) }
