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
	cstZone             = time.FixedZone("GMT", 8*3600)
	debugColorFormatter = color.New(color.FgHiYellow).SprintFunc()
	infoColorFormatter  = color.New(color.FgGreen).SprintFunc()
	warnColorFormatter  = color.New(color.FgYellow).SprintFunc()
	errorColorFormatter = color.New(color.FgRed).SprintFunc()
)

var _ Logger = &LogrusLogger{}

func NewLogrusLogger(lv Level) *LogrusLoggerFactory {
	l := &LogrusLoggerFactory{
		logrus: logrus.New(),
	}

	l.logrus.SetFormatter(l)
	l.logrus.SetLevel(lv.ToLogrusLevel())
	l.logrus.Info(fmt.Sprintf("log level: %s", l.logrus.GetLevel().String()))

	return l
}

type LogrusLoggerFactory struct {
	logrus *logrus.Logger
}

func (l *LogrusLoggerFactory) Get(name string) Logger {
	return &LogrusLogger{
		Entry: logrus.NewEntry(l.logrus),
		name:  name,
	}
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

	timestamp := time.Now().In(cstZone).Format("2006-01-02 15:04:05.999")
	msg := fmt.Sprintf("%s [%s] -- %s\n",
		timestamp,
		colorFormatter(strings.ToUpper(entry.Level.String())),
		entry.Message)
	if entry.Data != nil && len(entry.Data) > 0 {
		msg = fmt.Sprintf("%s\n%s\n", msg, util.PrettyMarshal(entry.Data))
	}

	return []byte(msg), nil
}

type LogrusLogger struct {
	*logrus.Entry
	name string
}

// Debug implements Logger.
func (l *LogrusLogger) Debug(format string, values ...any) {
	l.Debugf(format, values...)
}

// DebugWith implements Logger.
func (l *LogrusLogger) DebugWith(opts []logOption, format string, values ...any) {
	lo := readOptions(opts)
	l.WithFields(logrus.Fields{"xx": lo.GetValues()}).Debugf(format, values...)
}

// Error implements Logger.
func (l *LogrusLogger) Error(format string, values ...any) {
	panic("unimplemented")
}

// ErrorWith implements Logger.
func (l *LogrusLogger) ErrorWith(opts []logOption, format string, values ...any) {
	panic("unimplemented")
}

// Fatal implements Logger.
func (l *LogrusLogger) Fatal(format string, values ...any) {
	panic("unimplemented")
}

// FatalWith implements Logger.
func (l *LogrusLogger) FatalWith(opts []logOption, format string, values ...any) {
	panic("unimplemented")
}

// Info implements Logger.
func (l *LogrusLogger) Info(format string, values ...any) {
	panic("unimplemented")
}

// InfoWith implements Logger.
func (l *LogrusLogger) InfoWith(opts []logOption, format string, values ...any) {
	panic("unimplemented")
}

// Panic implements Logger.
func (l *LogrusLogger) Panic(format string, values ...any) {
	panic("unimplemented")
}

// PanicWith implements Logger.
func (l *LogrusLogger) PanicWith(opts []logOption, format string, values ...any) {
	panic("unimplemented")
}

// Trace implements Logger.
func (l *LogrusLogger) Trace(format string, values ...any) {
	panic("unimplemented")
}

// TraceWith implements Logger.
func (l *LogrusLogger) TraceWith(opts []logOption, format string, values ...any) {
	panic("unimplemented")
}

// Warn implements Logger.
func (l *LogrusLogger) Warn(format string, values ...any) {
	panic("unimplemented")
}

// WarnWith implements Logger.
func (l *LogrusLogger) WarnWith(opts []logOption, format string, values ...any) {
	panic("unimplemented")
}
