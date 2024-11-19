package logx

import (
	"log/slog"
	"strings"
)

func ParseLevel(lvStr string) slog.Level {
	switch strings.ToUpper(lvStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
