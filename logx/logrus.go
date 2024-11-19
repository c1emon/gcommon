package logx

import (
	"log/slog"

	"github.com/c1emon/gcommon/util"
	sloglogrus "github.com/samber/slog-logrus/v2"
	"github.com/sirupsen/logrus"
)

// NewLogrusSlogHandler use logrus as slog handler
func NewLogrusSlogHandler(opts ...util.Option[sloglogrus.Option]) slog.Handler {
	o := sloglogrus.Option{}
	for _, opt := range opts {
		opt.Apply(&o)
	}
	return o.NewLogrusHandler()
}

// Format log format
// func  Format(entry *logrus.Entry) ([]byte, error) {

// 	timestamp := time.Now().In(s.timeZone).Format(s.timeFormat)
// 	msg := fmt.Sprintf("%s %-7s -- %s\n",
// 		timestamp,
// 		colorFormatter(strings.ToUpper(fmt.Sprintf("[%s]", entry.Level.String()))),
// 		entry.Message)
// 	if entry.Data != nil && len(entry.Data) > 0 {
// 		msg = fmt.Sprintf("%s\n%s\n", msg, util.PrettyMarshal(entry.Data))
// 	}

// 	return []byte(msg), nil
// }

type SlogLogrusOptionHolder struct{}

func (SlogLogrusOptionHolder) WithLogger(logger *logrus.Logger) util.Option[sloglogrus.Option] {
	return util.WrapFuncOption(func(t *sloglogrus.Option) {
		t.Logger = logger
	})
}

func (SlogLogrusOptionHolder) WithLevel(lv slog.Level) util.Option[sloglogrus.Option] {
	return util.WrapFuncOption(func(t *sloglogrus.Option) {
		t.Level = lv
	})
}

// func (SlogLogrusOptionHolder) WithLevel(lv slog.Level) util.Option[sloglogrus.Option] {
// 	return util.WrapFuncOption(func(t *sloglogrus.Option) {
// 		t.Level = lv
// 	})
// }
