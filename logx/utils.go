package logx

import (
	"fmt"
	"log/slog"
	"strings"
)

// ParseLogLevel converts a level string into slog.Level.
func ParseLogLevel(lvStr string) (slog.Level, error) {
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

func ParseLogFormat(format string) string {
	switch strings.ToLower(format) {
	case "text":
		return FormatText
	case "json":
		return FormatJSON
	default:
		return FormatJSON
	}
}
