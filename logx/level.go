package logx

import (
	"fmt"
	"log/slog"
	"strings"
)

// ParseLevel converts a level string into slog.Level.
func ParseLevel(lvStr string) (slog.Level, error) {
	switch strings.ToUpper(strings.TrimSpace(lvStr)) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN", "WARNING":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid slog level: %s", lvStr)
	}
}
