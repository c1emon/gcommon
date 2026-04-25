package logx

import (
	"context"
	"log/slog"
)

type contextHandler struct {
	slog.Handler
	fixed []slog.Attr
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if len(h.fixed) > 0 {
		r.AddAttrs(h.fixed...)
	}
	if attrs := attrsFromContext(ctx); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{
		Handler: h.Handler.WithAttrs(attrs),
		fixed:   h.fixed,
	}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{
		Handler: h.Handler.WithGroup(name),
		fixed:   h.fixed,
	}
}
