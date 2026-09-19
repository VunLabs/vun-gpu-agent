package logger

import (
	"log/slog"
	"os"
)

// New returns the process logger configured for text output.
func New() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
