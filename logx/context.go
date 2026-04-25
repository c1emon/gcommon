package logx

import (
	"context"
	"log/slog"
)

type contextAttrsKey struct{}

// WithAttrs appends slog attrs into context for downstream logging.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	existing := attrsFromContext(ctx)
	merged := make([]slog.Attr, len(existing), len(existing)+len(attrs))
	copy(merged, existing)
	merged = append(merged, attrs...)
	return context.WithValue(ctx, contextAttrsKey{}, merged)
}

func attrsFromContext(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	v, _ := ctx.Value(contextAttrsKey{}).([]slog.Attr)
	return v
}

// FromContext returns the default logger.
// Context attrs are injected by the package handler during Handle.
func FromContext(_ context.Context) *slog.Logger {
	return Default()
}
