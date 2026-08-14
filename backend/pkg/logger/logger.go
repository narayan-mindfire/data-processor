package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// Initialize a global structured JSON logger
func InitLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Log = slog.New(handler)
	slog.SetDefault(Log)
}
