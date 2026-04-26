package ginx

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const maxGinWriterBuffer = 64 << 10

// slogLineWriter buffers writes until a newline, then emits one slog record per line.
type slogLineWriter struct {
	mu    sync.Mutex
	log   *slog.Logger
	level slog.Level
	buf   []byte
}

func (w *slogLineWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			if len(w.buf) > maxGinWriterBuffer {
				w.emitLine(w.buf)
				w.buf = w.buf[:0]
			}
			break
		}
		line := w.buf[:i]
		w.buf = w.buf[i+1:]
		w.emitLine(line)
	}
	return len(p), nil
}

func (w *slogLineWriter) emitLine(line []byte) {
	s := strings.TrimSpace(strings.ReplaceAll(string(line), "\r", ""))
	if s == "" {
		return
	}
	w.log.Log(context.Background(), w.level, "gin", slog.String("message", s))
}

// SetGinSlogWriters assigns [gin.DefaultWriter] and [gin.DefaultErrorWriter] to adapters
// that forward each line to logger as structured logs (info vs error level).
// logger must be non-nil.
//
// This mutates package-level variables in [github.com/gin-gonic/gin]; call once at process
// startup, or rely on [NewDefaultEngine] which invokes this for you.
func SetGinSlogWriters(logger *slog.Logger) {
	if logger == nil {
		panic("ginx: SetGinSlogWriters called with nil *slog.Logger")
	}
	gin.DefaultWriter = &slogLineWriter{log: logger, level: slog.LevelInfo}
	gin.DefaultErrorWriter = &slogLineWriter{log: logger, level: slog.LevelError}
}
