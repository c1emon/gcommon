package page

import "github.com/c1emon/gcommon/util"

type options struct {
	offset         int
	nonZeroStartup bool
	ignoreErr      bool
}

func WithOffset(offset int) util.Option[options] {
	return util.WrapFuncOption(func(o *options) {
		o.offset = offset
	})
}

func WithNonZeroStarting(b bool) util.Option[options] {
	return util.WrapFuncOption(func(o *options) {
		o.nonZeroStartup = b
	})
}

func WithIgnoreErr(b bool) util.Option[options] {
	return util.WrapFuncOption(func(o *options) {
		o.ignoreErr = b
	})
}
