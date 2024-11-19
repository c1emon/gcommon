package logx

import (
	"log/slog"
	"os"

	"github.com/c1emon/gcommon/util"
	"github.com/phsym/console-slog"
)

// NewConsoleSlogHandler slog handler for easy reading
func NewConsoleSlogHandler(opts ...util.Option[console.HandlerOptions]) slog.Handler {
	o := &console.HandlerOptions{}
	for _, opt := range opts {
		opt.Apply(o)
	}
	return console.NewHandler(os.Stderr, o)
}

type SlogConsoleOptionHolder struct{}

func (SlogConsoleOptionHolder) WithTimeFormat(timeFormat string) util.Option[console.HandlerOptions] {
	return util.WrapFuncOption(func(t *console.HandlerOptions) {
		t.TimeFormat = timeFormat
	})
}

func (SlogConsoleOptionHolder) WithLevel(lv slog.Level) util.Option[console.HandlerOptions] {
	return util.WrapFuncOption(func(t *console.HandlerOptions) {
		t.Level = lv
	})
}

func (SlogConsoleOptionHolder) WithColor(enable bool) util.Option[console.HandlerOptions] {
	return util.WrapFuncOption(func(t *console.HandlerOptions) {
		t.NoColor = !enable
	})
}
