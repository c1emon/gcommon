package logx

import (
	"context"

	"github.com/c1emon/gcommon/util"
)

type logOptions struct {
	ctx    context.Context
	values map[string]any
}

type logOption util.Option[logOptions]

func (o logOptions) GetCtx() (context.Context, bool) {
	if o.ctx == nil {
		return nil, false
	}
	return o.ctx, true
}

func (o logOptions) GetValues() (map[string]any, bool) {
	if o.values == nil {
		return nil, false
	}
	return o.values, true
}

func readOptions(opts []logOption) *logOptions {
	lo := &logOptions{
		ctx:    nil,
		values: nil,
	}
	for _, o := range opts {
		o.Apply(lo)
	}
	return lo
}

func WithContext(ctx context.Context) logOption {
	return util.WrapFuncOption[logOptions](
		func(lo *logOptions) {
			lo.ctx = ctx
		})
}

func WithValues(values map[string]any) logOption {
	return util.WrapFuncOption[logOptions](
		func(lo *logOptions) {
			lo.values = values
		})
}

type Logger interface {
	GetLevel() Level

	Trace(format string, values ...any)
	Debug(format string, values ...any)
	Info(format string, values ...any)
	Warn(format string, values ...any)
	Error(format string, values ...any)
	Fatal(format string, values ...any)
	Panic(format string, values ...any)

	TraceWith(opts []logOption, format string, values ...any)
	DebugWith(opts []logOption, format string, values ...any)
	InfoWith(opts []logOption, format string, values ...any)
	WarnWith(opts []logOption, format string, values ...any)
	ErrorWith(opts []logOption, format string, values ...any)
	FatalWith(opts []logOption, format string, values ...any)
	PanicWith(opts []logOption, format string, values ...any)
}

type LoggerFactory interface {
	Get(name string) Logger
	GetLevel() Level
}
