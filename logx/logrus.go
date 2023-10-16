package logx

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/c1emon/gcommon/util"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var once = sync.Once{}
var loggerInstance *logrus.Logger

func GetLoggerInstance() *logrus.Logger {
	once.Do(func() {
		loggerInstance = logrus.New()
	})
	return loggerInstance
}

func GetLogger() *logrus.Logger {
	// return GetLoggerInstance().WithField("", "").Logger
	return GetLoggerInstance()
}

var cstZone = time.FixedZone("GMT", 8*3600)

var (
	debugColorFormatter = color.New(color.FgHiYellow).SprintFunc()
	infoColorFormatter  = color.New(color.FgGreen).SprintFunc()
	warnColorFormatter  = color.New(color.FgYellow).SprintFunc()
	errorColorFormatter = color.New(color.FgRed).SprintFunc()
)

// CostumeLogFormatter Custom log format definition
type costumeLogFormatter struct{}

// Format log format
func (s *costumeLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {

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

func Init(level string) {
	logger := GetLogger()
	lv, err := logrus.ParseLevel(level)
	if err != nil {
		logger.Fatal(err)
	}
	logger.SetFormatter(new(costumeLogFormatter))
	logger.SetLevel(lv)
	logger.Info(fmt.Sprintf("log level: %s", logger.GetLevel().String()))
}
