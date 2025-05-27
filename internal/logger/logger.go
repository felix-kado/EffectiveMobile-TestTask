package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

func NewLogger(level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func parseLevel(l string) slog.Level {
	switch l {
	case "debug":
		return slog.LevelDebug
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
