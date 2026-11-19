// Package logger membuat structured logger (log/slog bawaan Go).
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New mengembalikan logger JSON untuk production (mudah dibaca oleh tool log
// aggregator seperti Loki/ELK) dan logger teks yang lebih ramah untuk lokal.
func New(env, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
