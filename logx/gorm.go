package logx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	gl "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var _ gl.Interface = &gormLogger{}

func NewGormLogger(loggerFactory LoggerFactory) *gormLogger {

	return &gormLogger{
		Logger:                    loggerFactory.Get("gorm_logger"),
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
	}
}

type gormLogger struct {
	Logger
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func (l *gormLogger) LogMode(level gl.LogLevel) gl.Interface {
	// lv := Gorm2LogrusLogLevel(level)
	// l.Logger.SetLevel(lv)
	l.Logger.Warn("can not change log level logrus logger")
	return l
}

func (l *gormLogger) Info(ctx context.Context, format string, values ...interface{}) {
	opts := make([]logOption, 0)
	opts = append(opts, WithContext(ctx))

	l.Logger.InfoWith(opts, format, values...)
}

func (l *gormLogger) Warn(ctx context.Context, format string, values ...interface{}) {
	opts := make([]logOption, 0)
	opts = append(opts, WithContext(ctx))

	l.Logger.WarnWith(opts, format, values...)
}

func (l *gormLogger) Error(ctx context.Context, format string, values ...interface{}) {
	opts := make([]logOption, 0)
	opts = append(opts, WithContext(ctx))

	l.Logger.ErrorWith(opts, format, values...)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.Logger.GetLevel() <= LevelFatal {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	caller := utils.FileWithLineNum()

	fields := make(map[string]any)
	fields["sql"] = sql
	fields["elapsed"] = fmt.Sprintf("%d ms", elapsed.Milliseconds())
	fields["caller"] = caller

	if err != nil && (!errors.Is(err, gl.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError) {
		opts := make([]logOption, 0)
		opts = append(opts, WithContext(ctx))
		opts = append(opts, WithValues(fields))

		l.Logger.ErrorWith(opts, "%s", err)

		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		opts := make([]logOption, 0)
		opts = append(opts, WithContext(ctx))
		opts = append(opts, WithValues(fields))

		l.Logger.WarnWith(opts, "slow sql (>%dms)", l.SlowThreshold.Milliseconds())
		return
	}

	if l.Logger.GetLevel() >= LevelDebug {

		opts := make([]logOption, 0)
		opts = append(opts, WithContext(ctx))
		opts = append(opts, WithValues(fields))

		l.Logger.DebugWith(opts, "exec sql (affect %d rows)", rows)
	}
}

func Logrus2GormLogLevel(level logrus.Level) gl.LogLevel {
	switch level {
	case logrus.FatalLevel, logrus.PanicLevel:
		return gl.Silent
	case logrus.ErrorLevel:
		return gl.Error
	case logrus.WarnLevel:
		return gl.Warn
	default:
		return gl.Info
	}
}

func Gorm2LogrusLogLevel(level gl.LogLevel) logrus.Level {
	switch level {
	case gl.Silent:
		return logrus.PanicLevel
	case gl.Error:
		return logrus.ErrorLevel
	case gl.Warn:
		return logrus.WarnLevel
	default:
		return logrus.InfoLevel
	}
}
