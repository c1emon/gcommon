package logx

import "log/slog"

const (
	// Common slog attr keys.
	KeyTraceID  = "trace_id"
	KeySpanID   = "span_id"
	KeyUserID   = "user_id"
	KeyClientIP = "client_ip"
	KeyMethod   = "method"
	KeyPath     = "path"
	KeyStatus   = "status"
	KeyDuration = "duration_ms"
	KeyError    = "error"
	KeyStack    = "stack"
)

func TraceID(id string) slog.Attr  { return slog.String(KeyTraceID, id) }
func SpanID(id string) slog.Attr   { return slog.String(KeySpanID, id) }
func UserID(id string) slog.Attr   { return slog.String(KeyUserID, id) }
func ClientIP(ip string) slog.Attr { return slog.String(KeyClientIP, ip) }
func Method(v string) slog.Attr    { return slog.String(KeyMethod, v) }
func Path(v string) slog.Attr      { return slog.String(KeyPath, v) }
func Status(v int) slog.Attr       { return slog.Int(KeyStatus, v) }
func DurationMs(v int64) slog.Attr { return slog.Int64(KeyDuration, v) }
func Stack(v string) slog.Attr     { return slog.String(KeyStack, v) }

func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.String(KeyError, err.Error())
}
