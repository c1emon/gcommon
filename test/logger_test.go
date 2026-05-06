package test

import (
	"testing"

	"github.com/c1emon/gcommon/logx/v2"
)

func Test_logx(t *testing.T) {
	logger := logx.NewLogger(logx.Config{
		Format: logx.FormatText,
	})
	logger.Info("hello", "name", "clemon")
}
