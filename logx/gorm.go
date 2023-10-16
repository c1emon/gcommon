package logx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	gl "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var gormLog *gormLogrus

var onceGorm = sync.Once{}

type gormLogrus struct {
	*logrus.Logger
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func (l *gormLogrus) LogMode(level gl.LogLevel) gl.Interface {
	lv := Gorm2LogrusLogLevel(level)
	l.Logger.SetLevel(lv)
	return l
}

func (l *gormLogrus) Info(ctx context.Context, format string, values ...interface{}) {
	l.Logger.WithContext(ctx).Infof(format, values...)
}

func (l *gormLogrus) Warn(ctx context.Context, format string, values ...interface{}) {
	l.Logger.WithContext(ctx).Warnf(format, values...)
}

func (l *gormLogrus) Error(ctx context.Context, format string, values ...interface{}) {
	l.Logger.WithContext(ctx).Errorf(format, values...)
}

func (l *gormLogrus) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.Logger.GetLevel() <= logrus.FatalLevel {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	caller := utils.FileWithLineNum()

	fields := logrus.Fields{}
	fields["sql"] = sql
	fields["elapsed"] = fmt.Sprintf("%d ms", elapsed.Milliseconds())
	fields["caller"] = caller

	if err != nil && (!errors.Is(err, gl.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError) {
		l.Logger.WithContext(ctx).WithFields(fields).Errorf("%s", err)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		l.Logger.WithContext(ctx).WithFields(fields).Warnf("slow sql (>%dms)", l.SlowThreshold.Milliseconds())
		return
	}

	if l.Logger.GetLevel() >= logrus.DebugLevel {
		l.Logger.WithContext(ctx).WithFields(fields).Debugf("exec sql (affect %d rows)", rows)
	}
}

func GetGormLogrusLogger() *gormLogrus {
	onceGorm.Do(func() {
		l := GetLogger()
		gormLog = &gormLogrus{
			Logger:                    l,
			SlowThreshold:             200 * time.Millisecond,
			IgnoreRecordNotFoundError: false,
		}
	})

	return gormLog
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
