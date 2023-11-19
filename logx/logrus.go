package logx

import (
	"fmt"
	"strings"
	"time"

	"github.com/c1emon/gcommon/util"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var (
	debugColorFormatter = color.New(color.FgHiYellow).SprintFunc()
	infoColorFormatter  = color.New(color.FgGreen).SprintFunc()
	warnColorFormatter  = color.New(color.FgYellow).SprintFunc()
	errorColorFormatter = color.New(color.FgRed).SprintFunc()
)

var _ Logger = &LogrusLogger{}
var _ LoggerFactory = &LogrusLoggerFactory{}

func NewLogrusLoggerFactory(lv Level) *LogrusLoggerFactory {
	l := &LogrusLoggerFactory{
		logrus:     logrus.New(),
		timeFormat: "2006-01-02 15:04:05.999",
		timeZone:   time.FixedZone("GMT", 8*3600),
		lv:         lv,
		loggers:    make(map[string]Logger),
	}

	l.logrus.SetFormatter(l)
	l.logrus.SetLevel(lv.ToLogrusLevel())
	l.logrus.Info(fmt.Sprintf("log level: %s", l.logrus.GetLevel().String()))

	return l
}

type LogrusLoggerFactory struct {
	logrus     *logrus.Logger
	timeFormat string
	timeZone   *time.Location
	lv         Level
	loggers    map[string]Logger
}

func (l *LogrusLoggerFactory) Get(name string) Logger {
	if logger, ok := l.loggers[name]; ok {
		return logger
	}
	logger := &LogrusLogger{
		Entry: logrus.NewEntry(l.logrus),
		name:  name,
		lv:    l.lv,
	}
	l.loggers[name] = logger
	return logger
}

func (l *LogrusLoggerFactory) GetLevel() Level {
	return l.lv
}

// Format log format
func (s *LogrusLoggerFactory) Format(entry *logrus.Entry) ([]byte, error) {

	var colorFormatter func(a ...interface{}) string
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		colorFormatter = debugColorFormatter
	case logrus.InfoLevel:
		colorFormatter = infoColorFormatter
	case logrus.WarnLevel:
		colorFormatter = warnColorFormatter
	default:
		colorFormatter = errorColorFormatter
	}

	timestamp := time.Now().In(s.timeZone).Format(s.timeFormat)
	msg := fmt.Sprintf("%s %-7s -- %s\n",
		timestamp,
		colorFormatter(strings.ToUpper(fmt.Sprintf("[%s]", entry.Level.String()))),
		entry.Message)
	if entry.Data != nil && len(entry.Data) > 0 {
		msg = fmt.Sprintf("%s\n%s\n", msg, util.PrettyMarshal(entry.Data))
	}

	return []byte(msg), nil
}

type LogrusLogger struct {
	*logrus.Entry
	name string
	lv   Level
}

func (l *LogrusLogger) GetLevel() Level {
	return l.lv
}

func fromOptions(l *logrus.Entry, opts []logOption) *logrus.Entry {
	var entry *logrus.Entry

	lo := readOptions(opts)
	if vals, ok := lo.GetValues(); ok {
		entry = l.WithFields(vals)
	} else {
		entry = l
	}
	if ctx, ok := lo.GetCtx(); ok {
		entry = entry.WithContext(ctx)
	}

	return entry
}

// Debug implements Logger.
func (l *LogrusLogger) Debug(format string, values ...any) {
	l.Debugf(format, values...)
}

// DebugWith implements Logger.
func (l *LogrusLogger) DebugWith(opts []logOption, format string, values ...any) {
	fromOptions(l.Entry, opts).Debugf(format, values...)
}

// Error implements Logger.
func (l *LogrusLogger) Error(format string, values ...any) {
	l.Errorf(format, values...)
}

// ErrorWith implements Logger.
func (l *LogrusLogger) ErrorWith(opts []logOption, format string, values ...any) {
	fromOptions(l.Entry, opts).Errorf(format, values...)
}

// Fatal implements Logger.
func (l *LogrusLogger) Fatal(format string, values ...any) {
	l.Fatalf(format, values...)
}

// FatalWith implements Logger.
func (l *LogrusLogger) FatalWith(opts []logOption, format string, values ...any) {
	fromOptions(l.Entry, opts).Fatalf(format, values...)
}

// Info implements Logger.
func (l *LogrusLogger) Info(format string, values ...any) {
	l.Infof(format, values...)
}

// InfoWith implements Logger.
func (l *LogrusLogger) InfoWith(opts []logOption, format string, values ...any) {
	fromOptions(l.Entry, opts).Infof(format, values...)
}

// Panic implements Logger.
func (l *LogrusLogger) Panic(format string, values ...any) {
	l.Panicf(format, values...)
}

// PanicWith implements Logger.
func (l *LogrusLogger) PanicWith(opts []logOption, format string, values ...any) {
	fromOptions(l.Entry, opts).Panicf(format, values...)
}

// Trace implements Logger.
func (l *LogrusLogger) Trace(format string, values ...any) {
	l.Tracef(format, values...)
}

// TraceWith implements Logger.
func (l *LogrusLogger) TraceWith(opts []logOption, format string, values ...any) {
	fromOptions(l.Entry, opts).Tracef(format, values...)
}

// Warn implements Logger.
func (l *LogrusLogger) Warn(format string, values ...any) {
	l.Warnf(format, values...)
}

// WarnWith implements Logger.
func (l *LogrusLogger) WarnWith(opts []logOption, format string, values ...any) {
	fromOptions(l.Entry, opts).Warnf(format, values...)
}
