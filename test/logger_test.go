package test

import (
	"testing"

	"github.com/c1emon/gcommon/logx"
)

func TestLogger(t *testing.T) {
	var factory logx.LoggerFactory
	factory = logx.NewLogrusLoggerFactory(logx.LevelDebug)
	logger := factory.Get("test")

	logger.Debug("debug")
}
