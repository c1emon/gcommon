package logx

import (
	"context"

	"github.com/c1emon/gcommon/util"
)

type logOptions struct {
	ctx    context.Context
	values any
}

func readOptions(opts []util.Option[logOptions]) *logOptions {
	lo := &logOptions{
		ctx:    context.TODO(),
		values: nil,
	}
	if opts != nil {
		for _, o := range opts {
			o.Apply(lo)
		}
	}
	return lo
}

func WithContext(ctx context.Context) util.Option[logOptions] {
	return util.WrapFuncOption[logOptions](
		func(lo *logOptions) {
			lo.ctx = ctx
		})
}

func WithValues(values any) util.Option[logOptions] {
	return util.WrapFuncOption[logOptions](
		func(lo *logOptions) {
			lo.values = values
		})
}

type Logger interface {
	Trace(format string, values ...any)
	Debug(format string, values ...any)
	Info(format string, values ...any)
	Warn(format string, values ...any)
	Error(format string, values ...any)
	Fatal(format string, values ...any)
	Panic(format string, values ...any)

	TraceWith(opts []util.Option[logOptions], format string, values ...any)
	DebugWith(opts []util.Option[logOptions], format string, values ...any)
	InfoWith(opts []util.Option[logOptions], format string, values ...any)
	WarnWith(opts []util.Option[logOptions], format string, values ...any)
	ErrorWith(opts []util.Option[logOptions], format string, values ...any)
	FatalWith(opts []util.Option[logOptions], format string, values ...any)
	PanicWith(opts []util.Option[logOptions], format string, values ...any)
}

type LoggerFactory interface {
	Get(service string) Logger
}
