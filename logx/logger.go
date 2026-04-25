package logx

import (
	"io"
	"log/slog"
	"os"
)

const (
	FormatJSON = "json"
	FormatText = "text"
)

type Config struct {
	Level       slog.Level
	Format      string
	Output      io.Writer
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	FixedAttrs  []slog.Attr
}

// NewHandler creates a stdlib slog handler from Config.
func NewHandler(cfg Config) slog.Handler {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	if cfg.Format == "" {
		cfg.Format = FormatText
	}

	opts := &slog.HandlerOptions{
		Level:       cfg.Level,
		AddSource:   cfg.AddSource,
		ReplaceAttr: cfg.ReplaceAttr,
	}

	var baseHandler slog.Handler
	switch cfg.Format {
	case FormatJSON:
		baseHandler = slog.NewJSONHandler(cfg.Output, opts)
	default:
		baseHandler = slog.NewTextHandler(cfg.Output, opts)
	}

	fixedAttrs := make([]slog.Attr, 0, len(cfg.FixedAttrs))
	fixedAttrs = append(fixedAttrs, cfg.FixedAttrs...)
	return &contextHandler{
		Handler: baseHandler,
		fixed:   fixedAttrs,
	}
}

// NewLogger creates a logger from Config.
func NewLogger(cfg Config) *slog.Logger {
	return slog.New(NewHandler(cfg))
}
