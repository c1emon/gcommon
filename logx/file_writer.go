package logx

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func writer(logFile string) io.WriteCloser {
	return &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
}

func writer1(logFile string) io.WriteCloser {
	log, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {

	}
	return log
}

func tog(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}

// test...
func test() {

	tog(writer1("hello"))
}
