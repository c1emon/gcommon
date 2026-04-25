package logx

import (
	"log/slog"
	"sync"
)

var (
	defaultMu     sync.RWMutex
	defaultLogger *slog.Logger
)

// Init builds and stores the package default logger, and also sets slog default logger.
func Init(cfg Config) {
	logger := NewLogger(cfg)

	defaultMu.Lock()
	defaultLogger = logger
	defaultMu.Unlock()

	slog.SetDefault(logger)
}

// Default returns logger initialized by Init, or slog.Default when not initialized.
func Default() *slog.Logger {
	defaultMu.RLock()
	logger := defaultLogger
	defaultMu.RUnlock()
	if logger != nil {
		return logger
	}
	return slog.Default()
}
