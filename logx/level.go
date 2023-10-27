package logx

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// Level is a logger level.
type Level int8

// LevelKey is logger level key.
const LevelKey = "level"

const (
	// LevelTrace is logger trace level.
	LevelTrace Level = iota - 1
	// LevelDebug is logger debug level.
	LevelDebug
	// LevelInfo is logger info level.
	LevelInfo
	// LevelWarn is logger warn level.
	LevelWarn
	// LevelError is logger error level.
	LevelError
	// LevelFatal is logger fatal level
	LevelFatal
)

func (l Level) Key() string {
	return LevelKey
}

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

func (l Level) ToLogrusLevel() logrus.Level {
	lv, _ := logrus.ParseLevel(l.String())
	return lv
}

// ParseLevel parses a level string into a logger Level value.
func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	}
	return LevelInfo
}
