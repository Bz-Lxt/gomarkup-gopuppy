package logger

import (
	"log/slog"
	"os"
	"strings"
)

func New(level, env string) *slog.Logger {
	lvl := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		if strings.EqualFold(env, "production") {
			lvl = slog.LevelInfo // production auto-suppresses debug
		} else {
			lvl = slog.LevelDebug
		}
	case "info":
		lvl = slog.LevelInfo
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}
