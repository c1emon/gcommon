package test

import (
	"log/slog"
	"testing"

	"github.com/c1emon/gcommon/logx"
)

func Test_logx(t *testing.T) {
	logger := slog.New(logx.NewLogrusSlogHandler())
	logger.Info("hello", "name", "clemon")
}
